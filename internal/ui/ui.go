// Package ui runs the game loop on a terminal screen and draws the simulation.
package ui

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/EnzoPatino/pacman-tui/internal/game"
)

const tickInterval = 140 * time.Millisecond

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

	r := newRenderer(screen)
	g := newGameForScreen(screen, rng)
	r.draw(g)
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
				r.redraw(g)
			}
		case <-ticker.C:
			if g != nil {
				g.Step()
				r.draw(g)
			}
		}
	}
}

// newGameForScreen fits a new game to the terminal, or returns nil if the
// terminal is smaller than the minimum supported size.
func newGameForScreen(screen tcell.Screen, rng *rand.Rand) *game.Game {
	columns, rows := screen.Size()
	fit, ok := fitTerminal(columns, rows)
	if !ok {
		return nil
	}
	return game.New(fit.mazeW, fit.mazeH, rng)
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
