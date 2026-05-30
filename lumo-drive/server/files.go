package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
)

func generateUUID() uuid.UUID {
	// Generate a UUID v4 string
	return uuid.New()
}

func NewS3Client(ctx context.Context) (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("ap-south-1"),
	)
	if err != nil {
		return nil, err
	}

	return s3.NewFromConfig(cfg), nil
}

type ChunkRequest struct {
	ChunkIndex int    `json:"chunk_index"`
	Checksum   string `json:"checksum"`
	ChunkSize  int64  `json:"chunk_size"`
}

type InitiateUploadRequest struct {
	FileName       string  `json:"file_name"`
	Extension      string  `json:"extension"`
	MimeType       string  `json:"mime_type"`
	ParentFolderID *string `json:"parent_folder_id,omitempty"`
	FileSize       int64   `json:"file_size"`
	Checksum       string  `json:"checksum"`

	CompressionAlgo CompressionAlgorithm `json:"compression_algo"`

	Chunks []ChunkRequest `json:"chunks"`
}

func initiateUploadHandler(c *fiber.Ctx) error {

	var req InitiateUploadRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if req.FileName == "" {
		return fiber.NewError(fiber.StatusBadRequest, "file_name is required")
	}

	if req.Extension == "" {
		return fiber.NewError(fiber.StatusBadRequest, "extension is required")
	}

	// Create file metadata record with status "uploading"

	ownerID := getUserID(c)

	chunkSize := int64(5 * 1024 * 1024)

	totalChunks := int((req.FileSize + chunkSize - 1) / chunkSize)

	// UUID V4

	fileID := generateUUID()

	S3ObjectKey := strconv.FormatInt(ownerID, 10) + "/" + fileID.String()

	newFile := FileMetadata{
		FileID:              fileID.String(),
		FileName:            req.FileName,
		Extension:           req.Extension,
		MimeType:            req.MimeType,
		ParentFolderID:      req.ParentFolderID,
		OwnerID:             ownerID,
		SizeBytes:           req.FileSize,
		Checksum:            req.Checksum,
		TotalChunks:         totalChunks,
		UploadedChunksCount: 0,
		S3ObjectKey:         S3ObjectKey,
		CompressionAlgo:     req.CompressionAlgo,
		Status:              FileUploadStatusUploading,
	}

	err := DB.QueryRow(context.Background(), "INSERT INTO file_metadata (file_id, owner_id, file_name, extension, mime_type, parent_folder_id, size_bytes, checksum, total_chunks, uploaded_chunks_count, s3_object_key, compression_algorithm, status) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) RETURNING file_id", newFile.FileID, newFile.OwnerID, newFile.FileName, newFile.Extension, newFile.MimeType, newFile.ParentFolderID, newFile.SizeBytes, newFile.Checksum, newFile.TotalChunks, newFile.UploadedChunksCount, newFile.S3ObjectKey, string(newFile.CompressionAlgo), string(newFile.Status)).Scan(&newFile.FileID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create file record")
	}

	client, err := NewS3Client(context.Background())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create S3 client")
	}

	// Step 1: Create multipart upload
	createResp, err := client.CreateMultipartUpload(context.Background(), &s3.CreateMultipartUploadInput{
		Bucket: aws.String(bucketName()),
		Key:    aws.String(newFile.S3ObjectKey),
	})
	if err != nil {
		return err
	}

	if createResp.UploadId == nil {
		return fiber.NewError(fiber.StatusInternalServerError, "S3 did not return multipart upload ID")
	}
	multipartUploadID := *createResp.UploadId

	// Step 2: Save upload ID in DB
	_, err = DB.Exec(context.Background(), "UPDATE file_metadata SET multipart_upload_id = $1 WHERE file_id = $2", multipartUploadID, newFile.FileID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to save multipart upload ID")
	}

	// Step 3: Generate Chunks
	// PreSigned URL and Etag will be added later
	// Save Chunks to db

	var chunks []FileChunk

	for _, chunkReq := range req.Chunks {
		chunk := FileChunk{
			FileID:            newFile.FileID,
			ChunkIndex:        chunkReq.ChunkIndex,
			Checksum:          chunkReq.Checksum,
			ChunkSize:         chunkReq.ChunkSize,
			ETag:              nil,
			MultipartUploadID: multipartUploadID,
			S3ObjectKey:       newFile.S3ObjectKey,
			Status:            FileChunkUploadStatusPending,
		}

		chunks = append(chunks, chunk)
	}

	log.Printf("Chunks to be uploaded: %d", len(chunks))

	// Insert chunk rows with a multi-row INSERT (text protocol) rather than a
	// binary CopyFrom. CopyFrom encodes by Go type into Postgres' binary COPY
	// format, which mismatches the column types here (string vs UUID file_id,
	// Go int vs INTEGER chunk_index) and fails. INSERT lets Postgres cast each
	// value per column. Batch to stay well under the 65535-parameter limit.
	const (
		colsPerChunk    = 6
		chunksPerInsert = 1000
	)
	for start := 0; start < len(chunks); start += chunksPerInsert {
		end := start + chunksPerInsert
		if end > len(chunks) {
			end = len(chunks)
		}
		batch := chunks[start:end]

		placeholders := make([]string, 0, len(batch))
		args := make([]any, 0, len(batch)*colsPerChunk)
		for i, chunk := range batch {
			var etag any
			if chunk.ETag != nil {
				etag = *chunk.ETag
			}
			n := i * colsPerChunk
			placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d)", n+1, n+2, n+3, n+4, n+5, n+6))
			args = append(args, chunk.FileID, chunk.ChunkIndex, chunk.Checksum, chunk.ChunkSize, etag, string(chunk.Status))
		}

		query := "INSERT INTO file_chunks (file_id, chunk_index, checksum, chunk_size, etag, status) VALUES " + strings.Join(placeholders, ", ")
		if _, err = DB.Exec(context.Background(), query, args...); err != nil {
			log.Printf("Insert file_chunks failed (file_id=%s, chunks=%d): %v", newFile.FileID, len(batch), err)
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to create chunk records")
		}
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Upload initiated successfully",
		"file_id": newFile.FileID,
		"chunks":  chunks,
	})

}

