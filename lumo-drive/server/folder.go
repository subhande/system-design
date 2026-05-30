package main

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v4"
)

type CreateFolderRequest struct {
	FolderName     string  `json:"folder_name"`
	ParentFolderID *string `json:"parent_folder_id,omitempty"`
}

type RenameFolderRequest struct {
	FolderID string `json:"folder_id"`
	NewName  string `json:"new_name"`
}

type MoveFolderRequest struct {
	FolderID    string  `json:"folder_id"`
	NewParentID *string `json:"new_parent_id,omitempty"`
}

func createFolderHandler(c *fiber.Ctx) error {

	var req CreateFolderRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if req.FolderName == "" {
		return fiber.NewError(fiber.StatusBadRequest, "folder_name is required")
	}

	if req.FolderName == "Trash" {
		return fiber.NewError(fiber.StatusBadRequest, "Folder name 'Trash' is reserved")
	}

	newFolder := Folder{
		FolderName:     req.FolderName,
		ParentFolderID: req.ParentFolderID,
		OwnerID:        getUserID(c),
	}

	// Insert into DB and return new folder ID

	var folderID string
	err := DB.QueryRow(context.Background(), "INSERT INTO folders (owner_id, name, parent_folder_id) VALUES ($1, $2, $3) RETURNING folder_id", newFolder.OwnerID, newFolder.FolderName, newFolder.ParentFolderID).Scan(&folderID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create folder")
	}

	// Create folder change log
	err = createFolderChangeLog(newFolder.OwnerID, EntityTypeFolder, folderID, newFolder.ParentFolderID, FileChangeActionCreated)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create folder change log")
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":   "Folder created successfully",
		"folder_id": folderID,
	})
}

func renameFolderHandler(c *fiber.Ctx) error {

	var req RenameFolderRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if req.FolderID == "" || req.NewName == "" {
		return fiber.NewError(fiber.StatusBadRequest, "folder_id and new_name are required")
	}

	if req.NewName == "Trash" {
		return fiber.NewError(fiber.StatusBadRequest, "Folder name 'Trash' is reserved")
	}

	ownerID := getUserID(c)

	// Check if folder exists and belongs to user
	var existingID string
	var parentFolderID *string
	err := DB.QueryRow(context.Background(), "SELECT folder_id, parent_folder_id FROM folders WHERE folder_id = $1 AND owner_id = $2", req.FolderID, ownerID).Scan(&existingID, &parentFolderID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "Folder not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to query folder")
	}

	// Update folder name
	_, err = DB.Exec(context.Background(), "UPDATE folders SET name = $1, updated_at = CURRENT_TIMESTAMP WHERE folder_id = $2", req.NewName, req.FolderID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to rename folder")
	}

	// Create folder change log
	err = createFolderChangeLog(ownerID, EntityTypeFolder, req.FolderID, parentFolderID, FileChangeActionRenamed)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create folder change log")
	}

	return c.JSON(fiber.Map{
		"message": "Folder renamed successfully",
	})
}

func moveFolderHandler(c *fiber.Ctx) error {

	var req MoveFolderRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if req.FolderID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "folder_id is required")
	}

	ownerID := getUserID(c)

	// Check if folder exists and belongs to user
	var existingID string
	err := DB.QueryRow(context.Background(), "SELECT folder_id FROM folders WHERE folder_id = $1 AND owner_id = $2", req.FolderID, ownerID).Scan(&existingID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "Folder not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to query folder")
	}

	// Update parent folder ID
	_, err = DB.Exec(context.Background(), "UPDATE folders SET parent_folder_id = $1, updated_at = CURRENT_TIMESTAMP WHERE folder_id = $2", req.NewParentID, req.FolderID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to move folder")
	}

	// Create folder change log
	err = createFolderChangeLog(ownerID, EntityTypeFolder, req.FolderID, req.NewParentID, FileChangeActionMoved)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create folder change log")
	}

	return c.JSON(fiber.Map{
		"message": "Folder moved successfully",
	})
}

