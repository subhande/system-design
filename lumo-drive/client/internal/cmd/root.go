// Package cmd defines the lumo CLI commands.
package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/desubhan/system-design/lumo-drive/client/internal/api"
	"github.com/desubhan/system-design/lumo-drive/client/internal/config"
	"github.com/desubhan/system-design/lumo-drive/client/internal/state"
	syncpkg "github.com/desubhan/system-design/lumo-drive/client/internal/sync"
)

// NewRootCmd builds the root cobra command with all subcommands attached.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "lumo",
		Short: "Lumo Drive sync client",
		Long: "lumo is a client for Lumo Drive. Run it with no arguments to start the " +
			"interactive app: it logs you in (or registers), then runs the background " +
			"sync daemon. Individual subcommands are also available for scripting.",
		Args:          cobra.NoArgs,
		RunE:          runApp,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.AddCommand(
		newRegisterCmd(),
		newLoginCmd(),
		newLogoutCmd(),
		newInitCmd(),
		newStatusCmd(),
		newPushCmd(),
		newPullCmd(),
		newSyncCmd(),
		newDaemonCmd(),
	)
	return root
}

// newEngine builds a sync Engine wired to log progress to stdout.
func newEngine(cfg *config.Config, st *state.Store, syncDir string) *syncpkg.Engine {
	eng := syncpkg.New(newClient(cfg), st, syncDir)
	eng.Logf = func(format string, args ...any) { fmt.Printf(format+"\n", args...) }
	return eng
}

// loadConfig loads the on-disk config.
func loadConfig() (*config.Config, error) {
	return config.Load()
}

// requireLogin loads config and ensures the user is authenticated.
func requireLogin() (*config.Config, error) {
	cfg, err := loadConfig()
	if err != nil {
		return nil, err
	}
	if !cfg.LoggedIn() {
		return nil, errors.New("not logged in: run `lumo login` first")
	}
	return cfg, nil
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