func getPresignedURLHandler(c *fiber.Ctx) error {
	// Get file_id and chunk_index from URL params
	fileID := c.Params("file_id")
	chunkIndexStr := c.Params("chunk_index")

	chunkIndex, err := strconv.Atoi(chunkIndexStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid chunk index")
	}

	// Validate Ownership of the file
	ownerID := getUserID(c)

	var existingFileID string
	err = DB.QueryRow(context.Background(), "SELECT file_id FROM file_metadata WHERE file_id = $1 AND owner_id = $2", fileID, ownerID).Scan(&existingFileID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "File not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to query file")
	}

	// chunk data from DB

	var fileChunk FileChunk
	var multipartUploadID sql.NullString

	err = DB.QueryRow(context.Background(), `
		SELECT
			fc.file_id,
			fc.chunk_index,
			fc.checksum,
			fc.chunk_size,
			fm.s3_object_key,
			fm.multipart_upload_id,
			fc.etag,
			fc.status,
			fc.created_at,
			fc.updated_at,
			fc.deleted_at
		FROM file_chunks fc
		JOIN file_metadata fm ON fm.file_id = fc.file_id
		WHERE fc.file_id = $1 AND fc.chunk_index = $2
	`, fileID, chunkIndex).Scan(
		&fileChunk.FileID,
		&fileChunk.ChunkIndex,
		&fileChunk.Checksum,
		&fileChunk.ChunkSize,
		&fileChunk.S3ObjectKey,
		&multipartUploadID,
		&fileChunk.ETag,
		&fileChunk.Status,
		&fileChunk.CreatedAt,
		&fileChunk.UpdatedAt,
		&fileChunk.DeletedAt,
	)
	// fileChunk.ChunkIndex starts from 1
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "Chunk not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to query chunk")
	}
	if !multipartUploadID.Valid {
		return fiber.NewError(fiber.StatusBadRequest, "Multipart upload has not been initialized")
	}
	fileChunk.MultipartUploadID = multipartUploadID.String

	client, err := NewS3Client(context.Background())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create S3 client")
	}

	// Generate presigned URL for the chunk
	presignClient := s3.NewPresignClient(client)

	presignResp, err := presignClient.PresignUploadPart(context.Background(), &s3.UploadPartInput{
		Bucket:        aws.String(bucketName()),
		Key:           aws.String(fileChunk.S3ObjectKey),
		UploadId:      aws.String(fileChunk.MultipartUploadID),
		PartNumber:    aws.Int32(int32(fileChunk.ChunkIndex)),
		ContentLength: aws.Int64(fileChunk.ChunkSize),
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to generate presigned URL")
	}

	// Update chunk status to "uploading"
	_, err = DB.Exec(context.Background(), "UPDATE file_chunks SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE file_id = $2 AND chunk_index = $3", FileChunkUploadStatusUploading, fileChunk.FileID, fileChunk.ChunkIndex)
	if err != nil {
		log.Printf("Failed to update chunk status: %v", err)
		// Not returning error to avoid failing the upload due to DB issue
	}

	return c.JSON(fiber.Map{
		"presigned_url": presignResp.URL,
	})
}

