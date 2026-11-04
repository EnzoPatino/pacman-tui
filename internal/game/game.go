// Package game simulates Glimmer: an autonomous firefly collects sparks in a
// maze while spiders hunt it. The simulation advances one tick per Step and has
// no knowledge of rendering.
package game

import (
	"math/rand/v2"

	"github.com/EnzoPatino/pacman-tui/internal/maze"
)

const (
	nectarQuadrants = 4
	glowDuration    = 60
	roundEndPause   = 12
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
	Spiders []*Spider
	Caught  bool // the firefly was caught and the round is about to restart

	width, height int
	rng           *rand.Rand
	tick          int
	items         []Item
	sparksLeft    int
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
	g := &Game{width: width, height: height, rng: rng}
	g.startRound()
	return g
}

func (g *Game) startRound() {
	g.Maze = maze.Generate(g.width, g.height, g.rng)
	g.Firefly = farthestCellFromDen(g.Maze)
	g.Spiders = newSpiders(g.Maze.Den, g.rng)
	g.Caught = false
	g.heading = maze.Left
	g.lastVisit = make([]int, g.Maze.W*g.Maze.H)
	g.idleTicks = 0
	g.glowTicks = 0
	g.scattering = false
	g.modeTicks = chaseDuration
	g.endPause = 0
	g.placeItems()
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

// Step advances the simulation by one tick. Once a round ends, it holds the
// final frame for a moment and then starts a new round automatically.
func (g *Game) Step() {
	if g.endPause > 0 {
		g.endPause--
		if g.endPause == 0 {
			g.startRound()
		}
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
	g.moveSpiders()
	g.resolveCollisions()
	g.roundOver()
}

func (g *Game) updateTimers() {
	if g.glowTicks > 0 {
		g.glowTicks--
		if g.glowTicks == 0 {
			g.calmSpiders()
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
	g.reverseSpiders()
}

func (g *Game) collectItem() {
	cell := g.Maze.Index(g.Firefly)
	g.idleTicks++
	switch g.items[cell] {
	case Spark:
		g.sparksLeft--
		g.idleTicks = 0
	case Nectar:
		g.glowTicks = glowDuration
		g.frightenSpiders()
	}
	g.items[cell] = NoItem
}

func (g *Game) resolveCollisions() {
	for _, s := range g.Spiders {
		if s.Pos != g.Firefly {
			continue
		}
		if s.Scared {
			s.sendHome(g.Maze.Den)
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
	if g.Caught || g.sparksLeft == 0 {
		g.endPause = roundEndPause
		return true
	}
	return false
}
