// Package game simulates Glimmer: an autonomous firefly collects sparks in a
// maze while robots hunt it. The simulation advances one tick per Step and has
// no knowledge of rendering.
package game

import (
	"math/rand/v2"

	"github.com/EnzoPatino/pacman-tui/internal/maze"
)

const (
	nectarQuadrants = 4
	roundEndPause   = 12
	gameOverPause   = 30
	readyPause      = 14 // ticks everyone holds still before a round starts

	startingLives = 3
	sparkPoints   = 10
	nectarPoints  = 50
	comboPoints   = 150 // for the first robot eaten on a nectar, then twice that...

	glowDuration    = 60
	glowLevelCut    = 5 // each level shortens the glow by this much
	glowMinDuration = 30
)

// Item is what a cell holds for the firefly to collect.
type Item int

// The collectable items.
const (
	NoItem Item = iota
	Spark
	Nectar
)

// Game holds the state of the current round.
type Game struct {
	Maze    *maze.Maze
	Firefly maze.Point
	Robots  []*Robot
	Caught  bool // the firefly was caught and the round is about to restart

	score, highScore int
	lives, level     int
	combo            int // robots eaten on the current nectar
	gameOver         bool
	readyTicks       int
	start            maze.Point // where the firefly starts in this maze
	glowFull         int        // the duration of the current glow

	width, height int
	rng           *rand.Rand
	tick          int
	items         []Item
	sparksLeft    int
	totalSparks   int
	heading       maze.Dir
	lastVisit     []int
	idleTicks     int // ticks since the firefly last collected a spark
	glowTicks     int
	modeTicks     int
	scattering    bool
	endPause      int
}

// New starts a game on a maze of the given size. Dimensions must be at least
// maze.MinWidth by maze.MinHeight.
func New(width, height int, rng *rand.Rand) *Game {
	g := &Game{width: width, height: height, rng: rng, lives: startingLives, level: 1}
	g.startRound()
	return g
}

// startRound begins a level in a new maze.
func (g *Game) startRound() {
	g.Maze = maze.Generate(g.width, g.height, g.rng)
	g.start = farthestCellFromDen(g.Maze)
	g.Firefly = g.start
	g.placeItems()
	g.placeCharacters()
}

// placeCharacters puts everyone at their starting positions, leaving items alone.
func (g *Game) placeCharacters() {
	g.Firefly = g.start
	g.Robots = newRobots(g.Maze.Den, g.rng)
	g.Caught = false
	g.gameOver = false
	g.heading = maze.Left
	g.lastVisit = make([]int, g.Maze.W*g.Maze.H)
	g.idleTicks = 0
	g.glowTicks = 0
	g.combo = 0
	g.scattering = false
	g.modeTicks = chaseDuration
	g.endPause = 0
	g.readyTicks = readyPause
}

// nextPhase runs when the pause after a round's end is over.
func (g *Game) nextPhase() {
	switch {
	case g.gameOver:
		g.score, g.lives, g.level = 0, startingLives, 1
		g.startRound()
	case g.sparksLeft == 0:
		g.level++
		g.startRound()
	default:
		g.placeCharacters()
	}
}

func farthestCellFromDen(m *maze.Maze) maze.Point {
	dist := m.Distances(m.Den)
	farthest := m.Den
	for y := range m.H {
		for x := range m.W {
			p := maze.Point{X: x, Y: y}
			if dist[m.Index(p)] > dist[m.Index(farthest)] {
				farthest = p
			}
		}
	}
	return farthest
}

func (g *Game) placeItems() {
	m := g.Maze
	g.items = make([]Item, m.W*m.H)
	g.sparksLeft = 0
	for y := range m.H {
		for x := range m.W {
			if p := (maze.Point{X: x, Y: y}); g.canHoldItem(p) {
				g.items[m.Index(p)] = Spark
				g.sparksLeft++
			}
		}
	}
	for quadrant := range nectarQuadrants {
		g.placeNectar(quadrant)
	}
	g.totalSparks = g.sparksLeft
}

func (g *Game) canHoldItem(p maze.Point) bool {
	return !g.Maze.IsWall(p) && !g.Maze.InDen(p) && p != g.Firefly
}

// placeNectar upgrades a random spark in the given quadrant.
func (g *Game) placeNectar(quadrant int) {
	m := g.Maze
	halfW, halfH := m.W/2, m.H/2
	originX, originY := (quadrant%2)*halfW, (quadrant/2)*halfH
	var candidates []int
	for y := originY; y < originY+halfH; y++ {
		for x := originX; x < originX+halfW; x++ {
			if cell := m.Index(maze.Point{X: x, Y: y}); g.items[cell] == Spark {
				candidates = append(candidates, cell)
			}
		}
	}
	g.items[candidates[g.rng.IntN(len(candidates))]] = Nectar
	g.sparksLeft--
}