func deleteFolderHandler(c *fiber.Ctx) error {

	folderID := c.Params("folder_id")
	if folderID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "folder_id is required")
	}

	ownerID := getUserID(c)

	// Move to trash instead of hard delete
	trashFolderID, err := getOrCreateTrashFolder(ownerID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to resolve Trash folder")
	}

	// Move folder to trash by updating parent_folder_id
	_, err = DB.Exec(context.Background(), "UPDATE folders SET parent_folder_id = $1, updated_at = CURRENT_TIMESTAMP WHERE folder_id = $2", trashFolderID, folderID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to delete folder")
	}

	// Create folder change log
	err = createFolderChangeLog(ownerID, EntityTypeFolder, folderID, &trashFolderID, FileChangeActionDeleted)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create folder change log")
	}

	return c.JSON(fiber.Map{
		"message": "Folder deleted successfully",
	})
}

func listFolderContentsHandler(c *fiber.Ctx) error {

	folderID := c.Params("folder_id")
	if folderID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "folder_id is required")
	}

	ownerID := getUserID(c)

	// Check if folder exists and belongs to user
	var existingID string
	err := DB.QueryRow(context.Background(), "SELECT folder_id FROM folders WHERE folder_id = $1 AND owner_id = $2", folderID, ownerID).Scan(&existingID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "Folder not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to query folder")
	}

	// Query subfolders
	subfolders := []Folder{}
	rows, err := DB.Query(context.Background(), "SELECT folder_id, name FROM folders WHERE parent_folder_id = $1 AND owner_id = $2", folderID, ownerID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to query subfolders")
	}
	defer rows.Close()

	for rows.Next() {
		var f Folder
		if err := rows.Scan(&f.FolderID, &f.FolderName); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to scan subfolder")
		}
		subfolders = append(subfolders, f)
	}

	// Query files
	files := []FileMetadata{}
	rows, err = DB.Query(context.Background(), "SELECT file_id, file_name, extension FROM file_metadata WHERE parent_folder_id = $1 AND owner_id = $2", folderID, ownerID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to query files")
	}
	defer rows.Close()

	for rows.Next() {
		var f FileMetadata
		if err := rows.Scan(&f.FileID, &f.FileName, &f.Extension); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to scan file")
		}
		files = append(files, f)
	}

	return c.JSON(fiber.Map{
		"subfolders": subfolders,
		"files":      files,
	})
}

// getFolderHandler returns a single folder's metadata (id, name, parent). The
// client uses this to resolve folder names/parents when reconstructing the tree
// from the sync-changes feed.
func getFolderHandler(c *fiber.Ctx) error {
	folderID := c.Params("folder_id")
	if folderID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "folder_id is required")
	}

	ownerID := getUserID(c)

	var folder Folder
	err := DB.QueryRow(context.Background(),
		"SELECT folder_id, owner_id, name, parent_folder_id FROM folders WHERE folder_id = $1 AND owner_id = $2",
		folderID, ownerID,
	).Scan(&folder.FolderID, &folder.OwnerID, &folder.FolderName, &folder.ParentFolderID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "Folder not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to query folder")
	}

	return c.JSON(folder)
}

// getOrCreateTrashFolder returns the id of the user's reserved "Trash" folder,
// creating it on first use. Used when soft-deleting files and folders.
func getOrCreateTrashFolder(ownerID int64) (string, error) {
	var trashFolderID string
	err := DB.QueryRow(context.Background(), "SELECT folder_id FROM folders WHERE owner_id = $1 AND name = 'Trash'", ownerID).Scan(&trashFolderID)
	if err == nil {
		return trashFolderID, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return "", err
	}

	// No Trash folder yet for this user; create one.
	err = DB.QueryRow(context.Background(), "INSERT INTO folders (owner_id, name) VALUES ($1, 'Trash') RETURNING folder_id", ownerID).Scan(&trashFolderID)
	if err != nil {
		return "", err
	}
	return trashFolderID, nil
}
