package main

import (
	"context"

	"github.com/gofiber/fiber/v2"
)

// CREATE TABLE IF NOT EXISTS changes (
//     change_id BIGSERIAL PRIMARY KEY,
//     owner_id BIGINT NOT NULL REFERENCES users(user_id),

//     entity_type VARCHAR(20) NOT NULL,
//     entity_id UUID NOT NULL,
//     parent_folder_id UUID,
//     action VARCHAR(20) NOT NULL,
//     created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
// );

func createFolderChangeLog(ownerID int64, entityType EntityType, entityID string, parentFolderID *string, action FileChangeAction) error {
	log := FileChangeLog{
		OwnerID:        ownerID,
		EntityType:     entityType,
		EntityID:       entityID,
		ParentFolderID: parentFolderID,
		Action:         action,
	}

	return insertSyncEntry(log)
}

func createFileChangeLog(ownerID int64, entityType EntityType, entityID string, parentFolderID *string, action FileChangeAction) error {
	log := FileChangeLog{
		OwnerID:        ownerID,
		EntityType:     entityType,
		EntityID:       entityID,
		ParentFolderID: parentFolderID,
		Action:         action,
	}

	return insertSyncEntry(log)
}

func insertSyncEntry(log FileChangeLog) error {
	_, err := DB.Exec(context.Background(), "INSERT INTO changes (owner_id, entity_type, entity_id, parent_folder_id, action) VALUES ($1, $2, $3, $4, $5)", log.OwnerID, log.EntityType, log.EntityID, log.ParentFolderID, log.Action)
	return err
}

func getSyncEntriesForUser(ownerID int64, changeID *string) ([]FileChangeLog, error) {

	rows, err := DB.Query(context.Background(), "SELECT change_id, entity_type, entity_id, parent_folder_id, action, created_at FROM changes WHERE owner_id = $1 AND change_id > $2::BIGINT ORDER BY change_id ASC", ownerID, changeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []FileChangeLog
	for rows.Next() {
		var log FileChangeLog
		err := rows.Scan(&log.ChangeID, &log.EntityType, &log.EntityID, &log.ParentFolderID, &log.Action, &log.CreatedAt)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, nil

}

func syncHandler(c *fiber.Ctx) error {
	ownerID := getUserID(c)

	changeID := c.Params("change_id")

	logs, err := getSyncEntriesForUser(ownerID, &changeID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to fetch sync entries")
	}

	return c.JSON(fiber.Map{
		"changes": logs,
	})
}
