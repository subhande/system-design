package cmd

import (
	"time"

	"github.com/spf13/cobra"

	syncpkg "github.com/desubhan/system-design/lumo-drive/client/internal/sync"
)

func newDaemonCmd() *cobra.Command {
	var pollInterval, debounce time.Duration
	cmd := &cobra.Command{
		Use:   "daemon",
		Short: "Continuously sync: watch the directory and poll the server",
		Long: "daemon runs until interrupted. It watches the sync directory for local " +
			"changes (debounced) and polls the server for remote changes, running a full " +
			"sync on each. While the server is unreachable it keeps running and queues " +
			"local edits for the next successful sync.",
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

			return newEngine(cfg, st, syncDir).RunDaemon(cmd.Context(), syncpkg.DaemonOptions{
				PollInterval: pollInterval,
				Debounce:     debounce,
			})
		},
	}
	cmd.Flags().DurationVar(&pollInterval, "poll", 15*time.Second, "how often to poll the server for remote changes")
	cmd.Flags().DurationVar(&debounce, "debounce", 2*time.Second, "coalesce local filesystem events within this window")
	return cmd
}
