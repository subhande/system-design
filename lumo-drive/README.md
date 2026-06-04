# Lumo Drive

A Dropbox-like file storage and synchronization service. It has two parts:

- **Server** (`server/`) — a Go HTTP API backed by PostgreSQL and S3 that stores
  file metadata, brokers chunked (multipart) uploads and presigned downloads, and
  exposes a cursor-based change feed for syncing.
- **Client** (`client/`) — a Go app (`lumo`) that maps a local directory to the
  user's remote drive and keeps them in sync, with resumable transfers, conflict
  resolution, offline tolerance, and a background daemon.

> Design details and diagrams: see [ARCHITECTURE.md](ARCHITECTURE.md).

## Features

- User registration / login (JWT auth, bcrypt password hashing).
- Folders with create / rename / move / soft-delete (to a per-user **Trash**).
- Files up to 50 GB via **S3 multipart upload** in fixed 5 MiB chunks, with
  presigned per-chunk upload URLs and presigned download URLs.
- **Resumable** uploads (server tracks pending chunks) and downloads (HTTP Range).
- **Sync engine**: cursor-based remote change feed + local filesystem mapping,
  with conflict resolution (**remote is source of truth; last edit wins**).
- **Offline access** and a **background daemon** (filesystem watch + polling).
- **Hidden entries excluded**: dotfiles and dot-directories (e.g. `.DS_Store`,
  `.git`, the client's own `.lumo/`) are never synced.
- **Multiple independent clients** on one machine via `CONFIG_DIR_NAME`
  (separate config dir → separate account / sync directory).

Scope notes: single-user drives (no file sharing), web UI not included.

## Repository layout

```
lumo-drive/
├── server/                 # Go + Fiber API (PostgreSQL + S3)
│   ├── main.go             # routes, middleware, server bootstrap
│   ├── auth.go             # register / login, JWT issuing
│   ├── folder.go           # folder CRUD + Trash
│   ├── files.go            # multipart upload, presign, download, status
│   ├── sync.go             # change-log writes + change feed
│   ├── models.go           # DB models / DTOs
│   ├── db.go middleware.go utils.go
│   ├── init.sql            # schema (applied on boot)
│   └── run.sh
├── client/                 # Go client app (lumo)
│   ├── main.go             # entrypoint + graceful shutdown
│   ├── internal/
│   │   ├── api/            # typed HTTP client (one method per endpoint)
│   │   ├── config/         # <os-config-dir>/<CONFIG_DIR_NAME>/config.json
│   │   ├── state/          # SQLite local state (path<->id, snapshots, journal)
│   │   ├── transfer/       # chunking, upload, download
│   │   ├── sync/           # reconcile engine, conflict policy, daemon
│   │   └── app/            # interactive app entry point
│   └── run.sh              # build + run under tmux
├── postgres.sh             # spin up a local Postgres in Docker
└── go.mod                  # single module: .../lumo-drive
```

Both binaries live in one Go module: `github.com/desubhan/system-design/lumo-drive`.

## Prerequisites

- Go 1.25+
- Docker (for local Postgres) or an existing PostgreSQL instance
- An S3 bucket + AWS credentials (for real file transfers)
- `tmux` (optional — used by `client/run.sh` to run the client in the background)

## Quick start

### 1. Database

```sh
cd lumo-drive
./postgres.sh           # runs Postgres 18 in Docker on :5432 (db: lumo_drive_db)
```

The server applies `server/init.sql` automatically on startup.

### 2. Server

Configure `server/.env` (or `lumo-drive/.env`):

| Variable        | Purpose                                   | Default                        |
|-----------------|-------------------------------------------|--------------------------------|
| `DATABASE_URL`  | Postgres connection string (required)     | —                              |
| `JWT_SECRET`    | HMAC signing key for JWTs                 | `dev-insecure-secret-change-me`|
| `BUCKET_NAME`   | S3 bucket for object storage              | —                              |
| `BODY_LIMIT_MB` | Max request body   | `2`                            |
| `AWS_ACCESS_KEY_ID`     | AWS access key ID | from environment            |
| `AWS_SECRET_ACCESS_KEY` | AWS secret access key | from environment            |
| `AWS_REGION`            | AWS region | from environment            |

```sh
cd lumo-drive/server
./run.sh                # builds to bin/lumo-server and listens on :6000
# health check:
curl localhost:6000/health   # -> {"status":"ok"}
```

### 3. Client

```sh
cd lumo-drive/client
./run.sh                # build + start the app in a background tmux session
./run.sh attach         # attach to log in / register, then Ctrl-b d to detach
./run.sh stop           # graceful shutdown
```

On first run the app shows a **Login / Register** menu, saves your token, asks
for a sync directory, then runs the sync daemon (filesystem watch + polling).
Attach to complete first-time login, then detach (`Ctrl-b` then `d`) and the
daemon keeps running. Logs stream to the tmux pane and to
`<sync_dir>/.lumo/lumo.log`.

`run.sh` verbs: `start` (default), `attach`, `stop`, and `--dev` (runs the
interactive app via `go run` without a separate build). Running the binary
directly (`./bin/lumo`) does the same interactive flow in the foreground.

**Shutdown is graceful:** `./run.sh stop` (or `Ctrl+C` while attached/foreground)
sends `SIGINT` so the daemon flushes state and closes the DB; a second `Ctrl+C`
forces an immediate quit.

#### Running multiple clients

To run several independent clients on one machine, set `CONFIG_DIR_NAME` (a
separate config directory → separate account, token, and sync dir) and `SESSION`
(a separate tmux session) per client:

```sh
CONFIG_DIR_NAME=lumo-drive-02 SESSION=client-2 ./run.sh start
SESSION=client-2 ./run.sh attach
SESSION=client-2 ./run.sh stop
```

`attach`/`stop` only need `SESSION`. Give each client a **different sync
directory** — two daemons must not share one `.lumo/state.db`.

#### Client app

`lumo` takes no subcommands — it's a single interactive app. On start it:

1. **Authenticates** — if there's no saved login, it shows a **Login / Register**
   menu (prompts for username/email/password) and stores the JWT in the config.
2. **Sets a sync directory** — if none is configured, it prompts for one and
   creates the local state database (`.lumo/state.db`) under it.
3. **Runs the sync daemon** — watches the directory for local changes (debounced)
   and polls the server for remote changes, reconciling both ways continuously
   until interrupted.

#### Client config & local state

- **Config:** `<os-config-dir>/<CONFIG_DIR_NAME>/config.json` (e.g.
  `~/.config/lumo/config.json` on Linux, `~/Library/Application
  Support/lumo/config.json` on macOS). The directory name defaults to `lumo` and
  is overridable with `CONFIG_DIR_NAME`; override the full path with
  `LUMO_CONFIG`. Holds a generated **client id** (a UUID identifying the install,
  created on first run), the server URL (default `http://localhost:6000`), auth
  token, sync directory, and user. Written `0600`.
- **Local state:** `<sync_dir>/.lumo/state.db` (SQLite) — maps local paths to
  remote ids, stores last-synced snapshots for fast change detection, and
  journals uploaded chunks for resumable uploads.
- **Excluded from sync:** any path with a dot-prefixed segment — dotfiles and
  dot-directories such as `.DS_Store`, `.git`, and the `.lumo/` metadata dir —
  is never uploaded, watched, or deleted remotely.

The sync/conflict model (remote source of truth, last-edit-wins) and the
resumable/offline behavior are described in [ARCHITECTURE.md](ARCHITECTURE.md).

## API summary

Base URL `http://localhost:6000/api/v1`; authenticated routes require
`Authorization: Bearer <token>`.

| Method & path                                   | Description                          |
|-------------------------------------------------|--------------------------------------|
| `POST /auth/register`, `POST /auth/login`       | Auth → `{token, user}`               |
| `POST /folders/`                                | Create folder                        |
| `GET /folders/:id`                              | Folder metadata (name, parent)       |
| `GET /folders/:id/contents`                     | List subfolders + files              |
| `PATCH /folders/:id`, `POST /folders/:id/move`, `DELETE /folders/:id` | Rename / move / soft-delete |
| `POST /files/`                                  | Initiate upload (returns file id + chunks) |
| `GET /files/:id/chunks/:idx/presign`            | Presigned S3 UploadPart URL          |
| `PATCH /files/:id/chunks/:idx/status/:status`   | Mark chunk (etag via `?etag=`)       |
| `PATCH /files/:id/status/:status`              | Mark file (`uploaded` completes multipart) |
| `POST /files/:id/resume`                        | Pending chunk indices                |
| `GET /files/:id`, `GET /files/:id/download`     | Metadata / presigned download URL    |
| `PATCH /files/:id/rename`, `POST /files/:id/move`, `DELETE /files/:id` | Rename / move / soft-delete |
| `GET /sync/changes/:change_id`                  | Changes with `change_id >` cursor    |
| `GET /health`                                   | Liveness                             |

## Testing

```sh
cd lumo-drive
go build ./...
go vet ./client/... ./server/
go test ./client/...
```
