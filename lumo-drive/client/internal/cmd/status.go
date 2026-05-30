package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show client configuration and sync state",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}

			fmt.Printf("Client ID: %s\n", cfg.ClientID)
			fmt.Printf("Server:    %s\n", cfg.APIURL)
			if cfg.LoggedIn() && cfg.User != nil {
				fmt.Printf("User:      %s (%s)\n", cfg.User.Username, cfg.User.Email)
			} else {
				fmt.Println("User:      (not logged in)")
			}

			if cfg.SyncDir == "" {
				fmt.Println("Sync dir:  (not set — run `lumo init <dir>`)")
				return nil
			}
			fmt.Printf("Sync dir:  %s\n", cfg.SyncDir)

			// Reachability probe.
			if err := newClient(cfg).Health(cmd.Context()); err != nil {
				fmt.Printf("Server:    UNREACHABLE (%v)\n", err)
			} else {
				fmt.Println("Server:    reachable")
			}

			st, _, err := openStore(cfg)
			if err != nil {
				return err
			}
			defer st.Close()

			lastChange, err := st.LastChangeID()
			if err != nil {
				return err
			}
			files, err := st.AllFiles()
			if err != nil {
				return err
			}
			fmt.Printf("Tracked files: %d\n", len(files))
			fmt.Printf("Last change id: %d\n", lastChange)
			return nil
		},
	}
}
