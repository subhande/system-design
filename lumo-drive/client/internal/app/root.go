// Package app runs the lumo interactive client: it sets up login and a sync
// directory, then runs the background sync daemon.
package app

import (
	"errors"
	"fmt"

	"github.com/desubhan/system-design/lumo-drive/client/internal/api"
	"github.com/desubhan/system-design/lumo-drive/client/internal/config"
	"github.com/desubhan/system-design/lumo-drive/client/internal/state"
)

// loadConfig loads the on-disk config.
func loadConfig() (*config.Config, error) {
	return config.Load()
}

// newClient builds an API client from config.
func newClient(cfg *config.Config) *api.Client {
	return api.New(cfg.APIURL, cfg.Token)
}

// requireSyncDir ensures a sync directory has been configured via `lumo init`.
func requireSyncDir(cfg *config.Config) (string, error) {
	if cfg.SyncDir == "" {
		return "", errors.New("no sync directory set: run `lumo init <dir>` first")
	}
	return cfg.SyncDir, nil
}

// openStore opens the local state store for the configured sync directory.
func openStore(cfg *config.Config) (*state.Store, string, error) {
	syncDir, err := requireSyncDir(cfg)
	if err != nil {
		return nil, "", err
	}
	st, err := state.Open(syncDir)
	if err != nil {
		return nil, "", fmt.Errorf("open local state: %w", err)
	}
	return st, syncDir, nil
}