func updateFileStatusHandler(c *fiber.Ctx) error {
	// Get file_id and status from URL params
	fileID := c.Params("file_id")
	status := c.Params("status")

	// verify status is valid
	if status != string(FileUploadStatusUploading) && status != string(FileUploadStatusUploaded) && status != string(FileUploadStatusFailed) {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid status value")
	}

	// Validate Ownership of the file
	ownerID := getUserID(c)

	var existingFileID string
	var parentFolderID *string
	var s3ObjectKey string
	var totalChunks int
	var multipartUploadID sql.NullString
	err := DB.QueryRow(context.Background(), "SELECT file_id, parent_folder_id, s3_object_key, total_chunks, multipart_upload_id FROM file_metadata WHERE file_id = $1 AND owner_id = $2", fileID, ownerID).Scan(&existingFileID, &parentFolderID, &s3ObjectKey, &totalChunks, &multipartUploadID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "File not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to query file")
	}

	// When marking a file as fully uploaded, complete the underlying S3 multipart
	// upload so the object actually materializes and becomes downloadable.
	if status == string(FileUploadStatusUploaded) {
		if !multipartUploadID.Valid {
			return fiber.NewError(fiber.StatusBadRequest, "Multipart upload has not been initialized")
		}
		if err := completeMultipartUpload(context.Background(), s3ObjectKey, multipartUploadID.String, fileID, totalChunks); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to complete upload: "+err.Error())
		}
	}

	// Update file status. On completion, also reflect the uploaded chunk count.
	if status == string(FileUploadStatusUploaded) {
		_, err = DB.Exec(context.Background(), "UPDATE file_metadata SET status = $1, uploaded_chunks_count = total_chunks, updated_at = CURRENT_TIMESTAMP WHERE file_id = $2", status, fileID)
	} else {
		_, err = DB.Exec(context.Background(), "UPDATE file_metadata SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE file_id = $2", status, fileID)
	}
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update file status")
	}

	if status == string(FileUploadStatusUploaded) {
		// Create file change log
		err = createFileChangeLog(ownerID, EntityTypeFile, fileID, parentFolderID, FileChangeActionCreated)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to create file change log")
		}
	}

	return c.JSON(fiber.Map{
		"message": "File status updated successfully",
	})
}

