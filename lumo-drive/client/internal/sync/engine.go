// Package sync reconciles a local sync directory with the Lumo Drive server.
//
// Conflict policy: the remote is the source of truth. When a file changed both
// locally and remotely since the last sync, the most recently edited version
// wins; on a tie the remote wins (so remote is preferred). This is implemented
// in resolveRemoteWins.
package sync

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/desubhan/system-design/lumo-drive/client/internal/api"
	"github.com/desubhan/system-design/lumo-drive/client/internal/state"
)

// Engine performs sync operations against the server for one sync directory.
type Engine struct {
	api     *api.Client
	store   *state.Store
	syncDir string

	// Logf logs progress. Defaults to a no-op; callers may override.
	Logf func(format string, args ...any)
}

// New builds an Engine.
func New(c *api.Client, st *state.Store, syncDir string) *Engine {
	return &Engine{
		api:     c,
		store:   st,
		syncDir: syncDir,
		Logf:    func(string, ...any) {},
	}
}

// Stats summarizes the work performed by a sync operation.
type Stats struct {
	Uploaded   int
	Downloaded int
	Deleted    int // entities deleted remotely (local deletions propagated)
	Removed    int // local files removed (remote deletions applied)
	Conflicts  int
	Skipped    int
}

// Add merges another Stats into the receiver.
func (s *Stats) Add(o Stats) {
	s.Uploaded += o.Uploaded
	s.Downloaded += o.Downloaded
	s.Deleted += o.Deleted
	s.Removed += o.Removed
	s.Conflicts += o.Conflicts
	s.Skipped += o.Skipped
}

// Sync runs a full reconciliation: pull remote changes first (remote is the
// source of truth), then push outstanding local changes.
func (e *Engine) Sync(ctx context.Context) (Stats, error) {
	var total Stats

	pulled, err := e.Pull(ctx)
	if err != nil {
		return total, err
	}
	total.Add(pulled)

	pushed, err := e.Push(ctx)
	if err != nil {
		return total, err
	}
	total.Add(pushed)
	return total, nil
}

// isHidden reports whether a relative path should be excluded from sync because
// any of its segments is a dotfile or dot-directory (e.g. .DS_Store, .git,
// .lumo). Such entries are never uploaded, watched, or deleted remotely.
func isHidden(rel string) bool {
	for _, seg := range strings.Split(filepath.ToSlash(rel), "/") {
		if seg != "" && seg != "." && strings.HasPrefix(seg, ".") {
			return true
		}
	}
	return false
}

// --- shared filesystem <-> remote mapping helpers -------------------------

// joinRel joins a parent relative path and a name with forward slashes.
func joinRel(parent, name string) string {
	if parent == "" {
		return name
	}
	return parent + "/" + name
}

// abs converts a forward-slash relative path to an absolute local path.
func (e *Engine) abs(rel string) string {
	return filepath.Join(e.syncDir, filepath.FromSlash(rel))
}

// folderIDForPath maps a relative directory path to its remote folder id (push
// direction). The sync root (".") maps to nil.
func (e *Engine) folderIDForPath(rel string) (*string, error) {
	if rel == "." || rel == "" {
		return nil, nil
	}
	f, err := e.store.FolderByPath(rel)
	if err != nil {
		return nil, err
	}
	if f == nil {
		return nil, fmt.Errorf("internal: no remote folder mapping for %q", rel)
	}
	id := f.RemoteID
	return &id, nil
}

// ensureRemoteFolder returns the remote id for a relative directory path,
// creating the remote folder (and recording the mapping) if needed (push).
func (e *Engine) ensureRemoteFolder(ctx context.Context, rel string) (string, error) {
	if existing, err := e.store.FolderByPath(rel); err != nil {
		return "", err
	} else if existing != nil {
		return existing.RemoteID, nil
	}

	parentID, err := e.folderIDForPath(filepath.ToSlash(filepath.Dir(rel)))
	if err != nil {
		return "", err
	}

	name := filepath.Base(rel)
	folderID, err := e.api.CreateFolder(ctx, name, parentID)
	if err != nil {
		return "", fmt.Errorf("create folder %s: %w", rel, err)
	}

	parent := ""
	if parentID != nil {
		parent = *parentID
	}
	if err := e.store.UpsertFolder(state.FolderRecord{
		RemoteID:       folderID,
		ParentRemoteID: parent,
		Name:           name,
		RelPath:        rel,
	}); err != nil {
		return "", err
	}
	return folderID, nil
}

// folderRelPath resolves a remote folder id to its relative path under the sync
// root (pull direction), fetching missing ancestors, creating local
// directories, and caching the mapping. The reserved "Trash" folder and the
// root map to "".
func (e *Engine) folderRelPath(ctx context.Context, folderID string) (string, error) {
	if folderID == "" {
		return "", nil
	}

	folder, err := e.api.GetFolder(ctx, folderID)
	if err != nil {
		return "", fmt.Errorf("get folder %s: %w", folderID, err)
	}
	if folder.FolderName == "Trash" {
		return "", nil
	}

	parentRel := ""
	if folder.ParentFolderID != nil {
		parentRel, err = e.folderRelPath(ctx, *folder.ParentFolderID)
		if err != nil {
			return "", err
		}
	}
	rel := joinRel(parentRel, folder.FolderName)

	if err := os.MkdirAll(e.abs(rel), 0o755); err != nil {
		return "", err
	}

	parent := ""
	if folder.ParentFolderID != nil {
		parent = *folder.ParentFolderID
	}
	if err := e.store.UpsertFolder(state.FolderRecord{
		RemoteID:       folder.FolderID,
		ParentRemoteID: parent,
		Name:           folder.FolderName,
		RelPath:        rel,
	}); err != nil {
		return "", err
	}
	return rel, nil
}
