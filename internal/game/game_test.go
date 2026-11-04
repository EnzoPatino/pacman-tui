package game

import (
	"math/rand/v2"
	"testing"

	"github.com/EnzoPatino/pacman-tui/internal/maze"
)

func newTestGame(seed uint64) *Game {
	return New(41, 21, rand.New(rand.NewPCG(seed, 7)))
}

func TestItemsAreReachableAndLeaveTheDenAndStartEmpty(t *testing.T) {
	for seed := range uint64(30) {
		g := newTestGame(seed)
		dist := g.Maze.Distances(g.Firefly)
		sparks, nectars := 0, 0
		for y := range g.Maze.H {
			for x := range g.Maze.W {
				p := maze.Point{X: x, Y: y}
				item := g.ItemAt(p)
				if item == NoItem {
					continue
				}
				if g.Maze.IsWall(p) || g.Maze.InDen(p) || p == g.Firefly {
					t.Fatalf("seed %d: item placed on invalid cell %v", seed, p)
				}
				if dist[g.Maze.Index(p)] == maze.Unreachable {
					t.Fatalf("seed %d: item at %v is unreachable", seed, p)
				}
				if item == Spark {
					sparks++
				} else {
					nectars++
				}
			}
		}
		if sparks != g.sparksLeft || nectars != nectarQuadrants {
			t.Fatalf("seed %d: sparks %d (counter %d), nectars %d", seed, sparks, g.sparksLeft, nectars)
		}
	}
}

func TestCharactersOnlyMoveOneCellThroughFloor(t *testing.T) {
	for seed := range uint64(10) {
		g := newTestGame(seed)
		for range 2000 {
			m, prevFirefly := g.Maze, g.Firefly
			prevSpiders := make([]maze.Point, len(g.Spiders))
			for i, s := range g.Spiders {
				prevSpiders[i] = s.Pos
			}
			g.Step()
			if g.Maze != m {
				continue // a new round started
			}
			if g.Maze.IsWall(g.Firefly) || !isOneMove(m, prevFirefly, g.Firefly) {
				t.Fatalf("seed %d: firefly jumped %v -> %v", seed, prevFirefly, g.Firefly)
			}
			for i, s := range g.Spiders {
				if s.Pos != m.Den && !isOneMove(m, prevSpiders[i], s.Pos) {
					t.Fatalf("seed %d: spider jumped %v -> %v", seed, prevSpiders[i], s.Pos)
				}
			}
		}
	}
}

func isOneMove(m *maze.Maze, from, to maze.Point) bool {
	if from == to {
		return true
	}
	for _, d := range maze.Dirs {
		if next, ok := m.Step(from, d); ok && next == to {
			return true
		}
	}
	return false
}

func TestSpiderDoesNotReverseOnItsOwn(t *testing.T) {
	g := newTestGame(5)
	for range 40 {
		g.Spiders[0].waitTicks = 0
		before := g.Spiders[0].dir
		if got := g.chooseSpiderDir(g.Spiders[0]); got == before.Opposite() {
			t.Fatalf("spider reversed from %v to %v", before, got)
		}
		g.Step()
	}
}

func TestAmbusherAimsAheadOfTheFirefly(t *testing.T) {
	g := newTestGame(1)
	g.Firefly = maze.Point{X: 10, Y: 10}
	g.heading = maze.Right
	got, ok := g.spiderTarget(&Spider{Personality: Ambusher})
	want := maze.Point{X: 10 + ambushLead, Y: 10}
	if !ok || got != want {
		t.Fatalf("got %v, %v; want %v", got, ok, want)
	}
}

func TestNectarFrightensSpidersUntilTheGlowEnds(t *testing.T) {
	g := newTestGame(1)
	g.items[g.Maze.Index(g.Firefly)] = Nectar
	g.collectItem()
	for _, s := range g.Spiders {
		if !s.Scared {
			t.Fatal("spider is not scared after nectar")
		}
	}
	g.glowTicks = 1
	g.updateTimers()
	for _, s := range g.Spiders {
		if s.Scared {
			t.Fatal("spider is still scared after the glow ended")
		}
	}
}

func TestEatingAScaredSpiderSendsItHome(t *testing.T) {
	g := newTestGame(1)
	s := g.Spiders[0]
	s.Scared = true
	s.Pos = g.Firefly
	g.resolveCollisions()
	if g.Caught {
		t.Fatal("a scared spider caught the firefly")
	}
	if s.Pos != g.Maze.Den || s.Scared || s.waitTicks != respawnDelay {
		t.Fatalf("spider was not sent home: %+v", s)
	}
}

func TestBeingCaughtRestartsWithANewMaze(t *testing.T) {
	g := newTestGame(1)
	original := g.Maze
	g.Spiders[0].Pos = g.Firefly
	g.resolveCollisions()
	if !g.Caught || !g.roundOver() {
		t.Fatal("collision did not end the round")
	}
	for range roundEndPause {
		g.Step()
	}
	if g.Maze == original || g.Caught {
		t.Fatal("the round did not restart with a new maze")
	}
}

func TestClearingAllSparksRestartsWithANewMaze(t *testing.T) {
	g := newTestGame(1)
	original := g.Maze
	g.sparksLeft = 0
	if !g.roundOver() {
		t.Fatal("clearing all sparks did not end the round")
	}
	for range roundEndPause {
		g.Step()
	}
	if g.Maze == original {
		t.Fatal("the round did not restart with a new maze")
	}
}

func TestGameStartsOnEveryTerminalSize(t *testing.T) {
	for w := maze.MinWidth; w <= 60; w++ {
		for h := maze.MinHeight; h <= 40; h++ {
			New(w, h, rand.New(rand.NewPCG(uint64(w), uint64(h))))
		}
	}
}

// TestFireflyPlaysConvincingly plays complete rounds: the autonomous firefly must
// clear some mazes, lose others, and never stall without collecting.
func TestFireflyPlaysConvincingly(t *testing.T) {
	const rounds = 200
	const maxTicksWithoutProgress = 400

	cleared, finished := 0, 0
	for seed := range uint64(rounds) {
		g := newTestGame(seed)
		m := g.Maze
		sparks, idle := g.sparksLeft, 0
		for g.Maze == m {
			g.Step()
			if g.sparksLeft < sparks {
				sparks, idle = g.sparksLeft, 0
			} else {
				idle++
			}
			if g.endPause > 0 {
				break
			}
			if idle > maxTicksWithoutProgress {
				t.Fatalf("seed %d: no spark collected in %d ticks, %d left", seed, idle, sparks)
			}
		}
		finished++
		if g.sparksLeft == 0 {
			cleared++
		}
	}
	t.Logf("cleared %d of %d mazes", cleared, finished)
	if cleared == 0 || cleared == finished {
		t.Fatalf("cleared %d of %d mazes; want some wins and some losses", cleared, finished)
	}
}
