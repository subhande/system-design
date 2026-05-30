package transfer

import (
	"bytes"
	"context"
	"fmt"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/desubhan/system-design/lumo-drive/client/internal/api"
	"github.com/desubhan/system-design/lumo-drive/client/internal/state"
)

// s3HTTPClient is used for direct transfers to S3 presigned URLs. Parts can be
// up to 5 MiB so we allow a generous timeout.
var s3HTTPClient = &http.Client{Timeout: 5 * time.Minute}

// UploadResult is returned after a successful upload.
type UploadResult struct {
	FileID   string
	Checksum string
	Size     int64
}

// UploadFile uploads a local file to the server using the chunked/multipart
// flow and returns the new remote file id. It journals each uploaded chunk in
// the store so a later call can resume via ResumeUpload.
func UploadFile(ctx context.Context, c *api.Client, st *state.Store, localPath, relPath string, parentFolderID *string) (*UploadResult, error) {
	plan, err := PlanFile(localPath)
	if err != nil {
		return nil, err
	}

	name := filepath.Base(relPath)
	ext := strings.TrimPrefix(filepath.Ext(name), ".")
	mimeType := mime.TypeByExtension(filepath.Ext(name))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	chunkReqs := make([]api.ChunkRequest, 0, len(plan.Chunks))
	for _, ch := range plan.Chunks {
		chunkReqs = append(chunkReqs, api.ChunkRequest{
			ChunkIndex: ch.Index,
			Checksum:   ch.Checksum,
			ChunkSize:  ch.Size,
		})
	}

	fileID, err := c.InitiateUpload(ctx, api.InitiateUploadRequest{
		FileName:        name,
		Extension:       ext,
		MimeType:        mimeType,
		ParentFolderID:  parentFolderID,
		FileSize:        plan.Size,
		Checksum:        plan.Checksum,
		CompressionAlgo: "none",
		Chunks:          chunkReqs,
	})
	if err != nil {
		return nil, fmt.Errorf("initiate upload: %w", err)
	}

	if err := uploadChunks(ctx, c, st, fileID, relPath, localPath, plan.Chunks); err != nil {
		return nil, err
	}

	if err := c.MarkFileStatus(ctx, fileID, api.FileStatusUploaded); err != nil {
		return nil, fmt.Errorf("mark file uploaded: %w", err)
	}
	_ = st.ClearPendingUploads(fileID)

	return &UploadResult{FileID: fileID, Checksum: plan.Checksum, Size: plan.Size}, nil
}

// ResumeUpload re-uploads only the chunks the server still reports as pending
// for an in-flight file, then marks the file uploaded.
func ResumeUpload(ctx context.Context, c *api.Client, st *state.Store, fileID, relPath, localPath string) error {
	pending, err := c.GetPendingChunks(ctx, fileID)
	if err != nil {
		return fmt.Errorf("get pending chunks: %w", err)
	}

	plan, err := PlanFile(localPath)
	if err != nil {
		return err
	}

	pendingSet := make(map[int]bool, len(pending))
	for _, idx := range pending {
		pendingSet[idx] = true
	}
	var todo []Chunk
	for _, ch := range plan.Chunks {
		if pendingSet[ch.Index] {
			todo = append(todo, ch)
		}
	}

	if err := uploadChunks(ctx, c, st, fileID, relPath, localPath, todo); err != nil {
		return err
	}
	if err := c.MarkFileStatus(ctx, fileID, api.FileStatusUploaded); err != nil {
		return fmt.Errorf("mark file uploaded: %w", err)
	}
	_ = st.ClearPendingUploads(fileID)
	return nil
}

// uploadChunks uploads the given chunks: presign -> PUT to S3 -> mark uploaded.
func uploadChunks(ctx context.Context, c *api.Client, st *state.Store, fileID, relPath, localPath string, chunks []Chunk) error {
	f, err := openForRead(localPath)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, ch := range chunks {
		data, err := readChunk(f, ch)
		if err != nil {
			return err
		}

		url, err := c.GetChunkPresignedURL(ctx, fileID, ch.Index)
		if err != nil {
			return fmt.Errorf("presign chunk %d: %w", ch.Index, err)
		}

		etag, err := putToS3(ctx, url, data)
		if err != nil {
			_ = c.MarkChunkStatus(ctx, fileID, ch.Index, api.ChunkStatusFailed, "")
			return fmt.Errorf("upload chunk %d: %w", ch.Index, err)
		}

		if err := c.MarkChunkStatus(ctx, fileID, ch.Index, api.ChunkStatusUploaded, etag); err != nil {
			return fmt.Errorf("mark chunk %d uploaded: %w", ch.Index, err)
		}
		if err := st.SetChunkUploaded(fileID, relPath, ch.Index, etag); err != nil {
			return err
		}
	}
	return nil
}

// putToS3 PUTs a chunk's bytes to a presigned URL and returns the ETag S3
// assigns to the uploaded part.
func putToS3(ctx context.Context, presignedURL string, data []byte) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, presignedURL, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.ContentLength = int64(len(data))

	resp, err := s3HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("S3 returned status %d", resp.StatusCode)
	}

	etag := resp.Header.Get("ETag")
	if etag == "" {
		return "", fmt.Errorf("S3 did not return an ETag")
	}
	return etag, nil
}
