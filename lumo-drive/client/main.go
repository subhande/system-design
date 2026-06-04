// Command lumo is the Lumo Drive sync client.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"github.com/desubhan/system-design/lumo-drive/client/internal/app"
)

func main() {
	// Load environment from a local or parent .env, if present. Missing files
	// are ignored so the client still runs from defaults.
	_ = godotenv.Load(".env", "../.env")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Graceful shutdown: the first SIGINT/SIGTERM cancels the context so the
	// daemon can stop cleanly (flush state, close the DB); a second signal
	// forces an immediate exit in case shutdown hangs.
	sigCh := make(chan os.Signal, 2)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Fprintln(os.Stderr, "\nShutting down gracefully… (press Ctrl+C again to force quit)")
		cancel()
		<-sigCh
		fmt.Fprintln(os.Stderr, "Forced quit.")
		os.Exit(130) // 128 + SIGINT
	}()

	if err := app.Run(ctx); err != nil {
		// A cancelled context is a clean, signal-driven shutdown.
		if errors.Is(err, context.Canceled) {
			os.Exit(0)
		}
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
