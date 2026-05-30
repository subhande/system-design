package sync

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/desubhan/system-design/lumo-drive/client/internal/api"
	"github.com/desubhan/system-design/lumo-drive/client/internal/state"
	"github.com/desubhan/system-design/lumo-drive/client/internal/transfer"
)

// Pull fetches remote changes since the stored cursor and applies them to the
// local tree, advancing the cursor. Remote is the source of truth; when a path
// was also modified locally, the conflict is resolved by last-edit-wins (remote
// wins on a tie).
func (e *Engine) Pull(ctx context.Context) (Stats, error) {
	var total Stats

	since, err := e.store.LastChangeID()
	if err != nil {
		return total, err
	}

	changes, err := e.api.GetChanges(ctx, since)
	if err != nil {
		return total, fmt.Errorf("get changes: %w", err)
	}

	maxID := since
	for _, ch := range changes {
		if ch.ChangeID > maxID {
			maxID = ch.ChangeID
		}

		switch ch.EntityType {
		case api.EntityTypeFolder:
			if err := e.applyFolderChange(ctx, ch); err != nil {
				return total, err
			}
		case api.EntityTypeFile:
			st, err := e.applyFileChange(ctx, ch)
			if err != nil {
				return total, err
			}
			total.Add(st)
		}
	}

	if err := e.store.SetLastChangeID(maxID); err != nil {
		return total, err
	}
	return total, nil
}

func (e *Engine) applyFolderChange(ctx context.Context, ch api.Change) error {
	if ch.Action == api.ActionDeleted {
		if f, err := e.store.FolderByID(ch.EntityID); err == nil && f != nil {
			f.Deleted = true
			_ = e.store.UpsertFolder(*f)
		}
		return nil
	}
	_, err := e.folderRelPath(ctx, ch.EntityID)
	return err
}

func (e *Engine) applyFileChange(ctx context.Context, ch api.Change) (Stats, error) {
	var st Stats

	if ch.Action == api.ActionDeleted {
		return e.applyRemoteDelete(ch)
	}

	meta, err := e.api.GetFileMetadata(ctx, ch.EntityID)
	if err != nil {
		return st, fmt.Errorf("get file %s: %w", ch.EntityID, err)
	}
	if meta.Status != api.FileStatusUploaded {
		return st, nil // not complete yet; a later change covers it
	}

	rel, err := e.fileRelPath(ctx, meta)
	if err != nil {
		return st, err
	}
	dest := e.abs(rel)

	// Rename/move with unchanged content: move the local file in place.
	if existing, _ := e.store.FileByID(meta.FileID); existing != nil &&
		existing.RelPath != rel && existing.Checksum == meta.Checksum {
		oldPath := e.abs(existing.RelPath)
		if exists, sum, _ := statAndSum(oldPath); exists && sum == meta.Checksum {
			if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
				return st, err
			}
			if err := os.Rename(oldPath, dest); err == nil {
				e.Logf("moved %s -> %s", existing.RelPath, rel)
				return st, e.recordSyncedFile(dest, rel, meta)
			}
		}
	}

	return e.applyRemoteContent(ctx, meta, rel, dest, ch.CreatedAt)
}

// applyRemoteContent downloads/overwrites the local file with the remote one,
// applying the conflict policy when the local copy was independently modified.
func (e *Engine) applyRemoteContent(ctx context.Context, meta *api.FileMetadata, rel, dest string, remoteTime time.Time) (Stats, error) {
	var st Stats

	exists, localSum, localMtime := statAndSum(dest)
	snap, err := e.store.FileByPath(rel)
	if err != nil {
		return st, err
	}

	if !exists {
		// New remote file, or one that was locally deleted; remote is truth.
		if err := e.download(ctx, meta, rel, dest); err != nil {
			return st, err
		}
		st.Downloaded++
		return st, nil
	}

	localModified := snap == nil || localSum != snap.LocalChecksum

	if !localModified {
		// Local matches last sync; apply remote if it differs.
		if localSum == meta.Checksum {
			st.Skipped++
			return st, e.recordSyncedFile(dest, rel, meta)
		}
		if err := e.download(ctx, meta, rel, dest); err != nil {
			return st, err
		}
		st.Downloaded++
		return st, nil
	}

	// Local was modified independently -> conflict (unless content is identical).
	if localSum == meta.Checksum {
		return st, e.recordSyncedFile(dest, rel, meta)
	}

	st.Conflicts++
	if resolveRemoteWins(localMtime, remoteTime) {
		e.Logf("conflict on %s: remote wins (last edit)", rel)
		if err := e.download(ctx, meta, rel, dest); err != nil {
			return st, err
		}
		st.Downloaded++
		return st, nil
	}

	// Local wins: leave the file and snapshot untouched so Push re-uploads it.
	e.Logf("conflict on %s: local wins (last edit); will upload", rel)
	return st, nil
}

// applyRemoteDelete applies a remote deletion to the local tree, keeping the
// local copy only if it was edited more recently than the remote deletion.
func (e *Engine) applyRemoteDelete(ch api.Change) (Stats, error) {
	var st Stats

	f, err := e.store.FileByID(ch.EntityID)
	if err != nil || f == nil {
		return st, nil
	}
	dest := e.abs(f.RelPath)

	if exists, localSum, localMtime := statAndSum(dest); exists {
		localModified := localSum != f.LocalChecksum
		if localModified && localMtime.After(ch.CreatedAt) {
			// Local edited after the remote delete -> local wins; keep & re-upload.
			e.Logf("conflict on %s: local edit newer than remote delete; keeping", f.RelPath)
			st.Conflicts++
			return st, nil
		}
		_ = os.Remove(dest)
		st.Removed++
	}

	f.Deleted = true
	_ = e.store.UpsertFile(*f)
	return st, nil
}

// download fetches the remote file to dest and records it as synced.
func (e *Engine) download(ctx context.Context, meta *api.FileMetadata, rel, dest string) error {
	if err := transfer.DownloadFile(ctx, e.api, meta.FileID, dest, meta.Checksum); err != nil {
		return fmt.Errorf("download %s: %w", rel, err)
	}
	e.Logf("downloaded %s", rel)
	return e.recordSyncedFile(dest, rel, meta)
}

// fileRelPath computes a file's relative path from its remote metadata.
func (e *Engine) fileRelPath(ctx context.Context, meta *api.FileMetadata) (string, error) {
	parentRel := ""
	if meta.ParentFolderID != nil {
		var err error
		parentRel, err = e.folderRelPath(ctx, *meta.ParentFolderID)
		if err != nil {
			return "", err
		}
	}
	return joinRel(parentRel, meta.FileName), nil
}

// recordSyncedFile updates the local snapshot to match the just-synced remote
// file (local content now equals the remote checksum).
func (e *Engine) recordSyncedFile(dest, rel string, meta *api.FileMetadata) error {
	info, err := os.Stat(dest)
	if err != nil {
		return err
	}
	parent := ""
	if meta.ParentFolderID != nil {
		parent = *meta.ParentFolderID
	}
	return e.store.UpsertFile(state.FileRecord{
		RemoteID:       meta.FileID,
		ParentRemoteID: parent,
		Name:           meta.FileName,
		RelPath:        rel,
		SizeBytes:      info.Size(),
		Checksum:       meta.Checksum,
		LocalMtime:     info.ModTime().Unix(),
		LocalChecksum:  meta.Checksum,
		Status:         meta.Status,
	})
}
