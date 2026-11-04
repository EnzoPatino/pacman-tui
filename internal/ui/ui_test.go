package ui

import (
	"context"
	"math/rand/v2"
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
)

func newTestScreen(t *testing.T, columns, rows int) tcell.SimulationScreen {
	t.Helper()
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(screen.Fini)
	screen.SetSize(columns, rows)
	return screen
}

func screenText(screen tcell.SimulationScreen) string {
	cells, _, _ := screen.GetContents()
	var text strings.Builder
	for _, cell := range cells {
		text.Write(cell.Bytes)
	}
	return text.String()
}

func startLoop(t *testing.T, screen tcell.Screen) (done <-chan struct{}) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	finished := make(chan struct{})
	go func() {
		loop(ctx, screen, rand.New(rand.NewPCG(1, 1)))
		close(finished)
	}()
	return finished
}

func waitFor(t *testing.T, done <-chan struct{}) {
	t.Helper()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("loop did not stop")
	}
}

func TestQuitKeys(t *testing.T) {
	keys := map[string]*tcell.EventKey{
		"q":      tcell.NewEventKey(tcell.KeyRune, 'q', tcell.ModNone),
		"escape": tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone),
		"ctrl+c": tcell.NewEventKey(tcell.KeyCtrlC, 0, tcell.ModCtrl),
	}
	for name, key := range keys {
		t.Run(name, func(t *testing.T) {
			screen := newTestScreen(t, 80, 24)
			done := startLoop(t, screen)
			if err := screen.PostEvent(key); err != nil {
				t.Fatal(err)
			}
			waitFor(t, done)
		})
	}
}

func TestContextCancellationStopsTheLoop(t *testing.T) {
	screen := newTestScreen(t, 80, 24)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		loop(ctx, screen, rand.New(rand.NewPCG(1, 1)))
		close(done)
	}()
	cancel()
	waitFor(t, done)
}

func TestGameFitsTerminalAndCellsLookSquare(t *testing.T) {
	screen := newTestScreen(t, 81, 25)
	g := newGameForScreen(screen, rand.New(rand.NewPCG(1, 1)))
	if g == nil {
		t.Fatal("no game for a large terminal")
	}
	if g.Maze.W*cellColumns > 81 || g.Maze.H > 25 {
		t.Fatalf("maze %dx%d overflows the terminal", g.Maze.W, g.Maze.H)
	}
}

func TestSmallTerminalShowsNoticeAndRecoversOnResize(t *testing.T) {
	screen := newTestScreen(t, 20, 5)
	rng := rand.New(rand.NewPCG(1, 1))
	g := newGameForScreen(screen, rng)
	if g != nil {
		t.Fatal("expected no game in a tiny terminal")
	}
	draw(screen, g)
	if text := screenText(screen); !strings.Contains(text, "too small") {
		t.Fatalf("notice is cut off or missing: %q", strings.TrimSpace(text))
	}

	screen.SetSize(80, 24)
	if g = newGameForScreen(screen, rng); g == nil {
		t.Fatal("no game after enlarging the terminal")
	}
	draw(screen, g)
	if !strings.ContainsRune(screenText(screen), fireflyGlyph) {
		t.Fatal("firefly is not drawn after resizing")
	}
}
