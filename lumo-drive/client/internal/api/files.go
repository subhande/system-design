package api

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// InitiateUpload registers file metadata + chunk records and starts the S3
// multipart upload. Returns the new file id.
func (c *Client) InitiateUpload(ctx context.Context, req InitiateUploadRequest) (string, error) {
	var out InitiateUploadResponse
	if err := c.doJSON(ctx, http.MethodPost, "/api/v1/files/", req, &out); err != nil {
		return "", err
	}
	return out.FileID, nil
}

type presignResponse struct {
	PresignedURL string `json:"presigned_url"`
}

// GetChunkPresignedURL returns a presigned S3 UploadPart URL for one chunk.
func (c *Client) GetChunkPresignedURL(ctx context.Context, fileID string, chunkIndex int) (string, error) {
	var out presignResponse
	path := fmt.Sprintf("/api/v1/files/%s/chunks/%d/presign", url.PathEscape(fileID), chunkIndex)
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &out); err != nil {
		return "", err
	}
	return out.PresignedURL, nil
}

// MarkChunkStatus updates a chunk's status. When marking it uploaded, etag is
// required (it's the ETag S3 returned for the uploaded part).
func (c *Client) MarkChunkStatus(ctx context.Context, fileID string, chunkIndex int, status, etag string) error {
	path := fmt.Sprintf("/api/v1/files/%s/chunks/%d/status/%s", url.PathEscape(fileID), chunkIndex, url.PathEscape(status))
	if etag != "" {
		path += "?etag=" + url.QueryEscape(etag)
	}
	return c.doJSON(ctx, http.MethodPatch, path, nil, nil)
}

// MarkFileStatus updates a file's status. Marking it "uploaded" triggers the
// server to complete the S3 multipart upload.
func (c *Client) MarkFileStatus(ctx context.Context, fileID, status string) error {
	path := fmt.Sprintf("/api/v1/files/%s/status/%s", url.PathEscape(fileID), url.PathEscape(status))
	return c.doJSON(ctx, http.MethodPatch, path, nil, nil)
}

type pendingChunksResponse struct {
	PendingChunks []int `json:"pending_chunks"`
}

// GetPendingChunks returns the chunk indices that still need uploading.
func (c *Client) GetPendingChunks(ctx context.Context, fileID string) ([]int, error) {
	var out pendingChunksResponse
	path := fmt.Sprintf("/api/v1/files/%s/resume", url.PathEscape(fileID))
	if err := c.doJSON(ctx, http.MethodPost, path, nil, &out); err != nil {
		return nil, err
	}
	return out.PendingChunks, nil
}

// GetFileMetadata fetches a file's metadata record.
func (c *Client) GetFileMetadata(ctx context.Context, fileID string) (*FileMetadata, error) {
	var out FileMetadata
	if err := c.doJSON(ctx, http.MethodGet, "/api/v1/files/"+url.PathEscape(fileID), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type downloadURLResponse struct {
	DownloadURL string `json:"download_url"`
}

// GetDownloadURL returns a presigned S3 GET URL for the whole file.
func (c *Client) GetDownloadURL(ctx context.Context, fileID string) (string, error) {
	var out downloadURLResponse
	path := fmt.Sprintf("/api/v1/files/%s/download", url.PathEscape(fileID))
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &out); err != nil {
		return "", err
	}
	return out.DownloadURL, nil
}

type renameFileRequest struct {
	NewFileName string `json:"new_file_name"`
}

// RenameFile renames a file.
func (c *Client) RenameFile(ctx context.Context, fileID, newName string) error {
	path := fmt.Sprintf("/api/v1/files/%s/rename", url.PathEscape(fileID))
	return c.doJSON(ctx, http.MethodPatch, path, renameFileRequest{NewFileName: newName}, nil)
}

type moveFileRequest struct {
	NewParentFolderID *string `json:"new_parent_folder_id,omitempty"`
}

// MoveFile reparents a file.
func (c *Client) MoveFile(ctx context.Context, fileID string, newParentID *string) error {
	path := fmt.Sprintf("/api/v1/files/%s/move", url.PathEscape(fileID))
	return c.doJSON(ctx, http.MethodPost, path, moveFileRequest{NewParentFolderID: newParentID}, nil)
}

// DeleteFile soft-deletes a file (moves it to Trash).
func (c *Client) DeleteFile(ctx context.Context, fileID string) error {
	return c.doJSON(ctx, http.MethodDelete, "/api/v1/files/"+url.PathEscape(fileID), nil, nil)
}
