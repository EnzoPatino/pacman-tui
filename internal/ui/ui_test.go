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

func containsText(screen tcell.SimulationScreen, text string) bool {
	return strings.Contains(screenText(screen), text)
}

func testRng() *rand.Rand { return rand.New(rand.NewPCG(1, 1)) }

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

// writeCounter counts what reaches the screen, since every write has a cost.
type writeCounter struct {
	tcell.Screen
	clears, writes int
}

func (c *writeCounter) Clear() {
	c.clears++
	c.Screen.Clear()
}

func (c *writeCounter) SetContent(x, y int, r rune, combining []rune, style tcell.Style) {
	c.writes++
	c.Screen.SetContent(x, y, r, combining, style)
}

func TestFramesOnlyWriteWhatChanged(t *testing.T) {
	screen := &writeCounter{Screen: newTestScreen(t, 120, 45)}
	r := newRenderer(screen)
	g := newGameForScreen(screen, testRng())
	for g.Ready() {
		g.Step()
	}
	r.draw(g)
	first := screen.writes
	if first == 0 {
		t.Fatal("the first frame wrote nothing")
	}

	screen.writes = 0
	r.draw(g)
	if screen.writes != 0 {
		t.Fatalf("redrawing an unchanged game wrote %d cells", screen.writes)
	}

	g.Step()
	r.draw(g)
	if screen.writes == 0 || screen.writes > first/4 {
		t.Fatalf("one step wrote %d cells, out of %d on screen", screen.writes, first)
	}
	if screen.clears != 0 {
		t.Fatalf("got %d clears while drawing frames", screen.clears)
	}
}

func TestRedrawLeavesNothingOfThePreviousLayout(t *testing.T) {
	screen := newTestScreen(t, 80, 40)
	r := newRenderer(screen)
	r.draw(nil)
	if !containsText(screen, "too small") {
		t.Fatal("notice was not drawn")
	}
	r.redraw(newGameForScreen(screen, testRng()))
	if containsText(screen, "too small") {
		t.Fatal("notice is still visible after the redraw")
	}
}

func TestNarrowestTerminalShowsTheGame(t *testing.T) {
	screen := newTestScreen(t, 33, 16)
	g := newGameForScreen(screen, testRng())
	if g == nil {
		t.Fatal("no game at the minimum size")
	}
	newRenderer(screen).draw(g)
	if !containsText(screen, "SCORE") || !strings.ContainsAny(screenText(screen), "◕◔◒◓") {
		t.Fatal("the header or the firefly is missing")
	}
}

func TestBannerAnnouncesTheLevelAndGoesAway(t *testing.T) {
	screen := newTestScreen(t, 120, 45)
	r := newRenderer(screen)
	g := newGameForScreen(screen, testRng())
	r.draw(g)
	if !containsText(screen, "L E V E L   1") {
		t.Fatal("the level was not announced")
	}
	for g.Ready() {
		g.Step()
	}
	r.draw(g)
	if containsText(screen, "L E V E L") {
		t.Fatal("the banner is still visible after the round started")
	}
}

func TestStatusLineFitsEveryWidth(t *testing.T) {
	for _, columns := range []int{33, 60, 89, 90, 120, 200} {
		screen := newTestScreen(t, columns, 45)
		g := newGameForScreen(screen, testRng())
		f := &frame{}
		f.reset(columns, 45)
		fitted, _ := fitTerminal(columns, 45)
		line := newCanvas(f, g, fitted.scale).statusLine()
		length := 0
		if columns >= wideStatusColumns {
			length = len("   GLOW ") + statusBarWidth/2 // only shown while glowing
		}
		for _, s := range line {
			length += len([]rune(s.text))
		}
		if length > columns {
			t.Errorf("%d columns: the status line can take %d", columns, length)
		}
	}
}

func TestSmallTerminalShowsNoticeAndRecoversOnResize(t *testing.T) {
	screen := newTestScreen(t, 20, 5)
	rng := rand.New(rand.NewPCG(1, 1))
	r := newRenderer(screen)
	g := newGameForScreen(screen, rng)
	if g != nil {
		t.Fatal("expected no game in a tiny terminal")
	}
	r.draw(g)
	if text := screenText(screen); !strings.Contains(text, "too small") {
		t.Fatalf("notice is cut off or missing: %q", strings.TrimSpace(text))
	}

	screen.SetSize(80, 24)
	if g = newGameForScreen(screen, rng); g == nil {
		t.Fatal("no game after enlarging the terminal")
	}
	r.redraw(g)
	if !strings.ContainsAny(screenText(screen), "◕◔◒◓") {
		t.Fatal("firefly is not drawn after resizing")
	}
}
