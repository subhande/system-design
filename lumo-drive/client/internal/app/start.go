package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/term"

	"github.com/desubhan/system-design/lumo-drive/client/internal/config"
	"github.com/desubhan/system-design/lumo-drive/client/internal/state"
	syncpkg "github.com/desubhan/system-design/lumo-drive/client/internal/sync"
)

// Run is the client entry point: ensure the user is logged in and a sync
// directory is set, then run the background sync daemon with logging.
func Run(ctx context.Context) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	// Interactive setup (login/sync-dir) needs a terminal. Without one, fail
	// with guidance instead of looping on prompts.
	needsSetup := !cfg.LoggedIn() || cfg.SyncDir == ""
	if needsSetup && !isInteractive() {
		return errors.New("no saved login or sync directory and no terminal available; " +
			"run `lumo login` and `lumo init <dir>` first, or start with `./run.sh attach`")
	}

	// 1. Ensure we have a saved login; otherwise offer login/register.
	if !cfg.LoggedIn() {
		if err := interactiveAuth(ctx, cfg); err != nil {
			return err
		}
	} else if cfg.User != nil {
		fmt.Printf("Welcome back, %s.\n", cfg.User.Username)
	}

	// 2. Ensure a sync directory is configured.
	if cfg.SyncDir == "" {
		if err := interactiveInit(cfg); err != nil {
			return err
		}
	}

	// 3. Open local state and start the background sync daemon.
	st, syncDir, err := openStore(cfg)
	if err != nil {
		return err
	}
	defer st.Close()

	logger, closeLog, err := newLogger(syncDir)
	if err != nil {
		return err
	}
	defer closeLog()

	eng := syncpkg.New(newClient(cfg), st, syncDir)
	eng.Logf = func(format string, args ...any) { logger.Printf(format, args...) }

	logger.Printf("lumo started: user=%s dir=%s server=%s", userLabel(cfg), syncDir, cfg.APIURL)
	return eng.RunDaemon(ctx, syncpkg.DaemonOptions{})
}

// interactiveAuth presents a login/register menu until authentication succeeds.
func interactiveAuth(ctx context.Context, cfg *config.Config) error {
	fmt.Println("No existing login found.")
	for {
		fmt.Println()
		fmt.Println("  1) Login")
		fmt.Println("  2) Register")
		switch strings.TrimSpace(prompt("Choose an option [1-2]: ")) {
		case "1":
			if err := runLogin(ctx, cfg, ""); err != nil {
				fmt.Println("Login failed:", err)
				continue
			}
			return nil
		case "2":
			if err := runRegister(ctx, cfg, "", ""); err != nil {
				fmt.Println("Registration failed:", err)
				continue
			}
			return nil
		default:
			fmt.Println("Please enter 1 or 2.")
		}
	}
}

// interactiveInit prompts for and configures the sync directory.
func interactiveInit(cfg *config.Config) error {
	for {
		input := strings.TrimSpace(prompt("Enter a directory to sync: "))
		if input == "" {
			fmt.Println("A directory is required.")
			continue
		}
		absDir, err := filepath.Abs(expandHome(input))
		if err != nil {
			fmt.Println("Invalid path:", err)
			continue
		}
		if err := os.MkdirAll(absDir, 0o755); err != nil {
			fmt.Println("Could not create directory:", err)
			continue
		}
		st, err := state.Open(absDir)
		if err != nil {
			fmt.Println("Could not initialize local state:", err)
			continue
		}
		_ = st.Close()

		cfg.SyncDir = absDir
		if err := cfg.Save(); err != nil {
			return err
		}
		fmt.Printf("Sync directory set to %s\n", absDir)
		return nil
	}
}

// newLogger returns a logger that writes timestamped lines to both stdout and
// <syncDir>/.lumo/lumo.log, plus a close func for the file.
func newLogger(syncDir string) (*log.Logger, func(), error) {
	logPath := filepath.Join(syncDir, state.DirName, "lumo.log")
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, nil, err
	}
	w := io.MultiWriter(os.Stdout, f)
	return log.New(w, "", log.LstdFlags), func() { _ = f.Close() }, nil
}

// isInteractive reports whether stdin is a terminal we can prompt on.
func isInteractive() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

// expandHome expands a leading ~ to the user's home directory.
func expandHome(p string) string {
	if p == "~" || strings.HasPrefix(p, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, strings.TrimPrefix(p, "~"))
		}
	}
	return p
}

// userLabel returns a short label for the logged-in user.
func userLabel(cfg *config.Config) string {
	if cfg.User != nil {
		return cfg.User.Username
	}
	return "unknown"
}
