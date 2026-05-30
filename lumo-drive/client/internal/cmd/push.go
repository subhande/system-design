package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newPushCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "push",
		Short: "Upload new and changed local files to the server",
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

			stats, err := newEngine(cfg, st, syncDir).Push(cmd.Context())
			if err != nil {
				return err
			}
			fmt.Printf("push complete: %d uploaded, %d deleted, %d unchanged\n",
				stats.Uploaded, stats.Deleted, stats.Skipped)
			return nil
		},
	}
}