// completeMultipartUpload assembles the uploaded part ETags for a file and calls
// S3 CompleteMultipartUpload. It errors if any chunk is missing an ETag (i.e. has
// not been uploaded), since S3 requires every part to be present.
func completeMultipartUpload(ctx context.Context, s3ObjectKey, multipartUploadID, fileID string, totalChunks int) error {
	rows, err := DB.Query(ctx, "SELECT chunk_index, etag FROM file_chunks WHERE file_id = $1 ORDER BY chunk_index ASC", fileID)
	if err != nil {
		return err
	}
	defer rows.Close()

	var parts []s3types.CompletedPart
	for rows.Next() {
		var chunkIndex int
		var etag sql.NullString
		if err := rows.Scan(&chunkIndex, &etag); err != nil {
			return err
		}
		if !etag.Valid || etag.String == "" {
			return fmt.Errorf("chunk %d has not been uploaded", chunkIndex)
		}
		parts = append(parts, s3types.CompletedPart{
			ETag:       aws.String(etag.String),
			PartNumber: aws.Int32(int32(chunkIndex)),
		})
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if len(parts) != totalChunks {
		return fmt.Errorf("expected %d uploaded chunks, found %d", totalChunks, len(parts))
	}

	client, err := NewS3Client(ctx)
	if err != nil {
		return err
	}

	_, err = client.CompleteMultipartUpload(ctx, &s3.CompleteMultipartUploadInput{
		Bucket:          aws.String(bucketName()),
		Key:             aws.String(s3ObjectKey),
		UploadId:        aws.String(multipartUploadID),
		MultipartUpload: &s3types.CompletedMultipartUpload{Parts: parts},
	})
	return err
}

func updateFileChunkStatusHandler(c *fiber.Ctx) error {
	// Get file_id and chunk_index from URL params
	fileID := c.Params("file_id")
	chunkIndexStr := c.Params("chunk_index")
	status := c.Params("status")
	etag := c.Query("etag")

	if status == string(FileChunkUploadStatusUploaded) && etag == "" {
		return fiber.NewError(fiber.StatusBadRequest, "ETag is required when marking chunk as uploaded")
	}

	if status != string(FileChunkUploadStatusUploaded) && etag != "" {
		return fiber.NewError(fiber.StatusBadRequest, "ETag should not be provided when marking chunk as uploaded")
	}

	chunkIndex, err := strconv.Atoi(chunkIndexStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid chunk index")
	}

	// verify status is valid
	if status != string(FileChunkUploadStatusUploading) && status != string(FileChunkUploadStatusUploaded) && status != string(FileChunkUploadStatusFailed) {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid status value")
	}

	// Validate Ownership of the file
	ownerID := getUserID(c)

	var existingFileID string
	err = DB.QueryRow(context.Background(), "SELECT file_id FROM file_metadata WHERE file_id = $1 AND owner_id = $2", fileID, ownerID).Scan(&existingFileID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "File not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to query file")
	}

	// Update chunk status and ETag
	_, err = DB.Exec(context.Background(), "UPDATE file_chunks SET status = $1, etag = $2, updated_at = CURRENT_TIMESTAMP WHERE file_id = $3 AND chunk_index = $4", status, etag, fileID, chunkIndex)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update chunk status")
	}

	return c.JSON(fiber.Map{
		"message": "Chunk status updated successfully",
	})
}

func getPendingChunksHandler(c *fiber.Ctx) error {
	// Get file_id from URL params
	fileID := c.Params("file_id")

	// Validate Ownership of the file
	ownerID := getUserID(c)

	var existingFileID string
	err := DB.QueryRow(context.Background(), "SELECT file_id FROM file_metadata WHERE file_id = $1 AND owner_id = $2", fileID, ownerID).Scan(&existingFileID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "File not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to query file")
	}

	rows, err := DB.Query(context.Background(), "SELECT chunk_index FROM file_chunks WHERE file_id = $1 AND status != $2", fileID, FileChunkUploadStatusUploaded)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to query chunks")
	}
	defer rows.Close()

	var pendingChunks []int
	for rows.Next() {
		var chunkIndex int
		if err := rows.Scan(&chunkIndex); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to scan chunk index")
		}
		pendingChunks = append(pendingChunks, chunkIndex)
	}

	return c.JSON(fiber.Map{
		"pending_chunks": pendingChunks,
	})
}

func fetchFileMetadataHandler(c *fiber.Ctx) error {
	// Get file_id from URL params
	fileID := c.Params("file_id")

	// Validate Ownership of the file
	ownerID := getUserID(c)

	var fileMeta FileMetadata
	err := DB.QueryRow(context.Background(), "SELECT file_id, file_name, extension, mime_type, parent_folder_id, owner_id, size_bytes, checksum, total_chunks, uploaded_chunks_count, s3_object_key, compression_algorithm, status FROM file_metadata WHERE file_id = $1 AND owner_id = $2", fileID, ownerID).Scan(&fileMeta.FileID, &fileMeta.FileName, &fileMeta.Extension, &fileMeta.MimeType, &fileMeta.ParentFolderID, &fileMeta.OwnerID, &fileMeta.SizeBytes, &fileMeta.Checksum, &fileMeta.TotalChunks, &fileMeta.UploadedChunksCount, &fileMeta.S3ObjectKey, &fileMeta.CompressionAlgo, &fileMeta.Status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "File not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to query file metadata")
	}

	return c.JSON(fileMeta)
}

func generateDownloadURLHandler(c *fiber.Ctx) error {
	// Get file_id from URL params
	fileID := c.Params("file_id")

	// Validate Ownership of the file
	ownerID := getUserID(c)

	var s3ObjectKey string
	err := DB.QueryRow(context.Background(), "SELECT s3_object_key FROM file_metadata WHERE file_id = $1 AND owner_id = $2", fileID, ownerID).Scan(&s3ObjectKey)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "File not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to query file metadata")
	}

	client, err := NewS3Client(context.Background())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create S3 client")
	}

	presignClient := s3.NewPresignClient(client)

	presignResp, err := presignClient.PresignGetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(bucketName()),
		Key:    aws.String(s3ObjectKey),
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to generate presigned URL")
	}

	return c.JSON(fiber.Map{
		"download_url": presignResp.URL,
	})
}

