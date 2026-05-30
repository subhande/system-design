package api

import (
	"context"
	"net/http"
	"net/url"
)

type createFolderRequest struct {
	FolderName     string  `json:"folder_name"`
	ParentFolderID *string `json:"parent_folder_id,omitempty"`
}

type createFolderResponse struct {
	Message  string `json:"message"`
	FolderID string `json:"folder_id"`
}

// CreateFolder creates a folder and returns its id.
func (c *Client) CreateFolder(ctx context.Context, name string, parentID *string) (string, error) {
	var out createFolderResponse
	err := c.doJSON(ctx, http.MethodPost, "/api/v1/folders/",
		createFolderRequest{FolderName: name, ParentFolderID: parentID}, &out)
	if err != nil {
		return "", err
	}
	return out.FolderID, nil
}

// Folder is a folder's metadata record.
type Folder struct {
	FolderID       string  `json:"folder_id"`
	OwnerID        int64   `json:"owner_id"`
	FolderName     string  `json:"folder_name"`
	ParentFolderID *string `json:"parent_folder_id,omitempty"`
}

// GetFolder fetches a single folder's metadata (name + parent).
func (c *Client) GetFolder(ctx context.Context, folderID string) (*Folder, error) {
	var out Folder
	if err := c.doJSON(ctx, http.MethodGet, "/api/v1/folders/"+url.PathEscape(folderID), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListFolderContents returns the subfolders and files of a folder.
func (c *Client) ListFolderContents(ctx context.Context, folderID string) (*FolderContents, error) {
	var out FolderContents
	err := c.doJSON(ctx, http.MethodGet, "/api/v1/folders/"+url.PathEscape(folderID)+"/contents", nil, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

type renameFolderRequest struct {
	FolderID string `json:"folder_id"`
	NewName  string `json:"new_name"`
}

// RenameFolder renames a folder.
func (c *Client) RenameFolder(ctx context.Context, folderID, newName string) error {
	return c.doJSON(ctx, http.MethodPatch, "/api/v1/folders/"+url.PathEscape(folderID),
		renameFolderRequest{FolderID: folderID, NewName: newName}, nil)
}

type moveFolderRequest struct {
	FolderID    string  `json:"folder_id"`
	NewParentID *string `json:"new_parent_id,omitempty"`
}

// MoveFolder reparents a folder.
func (c *Client) MoveFolder(ctx context.Context, folderID string, newParentID *string) error {
	return c.doJSON(ctx, http.MethodPost, "/api/v1/folders/"+url.PathEscape(folderID)+"/move",
		moveFolderRequest{FolderID: folderID, NewParentID: newParentID}, nil)
}

// DeleteFolder soft-deletes a folder (moves it to Trash).
func (c *Client) DeleteFolder(ctx context.Context, folderID string) error {
	return c.doJSON(ctx, http.MethodDelete, "/api/v1/folders/"+url.PathEscape(folderID), nil, nil)
}
