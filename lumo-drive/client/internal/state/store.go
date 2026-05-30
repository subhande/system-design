// Package state persists the client's local sync state in a SQLite database
// stored under the sync directory (.lumo/state.db). It records the mapping
// between local paths and remote ids plus the last-synced snapshot used for
// fast change detection and resumable uploads.
package state

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	_ "modernc.org/sqlite"
)

// DirName is the per-sync-dir metadata directory.
const DirName = ".lumo"

// FileRecord is the last-synced snapshot of a file.
type FileRecord struct {
	RemoteID       string
	ParentRemoteID string
	Name           string
	RelPath        string
	SizeBytes      int64
	Checksum       string // last-synced remote checksum
	LocalMtime     int64  // unix seconds of the local file at last sync
	LocalChecksum  string // local content checksum at last sync
	Status         string
	Deleted        bool
}

// FolderRecord is the last-synced snapshot of a folder.
type FolderRecord struct {
	RemoteID       string
	ParentRemoteID string
	Name           string
	RelPath        string
	Deleted        bool
}

// Store wraps the SQLite database.
type Store struct {
	db *sql.DB
}

// DBPath returns the state database path for a given sync directory.
func DBPath(syncDir string) string {
	return filepath.Join(syncDir, DirName, "state.db")
}

// Open opens (creating if needed) the state database for the sync directory and
// applies the schema.
func Open(syncDir string) (*Store, error) {
	dir := filepath.Join(syncDir, DirName)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", DBPath(syncDir))
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("apply schema: %w", err)
	}
	return &Store{db: db}, nil
}

// Close closes the underlying database.
func (s *Store) Close() error { return s.db.Close() }

// --- meta -----------------------------------------------------------------

// GetMeta returns a meta value, or "" if absent.
func (s *Store) GetMeta(key string) (string, error) {
	var v string
	err := s.db.QueryRow("SELECT value FROM meta WHERE key = ?", key).Scan(&v)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return v, err
}

// SetMeta upserts a meta value.
func (s *Store) SetMeta(key, value string) error {
	_, err := s.db.Exec(
		"INSERT INTO meta (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value",
		key, value)
	return err
}

// LastChangeID returns the highest sync change_id applied locally (0 if none).
func (s *Store) LastChangeID() (int64, error) {
	v, err := s.GetMeta("last_change_id")
	if err != nil || v == "" {
		return 0, err
	}
	return strconv.ParseInt(v, 10, 64)
}

// SetLastChangeID records the highest sync change_id applied locally.
func (s *Store) SetLastChangeID(id int64) error {
	return s.SetMeta("last_change_id", strconv.FormatInt(id, 10))
}

// --- folders --------------------------------------------------------------

// UpsertFolder inserts or updates a folder record. Any other folder previously
// mapped to the same rel_path is removed first, since a path maps to at most
// one remote folder.
func (s *Store) UpsertFolder(f FolderRecord) error {
	if _, err := s.db.Exec("DELETE FROM folders WHERE rel_path = ? AND remote_id != ?", f.RelPath, f.RemoteID); err != nil {
		return err
	}
	_, err := s.db.Exec(`
		INSERT INTO folders (remote_id, parent_remote_id, name, rel_path, deleted)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(remote_id) DO UPDATE SET
			parent_remote_id = excluded.parent_remote_id,
			name             = excluded.name,
			rel_path         = excluded.rel_path,
			deleted          = excluded.deleted`,
		f.RemoteID, f.ParentRemoteID, f.Name, f.RelPath, boolToInt(f.Deleted))
	return err
}

// FolderByPath returns the folder mapped to a relative path, or nil if none.
func (s *Store) FolderByPath(relPath string) (*FolderRecord, error) {
	row := s.db.QueryRow(
		"SELECT remote_id, parent_remote_id, name, rel_path, deleted FROM folders WHERE rel_path = ?", relPath)
	return scanFolder(row)
}

// FolderByID returns the folder with a remote id, or nil if none.
func (s *Store) FolderByID(remoteID string) (*FolderRecord, error) {
	row := s.db.QueryRow(
		"SELECT remote_id, parent_remote_id, name, rel_path, deleted FROM folders WHERE remote_id = ?", remoteID)
	return scanFolder(row)
}

// AllFolders returns every non-deleted folder record.
func (s *Store) AllFolders() ([]FolderRecord, error) {
	rows, err := s.db.Query(
		"SELECT remote_id, parent_remote_id, name, rel_path, deleted FROM folders WHERE deleted = 0")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []FolderRecord
	for rows.Next() {
		var f FolderRecord
		var deleted int
		if err := rows.Scan(&f.RemoteID, &f.ParentRemoteID, &f.Name, &f.RelPath, &deleted); err != nil {
			return nil, err
		}
		f.Deleted = deleted != 0
		out = append(out, f)
	}
	return out, rows.Err()
}