// ItemAt returns the item held by a cell.
func (g *Game) ItemAt(p maze.Point) Item { return g.items[g.Maze.Index(p)] }

// GlowTicks returns how many ticks of nectar power remain, or zero.
func (g *Game) GlowTicks() int { return g.glowTicks }

// SparkProgress returns the fraction of this round's sparks already collected.
func (g *Game) SparkProgress() float64 {
	return 1 - float64(g.sparksLeft)/float64(g.totalSparks)
}

// GlowFraction returns the fraction of the nectar power still remaining.
func (g *Game) GlowFraction() float64 {
	if g.glowFull == 0 {
		return 0
	}
	return float64(g.glowTicks) / float64(g.glowFull)
}

// Score returns the points of the current game.
func (g *Game) Score() int { return g.score }

// HighScore returns the best score since the program started.
func (g *Game) HighScore() int { return g.highScore }

// Lives returns how many times the firefly can still be caught.
func (g *Game) Lives() int { return g.lives }

// Level returns the current level, starting from 1.
func (g *Game) Level() int { return g.level }

// Ready reports whether a round is about to start and everyone is holding still.
func (g *Game) Ready() bool { return g.readyTicks > 0 }

// GameOver reports whether the last life was lost and a new game is about to start.
func (g *Game) GameOver() bool { return g.gameOver }

func (g *Game) addScore(points int) {
	g.score += points
	g.highScore = max(g.highScore, g.score)
}

// Tick returns the number of simulated ticks, which is useful to animate sprites.
func (g *Game) Tick() int { return g.tick }

// Heading returns the direction the firefly last moved in.
func (g *Game) Heading() maze.Dir { return g.heading }

// TicksSinceVisit returns how long ago the firefly entered a cell, and false if
// it has not been there this round.
func (g *Game) TicksSinceVisit(p maze.Point) (int, bool) {
	visit := g.lastVisit[g.Maze.Index(p)]
	return g.tick - visit, visit > 0
}

// SparksLeft returns how many sparks remain in the current maze.
func (g *Game) SparksLeft() int { return g.sparksLeft }

// Step advances the simulation by one tick. Once a round ends, it holds the
// final frame for a moment and then continues automatically.
func (g *Game) Step() {
	if g.endPause > 0 {
		g.endPause--
		if g.endPause == 0 {
			g.nextPhase()
		}
		return
	}
	if g.readyTicks > 0 {
		g.readyTicks--
		return
	}
	g.tick++
	g.updateTimers()
	g.moveFirefly()
	g.collectItem()
	g.resolveCollisions()
	if g.roundOver() {
		return
	}
	g.moveRobots()
	g.resolveCollisions()
	g.roundOver()
}

func (g *Game) updateTimers() {
	if g.glowTicks > 0 {
		g.glowTicks--
		if g.glowTicks == 0 {
			g.calmRobots()
		}
		return
	}
	g.modeTicks--
	if g.modeTicks > 0 {
		return
	}
	g.scattering = !g.scattering
	g.modeTicks = chaseDuration
	if g.scattering {
		g.modeTicks = scatterDuration
	}
	g.reverseRobots()
}

func (g *Game) collectItem() {
	cell := g.Maze.Index(g.Firefly)
	g.idleTicks++
	switch g.items[cell] {
	case Spark:
		g.sparksLeft--
		g.idleTicks = 0
		g.addScore(sparkPoints)
	case Nectar:
		g.glowFull = max(glowDuration-(g.level-1)*glowLevelCut, glowMinDuration)
		g.glowTicks = g.glowFull
		g.combo = 0
		g.addScore(nectarPoints)
		g.frightenRobots()
	}
	g.items[cell] = NoItem
}

func (g *Game) resolveCollisions() {
	for _, s := range g.Robots {
		if s.Pos != g.Firefly {
			continue
		}
		if s.Scared {
			s.sendHome(g.Maze.Den)
			g.addScore(comboPoints << g.combo)
			g.combo++
		} else {
			g.Caught = true
		}
	}
}

// roundOver reports whether the round has ended, scheduling the restart if so.
func (g *Game) roundOver() bool {
	if g.endPause > 0 {
		return true
	}
	if !g.Caught && g.sparksLeft > 0 {
		return false
	}
	g.endPause = roundEndPause
	if g.Caught {
		g.lives--
		if g.lives == 0 {
			g.gameOver = true
			g.endPause = gameOverPause
		}
	}
	return true
}
