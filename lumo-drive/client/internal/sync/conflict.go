package sync

import "time"

// resolveRemoteWins decides who wins when a path was edited both locally and
// remotely since the last sync. Last edit wins; on a tie the remote wins (the
// remote is the preferred source of truth).
//
// localMtime is the local file's modification time; remoteTime is the time of
// the remote change-log entry.
func resolveRemoteWins(localMtime, remoteTime time.Time) bool {
	return !remoteTime.Before(localMtime)
}
