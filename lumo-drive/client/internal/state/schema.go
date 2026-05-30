package state

// schema is the local SQLite schema. It maps local filesystem paths to remote
// ids and records the last-synced snapshot of each entity so we can do fast
// change detection and resumable uploads.
const schema = `
CREATE TABLE IF NOT EXISTS meta (
    key   TEXT PRIMARY KEY,
    value TEXT
);

CREATE TABLE IF NOT EXISTS folders (
    remote_id        TEXT PRIMARY KEY,
    parent_remote_id TEXT,
    name             TEXT NOT NULL,
    rel_path         TEXT NOT NULL UNIQUE,
    deleted          INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS files (
    remote_id        TEXT PRIMARY KEY,
    parent_remote_id TEXT,
    name             TEXT NOT NULL,
    rel_path         TEXT NOT NULL UNIQUE,
    size_bytes       INTEGER NOT NULL DEFAULT 0,
    checksum         TEXT,
    local_mtime      INTEGER NOT NULL DEFAULT 0,
    local_checksum   TEXT,
    status           TEXT,
    deleted          INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS pending_uploads (
    remote_id   TEXT NOT NULL,
    rel_path    TEXT NOT NULL,
    chunk_index INTEGER NOT NULL,
    etag        TEXT,
    done        INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (remote_id, chunk_index)
);
`
