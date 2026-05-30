package api

import "time"

// User mirrors the server user record (sans password hash).
type User struct {
	UserID    int64     `json:"user_id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// AuthResponse is returned by register and login.
type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// ChunkRequest describes one chunk in an InitiateUpload request.
type ChunkRequest struct {
	ChunkIndex int    `json:"chunk_index"`
	Checksum   string `json:"checksum"`
	ChunkSize  int64  `json:"chunk_size"`
}

// InitiateUploadRequest starts a chunked/multipart upload.
type InitiateUploadRequest struct {
	FileName        string         `json:"file_name"`
	Extension       string         `json:"extension"`
	MimeType        string         `json:"mime_type"`
	ParentFolderID  *string        `json:"parent_folder_id,omitempty"`
	FileSize        int64          `json:"file_size"`
	Checksum        string         `json:"checksum"`
	CompressionAlgo string         `json:"compression_algo"`
	Chunks          []ChunkRequest `json:"chunks"`
}

// InitiateUploadResponse carries the new file id back to the client.
type InitiateUploadResponse struct {
	Message string `json:"message"`
	FileID  string `json:"file_id"`
}

// FileMetadata mirrors the server's file_metadata record.
type FileMetadata struct {
	FileID              string  `json:"file_id"`
	OwnerID             int64   `json:"owner_id"`
	ParentFolderID      *string `json:"parent_folder_id,omitempty"`
	FileName            string  `json:"file_name"`
	Extension           string  `json:"extension"`
	MimeType            string  `json:"mime_type"`
	SizeBytes           int64   `json:"size_bytes"`
	Checksum            string  `json:"checksum"`
	TotalChunks         int     `json:"total_chunks"`
	UploadedChunksCount int     `json:"uploaded_chunks_count"`
	S3ObjectKey         string  `json:"s3_object_key"`
	CompressionAlgo     string  `json:"compression_algo"`
	Status              string  `json:"status"`
}

// FolderEntry is a folder as returned by folder-listing.
type FolderEntry struct {
	FolderID   string `json:"folder_id"`
	FolderName string `json:"folder_name"`
}

// FolderContents is the response from listing a folder.
type FolderContents struct {
	Subfolders []FolderEntry  `json:"subfolders"`
	Files      []FileMetadata `json:"files"`
}

// Change is one entry in the sync-changes feed.
type Change struct {
	ChangeID       int64     `json:"change_id"`
	EntityType     string    `json:"entity_type"` // "file" | "folder"
	EntityID       string    `json:"entity_id"`
	ParentFolderID *string   `json:"parent_folder_id,omitempty"`
	Action         string    `json:"action"` // created | updated | renamed | moved | deleted
	CreatedAt      time.Time `json:"created_at"`
}

// Entity types and actions in the sync feed.
const (
	EntityTypeFile   = "file"
	EntityTypeFolder = "folder"

	ActionCreated = "created"
	ActionUpdated = "updated"
	ActionRenamed = "renamed"
	ActionMoved   = "moved"
	ActionDeleted = "deleted"
)

// File upload statuses understood by the server.
const (
	FileStatusUploading = "uploading"
	FileStatusUploaded  = "uploaded"
	FileStatusFailed    = "failed"

	ChunkStatusUploading = "uploading"
	ChunkStatusUploaded  = "uploaded"
	ChunkStatusFailed    = "failed"
)
