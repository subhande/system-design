#!/usr/bin/env bash
# Build and run the Lumo Drive client (lumo).
#
# Background (tmux) usage:
#   ./run.sh                # build and start the app in a detached tmux session
#   ./run.sh start          # same as above
#   ./run.sh attach         # attach to the session (to log in / watch logs)
#   ./run.sh stop           # stop the background session
#
# Foreground usage:
#   ./run.sh --dev          # run the interactive app via `go run .` (no build)
#
# On first start the app has no saved login: attach to the session
# (`./run.sh attach`) to log in or register, then detach (Ctrl-b then d) and the
# sync daemon keeps running in the background. Config is at
# ~/.config/lumo/config.json; logs also go to <sync_dir>/.lumo/lumo.log.
#
# Shutdown is graceful: `./run.sh stop` (or Ctrl+C while attached/foreground)
# sends SIGINT so the daemon flushes state and closes cleanly. A second Ctrl+C
# forces an immediate quit.

set -euo pipefail

# Always operate from the directory this script lives in.
cd "$(dirname "$0")"

export CONFIG_DIR_NAME="${CONFIG_DIR_NAME:-lumo}"
export SESSION="${SESSION:-lumo}"

echo "Using config dir name: $CONFIG_DIR_NAME"
echo "Using tmux session name: $SESSION"

# CONFIG_DIR_NAME=lumo-drive-01 SESSION=client-1 ./run.sh start
# SESSION=client-1 ./run.sh attach
# SESSION=client-1 ./run.sh stop

# CONFIG_DIR_NAME=lumo-drive-02 SESSION=client-2 ./run.sh start
# SESSION=client-2 ./run.sh attach
# SESSION=client-2 ./run.sh stop


BIN_DIR="bin"
BIN_PATH="$BIN_DIR/lumo"

build() {
    mkdir -p "$BIN_DIR"
    echo "Building client -> $BIN_PATH ..." >&2
    go build -o "$BIN_PATH" .
}

case "${1:-start}" in
--dev)
    shift
    exec go run . "$@"
    ;;

attach)
    exec tmux attach -t "$SESSION"
    ;;

stop)
    if ! tmux has-session -t "$SESSION" 2>/dev/null; then
        echo "No lumo session '$SESSION' is running."
        exit 0
    fi
    # Send Ctrl-C to the pane so the app shuts down gracefully (flush state,
    # close the DB), then wait briefly before tearing down the session.
    echo "Stopping lumo gracefully…"
    tmux send-keys -t "$SESSION" C-c
    for _ in 1 2 3 4 5 6 7 8 9 10; do
        tmux has-session -t "$SESSION" 2>/dev/null || break
        sleep 0.5
    done
    tmux kill-session -t "$SESSION" 2>/dev/null || true
    echo "Stopped lumo session '$SESSION'."
    ;;

start)
    if ! command -v tmux >/dev/null 2>&1; then
        echo "tmux is not installed. Install tmux, or run the app in the" >&2
        echo "foreground with ./run.sh --dev" >&2
        exit 1
    fi
    if tmux has-session -t "$SESSION" 2>/dev/null; then
        echo "lumo is already running in tmux session '$SESSION'."
        echo "Attach with:  ./run.sh attach"
        exit 0
    fi
    build
    # Keep the pane open after the app exits so any logs/errors stay visible.
    # Pass CONFIG_DIR_NAME explicitly so the daemon uses the right config dir
    # regardless of tmux environment inheritance.
    tmux new-session -d -s "$SESSION" "exec env CONFIG_DIR_NAME='$CONFIG_DIR_NAME' ./$BIN_PATH; echo; echo '[lumo exited]'; read -n 1 -s -r"
    echo "Started lumo in background tmux session '$SESSION'."
    echo "Attach to log in / watch logs:  ./run.sh attach"
    echo "Stop the daemon:                ./run.sh stop"
    ;;

*)
    echo "Unknown command: $1" >&2
    echo "Usage: ./run.sh [start|attach|stop|--dev]" >&2
    exit 1
    ;;
esac