// --- files ----------------------------------------------------------------

// UpsertFile inserts or updates a file record. Any other file previously mapped
// to the same rel_path is removed first, since a path maps to at most one
// remote file (re-uploading content produces a new remote_id for the path).
func (s *Store) UpsertFile(f FileRecord) error {
	if _, err := s.db.Exec("DELETE FROM files WHERE rel_path = ? AND remote_id != ?", f.RelPath, f.RemoteID); err != nil {
		return err
	}
	_, err := s.db.Exec(`
		INSERT INTO files (remote_id, parent_remote_id, name, rel_path, size_bytes, checksum, local_mtime, local_checksum, status, deleted)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(remote_id) DO UPDATE SET
			parent_remote_id = excluded.parent_remote_id,
			name             = excluded.name,
			rel_path         = excluded.rel_path,
			size_bytes       = excluded.size_bytes,
			checksum         = excluded.checksum,
			local_mtime      = excluded.local_mtime,
			local_checksum   = excluded.local_checksum,
			status           = excluded.status,
			deleted          = excluded.deleted`,
		f.RemoteID, f.ParentRemoteID, f.Name, f.RelPath, f.SizeBytes, f.Checksum,
		f.LocalMtime, f.LocalChecksum, f.Status, boolToInt(f.Deleted))
	return err
}

// FileByPath returns the file mapped to a relative path, or nil if none.
func (s *Store) FileByPath(relPath string) (*FileRecord, error) {
	row := s.db.QueryRow(
		"SELECT remote_id, parent_remote_id, name, rel_path, size_bytes, checksum, local_mtime, local_checksum, status, deleted FROM files WHERE rel_path = ?", relPath)
	return scanFile(row)
}

// FileByID returns the file with a remote id, or nil if none.
func (s *Store) FileByID(remoteID string) (*FileRecord, error) {
	row := s.db.QueryRow(
		"SELECT remote_id, parent_remote_id, name, rel_path, size_bytes, checksum, local_mtime, local_checksum, status, deleted FROM files WHERE remote_id = ?", remoteID)
	return scanFile(row)
}

// AllFiles returns every non-deleted file record.
func (s *Store) AllFiles() ([]FileRecord, error) {
	rows, err := s.db.Query(
		"SELECT remote_id, parent_remote_id, name, rel_path, size_bytes, checksum, local_mtime, local_checksum, status, deleted FROM files WHERE deleted = 0")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []FileRecord
	for rows.Next() {
		var f FileRecord
		var deleted int
		if err := rows.Scan(&f.RemoteID, &f.ParentRemoteID, &f.Name, &f.RelPath, &f.SizeBytes,
			&f.Checksum, &f.LocalMtime, &f.LocalChecksum, &f.Status, &deleted); err != nil {
			return nil, err
		}
		f.Deleted = deleted != 0
		out = append(out, f)
	}
	return out, rows.Err()
}

// --- pending uploads (resumable) ------------------------------------------

// SetChunkUploaded records a chunk's ETag and marks it done in the journal.
func (s *Store) SetChunkUploaded(remoteID, relPath string, chunkIndex int, etag string) error {
	_, err := s.db.Exec(`
		INSERT INTO pending_uploads (remote_id, rel_path, chunk_index, etag, done)
		VALUES (?, ?, ?, ?, 1)
		ON CONFLICT(remote_id, chunk_index) DO UPDATE SET etag = excluded.etag, done = 1`,
		remoteID, relPath, chunkIndex, etag)
	return err
}

// ClearPendingUploads removes the upload journal for a file once complete.
func (s *Store) ClearPendingUploads(remoteID string) error {
	_, err := s.db.Exec("DELETE FROM pending_uploads WHERE remote_id = ?", remoteID)
	return err
}

// --- helpers ---------------------------------------------------------------

func scanFolder(row *sql.Row) (*FolderRecord, error) {
	var f FolderRecord
	var deleted int
	err := row.Scan(&f.RemoteID, &f.ParentRemoteID, &f.Name, &f.RelPath, &deleted)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	f.Deleted = deleted != 0
	return &f, nil
}

func scanFile(row *sql.Row) (*FileRecord, error) {
	var f FileRecord
	var deleted int
	err := row.Scan(&f.RemoteID, &f.ParentRemoteID, &f.Name, &f.RelPath, &f.SizeBytes,
		&f.Checksum, &f.LocalMtime, &f.LocalChecksum, &f.Status, &deleted)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	f.Deleted = deleted != 0
	return &f, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
