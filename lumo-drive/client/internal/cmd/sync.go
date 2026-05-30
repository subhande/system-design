package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newSyncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Reconcile local and remote in one pass (pull then push)",
		Long: "sync pulls remote changes first (remote is the source of truth) and " +
			"then pushes outstanding local changes. On a file edited both locally and " +
			"remotely, the most recent edit wins (remote wins on a tie).",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := requireLogin()
			if err != nil {
				return err
			}
			st, syncDir, err := openStore(cfg)
			if err != nil {
				return err
			}
			defer st.Close()

			stats, err := newEngine(cfg, st, syncDir).Sync(cmd.Context())
			if err != nil {
				return err
			}
			fmt.Printf("sync complete: %d downloaded, %d uploaded, %d removed, %d deleted, %d conflicts, %d unchanged\n",
				stats.Downloaded, stats.Uploaded, stats.Removed, stats.Deleted, stats.Conflicts, stats.Skipped)
			return nil
		},
	}
}
