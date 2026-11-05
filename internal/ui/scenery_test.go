package ui

import (
	"math/rand/v2"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/EnzoPatino/pacman-tui/internal/game"
	"github.com/EnzoPatino/pacman-tui/internal/maze"
)

func TestEveryWallConnectionHasAJoint(t *testing.T) {
	for links, joint := range wallJoints {
		if joint == 0 {
			t.Errorf("links %04b have no joint", links)
		}
	}
}

func TestWallJointsMatchTheirNeighbors(t *testing.T) {
	m := maze.Generate(41, 21, rand.New(rand.NewPCG(1, 1)))
	if got := wallLinks(m, maze.Point{X: 2, Y: 0}); got&linkLeft == 0 || got&linkRight == 0 || got&linkUp != 0 {
		t.Fatalf("top border: got links %04b", got)
	}
}

// paintedFrame returns the frame of a game with the given terminal size.
func paintedFrame(columns, rows int, g *game.Game) *frame {
	f := &frame{}
	f.reset(columns, rows)
	paintGame(f, g)
	return f
}

// TestSpritesNeverCoverWalls plays at both scales, checking that every wall's
// center still shows a line. Sprites are designed to fit inside corridors.
func TestSpritesNeverCoverWalls(t *testing.T) {
	sizes := [][2]int{{120, 45}, {80, 24}}
	for _, size := range sizes {
		fitted, ok := fitTerminal(size[0], size[1])
		if !ok {
			t.Fatalf("%v does not fit", size)
		}
		g := game.New(fitted.mazeW, fitted.mazeH, rand.New(rand.NewPCG(4, 4)))
		for range 600 {
			g.Step()
			if g.Ready() || g.GameOver() {
				continue // the banner is meant to cover the maze
			}
			f := paintedFrame(size[0], size[1], g)
			c := newCanvas(f, g, fitted.scale)
			for y := range g.Maze.H {
				for x := range g.Maze.W {
					p := maze.Point{X: x, Y: y}
					if !g.Maze.IsWall(p) {
						continue
					}
					px, py := c.position(p)
					want := wallJoints[wallLinks(g.Maze, p)]
					if got := f.cells[py*f.w+px].r; got != want {
						t.Fatalf("%v: wall %v shows %q, want %q", size, p, got, want)
					}
				}
			}
		}
	}
}

func TestMazeAndSpritesFitTheTerminal(t *testing.T) {
	sizes := [][2]int{{33, 16}, {63, 27}, {80, 24}, {100, 40}, {181, 50}, {400, 120}}
	for _, size := range sizes {
		fitted, ok := fitTerminal(size[0], size[1])
		if !ok {
			t.Errorf("%v: no fit", size)
			continue
		}
		g := game.New(fitted.mazeW, fitted.mazeH, rand.New(rand.NewPCG(1, 1)))
		f := &frame{}
		f.reset(size[0], size[1])
		c := newCanvas(f, g, fitted.scale)
		left, top := c.position(maze.Point{})
		right, bottom := c.position(maze.Point{X: g.Maze.W - 1, Y: g.Maze.H - 1})
		if left < fitted.marginX || right >= size[0]-fitted.marginX {
			t.Errorf("%v: columns %d..%d leave no room for sprites in tunnels", size, left, right)
		}
		if top < hudRows || bottom >= size[1]-footerRows+1 {
			t.Errorf("%v: rows %d..%d overlap the header or footer", size, top, bottom)
		}
	}
}

func TestLargeScaleIsUsedWhenItFits(t *testing.T) {
	if f, ok := fitTerminal(180, 50); !ok || !f.large {
		t.Fatalf("got %+v, %v; want large scale", f, ok)
	}
	if f, ok := fitTerminal(80, 24); !ok || f.large {
		t.Fatalf("got %+v, %v; want small scale", f, ok)
	}
	if _, ok := fitTerminal(20, 5); ok {
		t.Fatal("a tiny terminal should not fit")
	}
}

func TestSparksTwinkleOccasionallyAndAtDifferentTimes(t *testing.T) {
	const cells = 2000
	twinkling := 0
	for i := range cells {
		if glyph, _ := spark(maze.Point{X: i % 80, Y: i / 80}, 0, 0, true); glyph == twinkleGlyph {
			twinkling++
		}
	}
	if twinkling == 0 || twinkling > cells/10 {
		t.Fatalf("%d of %d sparks are twinkling at once", twinkling, cells)
	}
}

func TestNectarStandsOutAndAlternatesShape(t *testing.T) {
	floor := nectarDim.mix(nectarLit, nectarMinBright)
	for tick := range 48 {
		if _, color := nectar(tick, 0); color.r < floor.r-0.01 {
			t.Fatalf("tick %d: nectar is dimmer than its minimum: %v", tick, color)
		}
	}
	first, _ := nectar(0, 0)
	second, _ := nectar(nectarBloomTicks, 0)
	if first == second {
		t.Fatal("nectar does not change shape")
	}
}

func TestTrailFades(t *testing.T) {
	_, young := trailSpark(1)
	_, old := trailSpark(trailLength - 1)
	if young == old {
		t.Fatal("trail does not fade with age")
	}
}

func TestProgressBarKeepsItsWidthAndShowsProgress(t *testing.T) {
	cases := []struct {
		fraction float64
		filled   int
	}{{0, 0}, {0.001, 1}, {0.5, 10}, {1, 20}, {1.5, 20}}
	for _, tc := range cases {
		filled, empty := progressBar(tc.fraction, 20)
		if got := utf8.RuneCountInString(filled); got != tc.filled {
			t.Errorf("fraction %v: %d filled, want %d", tc.fraction, got, tc.filled)
		}
		if got := utf8.RuneCountInString(filled + empty); got != 20 {
			t.Errorf("fraction %v: total width %d, want 20", tc.fraction, got)
		}
	}
}

func TestClearedMazeLightsUpAndAnnouncesIt(t *testing.T) {
	g := game.New(39, 31, rand.New(rand.NewPCG(1, 1)))
	far := maze.Point{X: 1, Y: 1}
	g.Firefly = maze.Point{X: g.Maze.W - 2, Y: g.Maze.H - 2}
	if lightAt(g, far) >= 1 {
		t.Fatal("a distant cell should not be fully lit during play")
	}
	cleared := playUntilCleared(t)
	if lightAt(cleared, far) != 1 {
		t.Fatal("a cleared maze should be fully lit")
	}
	if got := statusFooter(cleared); !strings.HasPrefix(got, "MAZE CLEARED") {
		t.Fatalf("got footer %q", got)
	}
}

// playUntilCleared plays real rounds on small mazes until the firefly wins one.
func playUntilCleared(t *testing.T) *game.Game {
	t.Helper()
	for seed := range uint64(100) {
		g := game.New(maze.MinWidth, maze.MinHeight, rand.New(rand.NewPCG(seed, 9)))
		for !g.Caught && g.SparksLeft() > 0 {
			g.Step()
		}
		if g.SparksLeft() == 0 {
			return g
		}
	}
	t.Fatal("the firefly never cleared a maze")
	return nil
}
