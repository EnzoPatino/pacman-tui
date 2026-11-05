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
			m, prevFirefly, wasPaused := g.Maze, g.Firefly, g.endPause > 0
			prevRobots := make([]maze.Point, len(g.Robots))
			for i, s := range g.Robots {
				prevRobots[i] = s.Pos
			}
			g.Step()
			if wasPaused {
				continue // characters may have gone back to their starting positions
			}
			if g.Maze.IsWall(g.Firefly) || !isOneMove(m, prevFirefly, g.Firefly) {
				t.Fatalf("seed %d: firefly jumped %v -> %v", seed, prevFirefly, g.Firefly)
			}
			for i, s := range g.Robots {
				if s.Pos != m.Den && !isOneMove(m, prevRobots[i], s.Pos) {
					t.Fatalf("seed %d: robot jumped %v -> %v", seed, prevRobots[i], s.Pos)
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

func TestRobotDoesNotReverseOnItsOwn(t *testing.T) {
	g := newTestGame(5)
	for range 40 {
		g.Robots[0].waitTicks = 0
		before := g.Robots[0].dir
		if got := g.chooseRobotDir(g.Robots[0]); got == before.Opposite() {
			t.Fatalf("robot reversed from %v to %v", before, got)
		}
		g.Step()
	}
}

func TestAmbusherAimsAheadOfTheFirefly(t *testing.T) {
	g := newTestGame(1)
	g.Firefly = maze.Point{X: 10, Y: 10}
	g.heading = maze.Right
	got, ok := g.robotTarget(&Robot{Personality: Ambusher})
	want := maze.Point{X: 10 + ambushLead, Y: 10}
	if !ok || got != want {
		t.Fatalf("got %v, %v; want %v", got, ok, want)
	}
}

func TestNectarFrightensRobotsUntilTheGlowEnds(t *testing.T) {
	g := newTestGame(1)
	g.items[g.Maze.Index(g.Firefly)] = Nectar
	g.collectItem()
	for _, s := range g.Robots {
		if !s.Scared {
			t.Fatal("robot is not scared after nectar")
		}
	}
	g.glowTicks = 1
	g.updateTimers()
	for _, s := range g.Robots {
		if s.Scared {
			t.Fatal("robot is still scared after the glow ended")
		}
	}
}

func TestEatingAScaredRobotSendsItHome(t *testing.T) {
	g := newTestGame(1)
	s := g.Robots[0]
	s.Scared = true
	s.Pos = g.Firefly
	g.resolveCollisions()
	if g.Caught {
		t.Fatal("a scared robot caught the firefly")
	}
	if s.Pos != g.Maze.Den || s.Scared || s.waitTicks != respawnDelay {
		t.Fatalf("robot was not sent home: %+v", s)
	}
}

func catchFirefly(t *testing.T, g *Game) {
	t.Helper()
	g.Robots[0].Scared = false
	g.Robots[0].Pos = g.Firefly
	g.resolveCollisions()
	if !g.Caught || !g.roundOver() {
		t.Fatal("collision did not end the round")
	}
}

func finishPause(g *Game) {
	for g.endPause > 0 {
		g.Step()
	}
}

func TestBeingCaughtCostsALifeAndKeepsTheMazeAndItsProgress(t *testing.T) {
	g := newTestGame(1)
	original, sparks := g.Maze, g.sparksLeft-1
	g.sparksLeft = sparks
	g.Firefly = maze.Point{X: 1, Y: 1}
	catchFirefly(t, g)
	finishPause(g)
	if g.Maze != original || g.Caught || g.Lives() != startingLives-1 {
		t.Fatalf("maze kept: %v, caught: %v, lives: %d", g.Maze == original, g.Caught, g.Lives())
	}
	if g.sparksLeft != sparks || g.Firefly != g.start || !g.Ready() {
		t.Fatalf("sparks %d, firefly %v (start %v), ready %v", g.sparksLeft, g.Firefly, g.start, g.Ready())
	}
}

func TestSeveralRobotsCatchingTheFireflyCostOnlyOneLife(t *testing.T) {
	g := newTestGame(1)
	g.Robots[1].Scared = false
	g.Robots[1].Pos = g.Firefly
	catchFirefly(t, g)
	if g.Lives() != startingLives-1 {
		t.Fatalf("lives: got %d, want %d", g.Lives(), startingLives-1)
	}
}

func TestLosingTheLastLifeStartsANewGameAndKeepsTheHighScore(t *testing.T) {
	g := newTestGame(1)
	g.addScore(500)
	g.level = 4
	original := g.Maze
	for life := range startingLives {
		catchFirefly(t, g)
		if life == startingLives-1 && !g.GameOver() {
			t.Fatal("the last life did not end the game")
		}
		finishPause(g)
	}
	if g.Maze == original || g.GameOver() || !g.Ready() {
		t.Fatal("a new game did not start in a new maze")
	}
	if g.Score() != 0 || g.Level() != 1 || g.Lives() != startingLives {
		t.Fatalf("score %d, level %d, lives %d", g.Score(), g.Level(), g.Lives())
	}
	if g.HighScore() != 500 {
		t.Fatalf("high score: got %d, want 500", g.HighScore())
	}
}

func TestClearingAllSparksAdvancesToANewMaze(t *testing.T) {
	g := newTestGame(1)
	original := g.Maze
	g.addScore(100)
	g.sparksLeft = 0
	if !g.roundOver() {
		t.Fatal("clearing all sparks did not end the round")
	}
	finishPause(g)
	if g.Maze == original || g.Level() != 2 || g.Score() != 100 || g.Lives() != startingLives {
		t.Fatalf("new maze: %v, level %d, score %d, lives %d", g.Maze != original, g.Level(), g.Score(), g.Lives())
	}
}

func TestNobodyMovesWhileGettingReady(t *testing.T) {
	g := newTestGame(1)
	firefly := g.Firefly
	for range readyPause {
		if !g.Ready() {
			t.Fatal("the ready pause ended too early")
		}
		g.Step()
		if g.Firefly != firefly || g.Tick() != 0 {
			t.Fatal("the game advanced during the ready pause")
		}
	}
	if g.Ready() {
		t.Fatal("the ready pause did not end")
	}
	g.Step()
	if g.Firefly == firefly {
		t.Fatal("the firefly did not start moving")
	}
}

func TestScoring(t *testing.T) {
	g := newTestGame(1)
	cell := g.Maze.Index(g.Firefly)
	g.items[cell] = Spark
	g.collectItem()
	if g.Score() != sparkPoints {
		t.Fatalf("spark: got %d", g.Score())
	}
	g.items[cell] = Nectar
	g.collectItem()
	want := sparkPoints + nectarPoints
	if g.Score() != want {
		t.Fatalf("nectar: got %d, want %d", g.Score(), want)
	}
	for i, s := range g.Robots {
		s.Pos = g.Firefly
		g.resolveCollisions()
		want += comboPoints << i
		if g.Score() != want {
			t.Fatalf("robot %d: got %d, want %d", i, g.Score(), want)
		}
	}
	g.items[cell] = Nectar
	g.collectItem()
	if g.combo != 0 {
		t.Fatal("a new nectar did not restart the combo")
	}
}

func TestGlowGetsShorterWithLevelsButNeverTooShort(t *testing.T) {
	for _, tc := range []struct{ level, want int }{{1, 60}, {3, 50}, {100, glowMinDuration}} {
		g := newTestGame(1)
		g.level = tc.level
		g.items[g.Maze.Index(g.Firefly)] = Nectar
		g.collectItem()
		if g.GlowTicks() != tc.want || g.GlowFraction() != 1 {
			t.Errorf("level %d: glow %d, fraction %v; want %d, 1", tc.level, g.GlowTicks(), g.GlowFraction(), tc.want)
		}
	}
}

func TestRobotsSpeedUpWithLevelAndWhenFewSparksRemain(t *testing.T) {
	g := newTestGame(1)
	if got := g.robotRestInterval(); got != robotRestEvery {
		t.Fatalf("start: got %d", got)
	}
	g.level = fastLevel
	if got := g.robotRestInterval(); got != fastRestEvery {
		t.Fatalf("fast level: got %d", got)
	}
	g.sparksLeft = g.totalSparks / enragedSparkShare
	if got := g.robotRestInterval(); got != 0 {
		t.Fatalf("few sparks: got %d", got)
	}
	g.Robots[0].Scared = false
	g.tick = 10
	if g.isResting(g.Robots[0]) {
		t.Fatal("an enraged robot should never rest")
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
