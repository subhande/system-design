#!/usr/bin/env bash
# Build and run the Lumo Drive server.
#
# Usage:
#   ./run.sh          # build to ./bin/lumo-server and run it
#   ./run.sh --dev    # run with `go run .` (no separate build step)
#
# Environment is loaded by the server itself from .env / ../.env, so a
# DATABASE_URL (and AWS_* / BUCKET_NAME for file ops) should be set there.

set -euo pipefail

# Always operate from the directory this script lives in.
cd "$(dirname "$0")"

if [[ "${1:-}" == "--dev" ]]; then
    echo "Starting server with 'go run'..."
    exec go run .
fi

BIN_DIR="bin"
BIN_PATH="$BIN_DIR/lumo-server"

mkdir -p "$BIN_DIR"

echo "Building server -> $BIN_PATH ..."
go build -o "$BIN_PATH" .

echo "Starting server..."
exec "./$BIN_PATH"
