package sync

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/desubhan/system-design/lumo-drive/client/internal/api"
	"github.com/desubhan/system-design/lumo-drive/client/internal/state"
	"github.com/desubhan/system-design/lumo-drive/client/internal/transfer"
)

// Push walks the sync directory, creating remote folders and uploading new or
// changed files, then propagates local deletions to the server. Change
// detection is fast (size+mtime) with a checksum fallback.
func (e *Engine) Push(ctx context.Context) (Stats, error) {
	var st Stats
	seen := map[string]bool{}

	err := filepath.WalkDir(e.syncDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		rel, err := filepath.Rel(e.syncDir, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)

		if d.IsDir() {
			if d.Name() == state.DirName || isHidden(rel) {
				return fs.SkipDir
			}
			if rel == "." {
				return nil
			}
			_, ferr := e.ensureRemoteFolder(ctx, rel)
			return ferr
		}

		// Skip dotfiles such as .DS_Store.
		if isHidden(rel) {
			return nil
		}

		seen[rel] = true
		info, err := d.Info()
		if err != nil {
			return err
		}

		snap, err := e.store.FileByPath(rel)
		if err != nil {
			return err
		}

		// Fast path: unchanged by size + mtime.
		if snap != nil && !snap.Deleted &&
			snap.SizeBytes == info.Size() && snap.LocalMtime == info.ModTime().Unix() {
			st.Skipped++
			return nil
		}

		// Confirm with a checksum before uploading.
		sum, err := transfer.FileChecksum(path)
		if err != nil {
			return err
		}
		if snap != nil && !snap.Deleted && sum == snap.LocalChecksum {
			// Content identical; only mtime changed. Refresh the snapshot.
			snap.LocalMtime = info.ModTime().Unix()
			snap.SizeBytes = info.Size()
			st.Skipped++
			return e.store.UpsertFile(*snap)
		}

		if err := e.uploadFile(ctx, path, rel, info); err != nil {
			return err
		}
		st.Uploaded++
		return nil
	})
	if err != nil {
		return st, err
	}

	// Propagate local deletions: tracked files no longer present on disk.
	del, err := e.pushDeletions(ctx, seen)
	if err != nil {
		return st, err
	}
	st.Add(del)
	return st, nil
}

// uploadFile uploads one local file and records the synced snapshot.
func (e *Engine) uploadFile(ctx context.Context, path, rel string, info os.FileInfo) error {
	parentID, err := e.folderIDForPath(filepath.ToSlash(filepath.Dir(rel)))
	if err != nil {
		return err
	}

	res, err := transfer.UploadFile(ctx, e.api, e.store, path, rel, parentID)
	if err != nil {
		return fmt.Errorf("upload %s: %w", rel, err)
	}

	parent := ""
	if parentID != nil {
		parent = *parentID
	}
	e.Logf("uploaded %s", rel)
	return e.store.UpsertFile(state.FileRecord{
		RemoteID:       res.FileID,
		ParentRemoteID: parent,
		Name:           filepath.Base(rel),
		RelPath:        rel,
		SizeBytes:      res.Size,
		Checksum:       res.Checksum,
		LocalMtime:     info.ModTime().Unix(),
		LocalChecksum:  res.Checksum,
		Status:         api.FileStatusUploaded,
	})
}

// pushDeletions deletes remotely any tracked file that is no longer on disk.
func (e *Engine) pushDeletions(ctx context.Context, seen map[string]bool) (Stats, error) {
	var st Stats

	files, err := e.store.AllFiles()
	if err != nil {
		return st, err
	}
	for _, f := range files {
		if f.Deleted || seen[f.RelPath] {
			continue
		}
		if _, statErr := os.Stat(e.abs(f.RelPath)); !os.IsNotExist(statErr) {
			continue // still present (or stat error); leave it
		}
		if err := e.api.DeleteFile(ctx, f.RemoteID); err != nil {
			return st, fmt.Errorf("delete %s: %w", f.RelPath, err)
		}
		f.Deleted = true
		_ = e.store.UpsertFile(f)
		e.Logf("deleted %s", f.RelPath)
		st.Deleted++
	}
	return st, nil
}

// statAndSum returns whether a file exists, its SHA-256 hex, and its mtime.
func statAndSum(path string) (exists bool, checksum string, mtime time.Time) {
	info, err := os.Stat(path)
	if err != nil {
		return false, "", time.Time{}
	}
	sum, err := transfer.FileChecksum(path)
	if err != nil {
		return true, "", info.ModTime()
	}
	return true, sum, info.ModTime()
}
