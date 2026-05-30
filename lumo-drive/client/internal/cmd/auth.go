package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/desubhan/system-design/lumo-drive/client/internal/api"
	"github.com/desubhan/system-design/lumo-drive/client/internal/config"
)

func newRegisterCmd() *cobra.Command {
	var username, email string
	cmd := &cobra.Command{
		Use:   "register",
		Short: "Create a new account",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			return runRegister(cmd.Context(), cfg, username, email)
		},
	}
	cmd.Flags().StringVar(&username, "username", "", "account username")
	cmd.Flags().StringVar(&email, "email", "", "account email")
	return cmd
}

func newLoginCmd() *cobra.Command {
	var email string
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Log in to an existing account",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			return runLogin(cmd.Context(), cfg, email)
		},
	}
	cmd.Flags().StringVar(&email, "email", "", "account email")
	return cmd
}

// runRegister prompts for any missing fields, registers, and saves credentials.
func runRegister(ctx context.Context, cfg *config.Config, username, email string) error {
	if username == "" {
		username = prompt("Username: ")
	}
	if email == "" {
		email = prompt("Email: ")
	}
	password, err := promptPassword("Password: ")
	if err != nil {
		return err
	}

	resp, err := newClient(cfg).Register(ctx, username, email, password)
	if err != nil {
		return err
	}
	return saveAuth(cfg, resp)
}

// runLogin prompts for any missing fields, logs in, and saves credentials.
func runLogin(ctx context.Context, cfg *config.Config, email string) error {
	if email == "" {
		email = prompt("Email: ")
	}
	password, err := promptPassword("Password: ")
	if err != nil {
		return err
	}

	resp, err := newClient(cfg).Login(ctx, email, password)
	if err != nil {
		return err
	}
	return saveAuth(cfg, resp)
}

func newLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Clear the saved auth token",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			cfg.Token = ""
			cfg.User = nil
			if err := cfg.Save(); err != nil {
				return err
			}
			fmt.Println("Logged out.")
			return nil
		},
	}
}

// saveAuth persists the token and user from an auth response.
func saveAuth(cfg *config.Config, resp *api.AuthResponse) error {
	cfg.Token = resp.Token
	cfg.User = &config.User{
		UserID:   resp.User.UserID,
		Username: resp.User.Username,
		Email:    resp.User.Email,
	}
	if err := cfg.Save(); err != nil {
		return err
	}
	fmt.Printf("Logged in as %s (%s).\n", resp.User.Username, resp.User.Email)
	return nil
}

// prompt reads a single trimmed line from stdin after printing a label.
func prompt(label string) string {
	fmt.Print(label)
	r := bufio.NewReader(os.Stdin)
	line, _ := r.ReadString('\n')
	return strings.TrimSpace(line)
}

// promptPassword reads a password from the terminal without echoing it.
func promptPassword(label string) (string, error) {
	fmt.Print(label)
	fd := int(os.Stdin.Fd())
	if term.IsTerminal(fd) {
		b, err := term.ReadPassword(fd)
		fmt.Println()
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	// Non-interactive (e.g. piped input): read a plain line.
	return prompt(""), nil
}
