package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/desubhan/system-design/lumo-drive/client/internal/state"
)

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init <dir>",
		Short: "Designate a local directory as the sync root",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := requireLogin()
			if err != nil {
				return err
			}

			absDir, err := filepath.Abs(args[0])
			if err != nil {
				return err
			}
			if err := os.MkdirAll(absDir, 0o755); err != nil {
				return fmt.Errorf("create sync dir: %w", err)
			}

			// Initialize the local state database under the sync dir.
			st, err := state.Open(absDir)
			if err != nil {
				return fmt.Errorf("init local state: %w", err)
			}
			defer st.Close()

			cfg.SyncDir = absDir
			if err := cfg.Save(); err != nil {
				return err
			}

			fmt.Printf("Sync directory set to %s\n", absDir)
			fmt.Printf("Local state initialized at %s\n", state.DBPath(absDir))
			return nil
		},
	}
}
