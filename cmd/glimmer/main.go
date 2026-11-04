// Command glimmer is an autonomous maze-chase animation for the terminal.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/EnzoPatino/pacman-tui/internal/ui"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := ui.Run(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "glimmer:", err)
		os.Exit(1)
	}
}
