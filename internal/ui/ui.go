// Package ui runs the game loop on a terminal screen and draws the simulation.
package ui

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/EnzoPatino/pacman-tui/internal/game"
	"github.com/EnzoPatino/pacman-tui/internal/maze"
)

const (
	tickInterval = 100 * time.Millisecond
	// cellColumns compensates for terminal cells being about twice as tall as
	// they are wide, so that maze cells look square.
	cellColumns = 2
)

// Run plays the simulation until the context ends or the user asks to quit.
// Terminal state is restored before it returns.
func Run(ctx context.Context) error {
	screen, err := tcell.NewScreen()
	if err != nil {
		return fmt.Errorf("opening terminal: %w", err)
	}
	if err := screen.Init(); err != nil {
		return fmt.Errorf("initializing terminal: %w", err)
	}
	defer screen.Fini()

	seed := uint64(time.Now().UnixNano())
	loop(ctx, screen, rand.New(rand.NewPCG(seed, seed)))
	return nil
}

func loop(ctx context.Context, screen tcell.Screen, rng *rand.Rand) {
	events := make(chan tcell.Event, 16)
	quitEvents := make(chan struct{})
	defer close(quitEvents)
	go screen.ChannelEvents(events, quitEvents)

	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	g := newGameForScreen(screen, rng)
	draw(screen, g)
	for {
		select {
		case <-ctx.Done():
			return
		case ev := <-events:
			switch ev := ev.(type) {
			case *tcell.EventKey:
				if isQuitKey(ev) {
					return
				}
			case *tcell.EventResize:
				g = newGameForScreen(screen, rng)
				screen.Sync()
				draw(screen, g)
			}
		case <-ticker.C:
			if g != nil {
				g.Step()
				draw(screen, g)
			}
		}
	}
}

// newGameForScreen fits a new game to the terminal, or returns nil if the
// terminal is smaller than the minimum supported size.
func newGameForScreen(screen tcell.Screen, rng *rand.Rand) *game.Game {
	columns, rows := screen.Size()
	width := columns / cellColumns
	if width < maze.MinWidth || rows < maze.MinHeight {
		return nil
	}
	return game.New(width, rows, rng)
}

func isQuitKey(ev *tcell.EventKey) bool {
	switch ev.Key() {
	case tcell.KeyEscape, tcell.KeyCtrlC:
		return true
	case tcell.KeyRune:
		return ev.Rune() == 'q' || ev.Rune() == 'Q'
	default:
		return false
	}
}
