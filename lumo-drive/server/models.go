package main

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// UTCTime wraps time.Time so that timestamps are always scanned from the
// database, stored, and serialized to JSON in UTC (RFC3339). It implements
// sql.Scanner, driver.Valuer, and the JSON marshal/unmarshal interfaces.
type UTCTime time.Time

// Scan implements sql.Scanner. pgx hands timestamptz columns to us as a
// time.Time, which we normalize to UTC.
func (t *UTCTime) Scan(src interface{}) error {
	switch v := src.(type) {
	case time.Time:
		*t = UTCTime(v.UTC())
	case nil:
		*t = UTCTime{}
	default:
		return fmt.Errorf("unsupported Scan source for UTCTime: %T", src)
	}
	return nil
}

// Value implements driver.Valuer so the type can also be used as a query
// argument. A zero value is written as NULL.
func (t UTCTime) Value() (driver.Value, error) {
	tt := time.Time(t)
	if tt.IsZero() {
		return nil, nil
	}
	return tt.UTC(), nil
}

// MarshalJSON renders the timestamp as an RFC3339 string in UTC, or null when
// the value is the zero time.
func (t UTCTime) MarshalJSON() ([]byte, error) {
	tt := time.Time(t)
	if tt.IsZero() {
		return []byte("null"), nil
	}
	return []byte(`"` + tt.UTC().Format(time.RFC3339Nano) + `"`), nil
}

// UnmarshalJSON parses an RFC3339 string into a UTC timestamp.
func (t *UTCTime) UnmarshalJSON(data []byte) error {
	s := string(data)
	if s == "null" || s == `""` {
		*t = UTCTime{}
		return nil
	}
	parsed, err := time.Parse(`"`+time.RFC3339Nano+`"`, s)
	if err != nil {
		return err
	}
	*t = UTCTime(parsed.UTC())
	return nil
}

type User struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`

	PasswordHash string `json:"-"`

	CreatedAt UTCTime `json:"created_at"`
}

type Folder struct {
	FolderID string `json:"folder_id"`
	OwnerID  int64  `json:"owner_id"`

	FolderName     string  `json:"folder_name"`
	ParentFolderID *string `json:"parent_folder_id,omitempty"`

	CreatedAt UTCTime  `json:"created_at"`
	UpdatedAt UTCTime  `json:"updated_at"`
	DeletedAt *UTCTime `json:"deleted_at,omitempty"`
}

type FileUploadStatus string

const (
	FileUploadStatusPending   FileUploadStatus = "pending"
	FileUploadStatusUploading FileUploadStatus = "uploading"
	FileUploadStatusUploaded  FileUploadStatus = "uploaded"
	FileUploadStatusFailed    FileUploadStatus = "failed"
	FileUploadStatusDeleted   FileUploadStatus = "deleted"
)

type CompressionAlgorithm string

const (
	CompressionAlgorithmNone   CompressionAlgorithm = "none"
	CompressionAlgorithmGzip   CompressionAlgorithm = "gzip"
	CompressionAlgorithmBzip2  CompressionAlgorithm = "bzip2"
	CompressionAlgorithmLzma   CompressionAlgorithm = "lzma"
	CompressionAlgorithmZstd   CompressionAlgorithm = "zstd"
	CompressionAlgorithmSnappy CompressionAlgorithm = "snappy"
)

type FileMetadata struct {
	FileID  string `json:"file_id"`
	OwnerID int64  `json:"owner_id"`

	ParentFolderID *string `json:"parent_folder_id,omitempty"`

	FileName  string `json:"file_name"`
	Extension string `json:"extension"`
	MimeType  string `json:"mime_type"`

	SizeBytes int64 `json:"size_bytes"`

	Checksum string `json:"checksum"`

	TotalChunks         int `json:"total_chunks"`
	UploadedChunksCount int `json:"uploaded_chunks_count"`

	S3ObjectKey string `json:"s3_object_key"`

	MultipartUploadID *string `json:"multipart_upload_id,omitempty"`

	CompressionAlgo CompressionAlgorithm `json:"compression_algo"`

	Status FileUploadStatus `json:"status"`

	CreatedAt UTCTime  `json:"created_at"`
	UpdatedAt UTCTime  `json:"updated_at"`
	DeletedAt *UTCTime `json:"deleted_at,omitempty"`
}

type FileChunkUploadStatus string

const (
	FileChunkUploadStatusPending   FileChunkUploadStatus = "pending"
	FileChunkUploadStatusUploading FileChunkUploadStatus = "uploading"
	FileChunkUploadStatusUploaded  FileChunkUploadStatus = "uploaded"
	FileChunkUploadStatusFailed    FileChunkUploadStatus = "failed"
	FileChunkUploadStatusDeleted   FileChunkUploadStatus = "deleted"
)

type FileChunk struct {
	FileID     string `json:"file_id"`
	ChunkIndex int    `json:"chunk_index"`

	Checksum string `json:"checksum"`

	ChunkSize int64 `json:"chunk_size"`

	S3ObjectKey string `json:"s3_object_key"`

	MultipartUploadID string `json:"multipart_upload_id"`

	ETag *string `json:"etag,omitempty"`

	Status FileChunkUploadStatus `json:"status"`

	CreatedAt UTCTime  `json:"created_at"`
	UpdatedAt UTCTime  `json:"updated_at"`
	DeletedAt *UTCTime `json:"deleted_at,omitempty"`
}

type FileChangeAction string

const (
	FileChangeActionCreated FileChangeAction = "created"
	FileChangeActionUpdated FileChangeAction = "updated"
	FileChangeActionRenamed FileChangeAction = "renamed"
	FileChangeActionMoved   FileChangeAction = "moved"
	FileChangeActionDeleted FileChangeAction = "deleted"
)

type EntityType string

const (
	EntityTypeFile   EntityType = "file"
	EntityTypeFolder EntityType = "folder"
)

type FileChangeLog struct {
	ChangeID int64 `json:"change_id"`
	OwnerID  int64 `json:"owner_id"`

	EntityType     EntityType `json:"entity_type"`
	EntityID       string     `json:"entity_id"`
	ParentFolderID *string    `json:"parent_folder_id,omitempty"`

	Action FileChangeAction `json:"action"`

	CreatedAt UTCTime `json:"created_at"`
}