func renameFileHandler(c *fiber.Ctx) error {
	// Get file_id from URL params
	fileID := c.Params("file_id")

	type RenameRequest struct {
		NewFileName string `json:"new_file_name"`
	}

	var req RenameRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if req.NewFileName == "" {
		return fiber.NewError(fiber.StatusBadRequest, "new_file_name is required")
	}

	// Validate Ownership of the file
	ownerID := getUserID(c)

	var existingFileID string
	var parentFolderID *string
	err := DB.QueryRow(context.Background(), "SELECT file_id, parent_folder_id FROM file_metadata WHERE file_id = $1 AND owner_id = $2", fileID, ownerID).Scan(&existingFileID, &parentFolderID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "File not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to query file")
	}

	// Update file name
	_, err = DB.Exec(context.Background(), "UPDATE file_metadata SET file_name = $1, updated_at = CURRENT_TIMESTAMP WHERE file_id = $2", req.NewFileName, fileID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to rename file")
	}

	// Create file change log
	err = createFileChangeLog(ownerID, EntityTypeFile, fileID, parentFolderID, FileChangeActionRenamed)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create file change log")
	}

	return c.JSON(fiber.Map{
		"message": "File renamed successfully",
	})
}

func moveFileHandler(c *fiber.Ctx) error {
	// Get file_id from URL params
	fileID := c.Params("file_id")

	type MoveRequest struct {
		NewParentFolderID *string `json:"new_parent_folder_id,omitempty"`
	}

	var req MoveRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// Validate Ownership of the file
	ownerID := getUserID(c)

	var existingFileID string
	var parentFolderID *string
	err := DB.QueryRow(context.Background(), "SELECT file_id, parent_folder_id FROM file_metadata WHERE file_id = $1 AND owner_id = $2", fileID, ownerID).Scan(&existingFileID, &parentFolderID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "File not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to query file")
	}

	// If new parent folder ID is provided, validate it belongs to the user
	if req.NewParentFolderID != nil {
		var existingFolderID string
		err := DB.QueryRow(context.Background(), "SELECT folder_id FROM folders WHERE folder_id = $1 AND owner_id = $2", *req.NewParentFolderID, ownerID).Scan(&existingFolderID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return fiber.NewError(fiber.StatusNotFound, "New parent folder not found")
			}
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to query new parent folder")
		}
	}

	// Update parent folder ID
	_, err = DB.Exec(context.Background(), "UPDATE file_metadata SET parent_folder_id = $1, updated_at = CURRENT_TIMESTAMP WHERE file_id = $2", req.NewParentFolderID, fileID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to move file")
	}

	// Create file change log
	err = createFileChangeLog(ownerID, EntityTypeFile, fileID, parentFolderID, FileChangeActionMoved)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create file change log")
	}

	return c.JSON(fiber.Map{
		"message": "File moved successfully",
	})
}

func deleteFileHandler(c *fiber.Ctx) error {
	// Get file_id from URL params
	fileID := c.Params("file_id")

	// Validate Ownership of the file
	ownerID := getUserID(c)

	var existingFileID string
	var parentFolderID *string
	err := DB.QueryRow(context.Background(), "SELECT file_id, parent_folder_id FROM file_metadata WHERE file_id = $1 AND owner_id = $2", fileID, ownerID).Scan(&existingFileID, &parentFolderID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "File not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to query file")
	}

	// Soft delete: Mark the file as deleted in DB. Move to a "Trash" folder
	trashFolderID, err := getOrCreateTrashFolder(ownerID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to resolve Trash folder")
	}

	// Move file to trash by updating parent_folder_id
	_, err = DB.Exec(context.Background(), "UPDATE file_metadata SET parent_folder_id = $1, status = $2, updated_at = CURRENT_TIMESTAMP WHERE file_id = $3", trashFolderID, FileUploadStatusDeleted, fileID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to delete file")
	}

	// Create file change log
	err = createFileChangeLog(ownerID, EntityTypeFile, fileID, parentFolderID, FileChangeActionDeleted)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create file change log")
	}

	return c.JSON(fiber.Map{
		"message": "File deleted successfully",
	})
}
