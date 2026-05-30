package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newPullCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "pull",
		Short: "Download remote changes into the local sync directory",
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

			stats, err := newEngine(cfg, st, syncDir).Pull(cmd.Context())
			if err != nil {
				return err
			}
			fmt.Printf("pull complete: %d downloaded, %d removed, %d conflicts\n",
				stats.Downloaded, stats.Removed, stats.Conflicts)
			return nil
		},
	}
}
