package api

import (
	"context"
	"fmt"
	"net/http"
)

type changesResponse struct {
	Changes []Change `json:"changes"`
}

// GetChanges returns all change-log entries with change_id greater than sinceID,
// ordered ascending. Pass 0 to replay the full history.
func (c *Client) GetChanges(ctx context.Context, sinceID int64) ([]Change, error) {
	var out changesResponse
	path := fmt.Sprintf("/api/v1/sync/changes/%d", sinceID)
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return out.Changes, nil
}
