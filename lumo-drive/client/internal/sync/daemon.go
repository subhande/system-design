package sync

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	stdsync "sync"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/desubhan/system-design/lumo-drive/client/internal/state"
)

// DaemonOptions configures the background sync loop.
type DaemonOptions struct {
	// PollInterval is how often to poll the server for remote changes.
	PollInterval time.Duration
	// Debounce coalesces a burst of filesystem events into one sync.
	Debounce time.Duration
}

// withDefaults fills in sensible defaults.
func (o DaemonOptions) withDefaults() DaemonOptions {
	if o.PollInterval <= 0 {
		o.PollInterval = 15 * time.Second
	}
	if o.Debounce <= 0 {
		o.Debounce = 2 * time.Second
	}
	return o
}

// RunDaemon watches the sync directory and periodically polls the server,
// running a full Sync on local changes (debounced) and on each poll tick. Sync
// runs are serialized. When the server is unreachable the loop logs and keeps
// running, so local edits are simply queued until connectivity returns
// (offline access). It blocks until ctx is cancelled.
func (e *Engine) RunDaemon(ctx context.Context, opts DaemonOptions) error {
	opts = opts.withDefaults()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	e.addWatches(watcher)

	var mu stdsync.Mutex
	runSync := func(reason string) {
		mu.Lock()
		defer mu.Unlock()

		if err := e.api.Health(ctx); err != nil {
			e.Logf("offline (%v); deferring sync, local changes are queued", err)
			return
		}
		stats, err := e.Sync(ctx)
		if err != nil {
			e.Logf("sync error (%s): %v", reason, err)
			return
		}
		e.Logf("[%s] %d downloaded, %d uploaded, %d removed, %d deleted, %d conflicts",
			reason, stats.Downloaded, stats.Uploaded, stats.Removed, stats.Deleted, stats.Conflicts)
		// New local directories may have appeared; watch them too.
		e.addWatches(watcher)
	}

	runSync("startup")

	ticker := time.NewTicker(opts.PollInterval)
	defer ticker.Stop()

	debounce := make(chan struct{}, 1)
	var timer *time.Timer

	e.Logf("daemon watching %s (poll every %s)", e.syncDir, opts.PollInterval)
	for {
		select {
		case <-ctx.Done():
			e.Logf("shutdown requested; daemon stopped cleanly")
			return nil

		case <-ticker.C:
			runSync("poll")

		case <-debounce:
			runSync("local change")

		case ev, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			// Ignore our own metadata directory and hidden entries (.DS_Store, etc.).
			if rel, relErr := filepath.Rel(e.syncDir, ev.Name); relErr == nil && isHidden(rel) {
				continue
			}
			// If a new directory was created, start watching it immediately.
			if ev.Op&fsnotify.Create != 0 {
				if info, statErr := os.Stat(ev.Name); statErr == nil && info.IsDir() {
					_ = watcher.Add(ev.Name)
				}
			}
			// Debounce: (re)arm a one-shot timer that fires a single sync.
			if timer != nil {
				timer.Stop()
			}
			timer = time.AfterFunc(opts.Debounce, func() {
				select {
				case debounce <- struct{}{}:
				default:
				}
			})

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			e.Logf("watch error: %v", err)
		}
	}
}

// addWatches adds a recursive watch on the sync directory tree, skipping the
// metadata directory. Adding an already-watched path is a no-op.
func (e *Engine) addWatches(w *fsnotify.Watcher) {
	_ = filepath.WalkDir(e.syncDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			rel, relErr := filepath.Rel(e.syncDir, path)
			if d.Name() == state.DirName || (relErr == nil && isHidden(rel)) {
				return fs.SkipDir
			}
			_ = w.Add(path)
		}
		return nil
	})
}
