package game

import (
	"math/rand/v2"

	"github.com/EnzoPatino/pacman-tui/internal/maze"
)

const (
	chaseDuration   = 200
	scatterDuration = 70
	releaseInterval = 20 // ticks between spiders leaving the den
	respawnDelay    = 30
	ambushLead      = 4  // cells ahead of the firefly the ambusher aims for
	spiderRestEvery = 10 // spiders skip one tick in this many, so the firefly is faster
	scaredRestEvery = 2  // scared spiders move at half speed
)

// Personality decides where a spider heads while hunting.
type Personality int

// The spider personalities.
const (
	Hunter   Personality = iota // heads straight for the firefly
	Ambusher                    // aims ahead of the firefly's path
	Drifter                     // wanders at random
)

var personalities = [...]Personality{Hunter, Ambusher, Drifter}

// Spider is an enemy. Scared spiders flee erratically and can be eaten.
type Spider struct {
	Pos         maze.Point
	Personality Personality
	Scared      bool

	dir       maze.Dir
	waitTicks int
}

func newSpiders(den maze.Point, rng *rand.Rand) []*Spider {
	spiders := make([]*Spider, len(personalities))
	for i, p := range personalities {
		spiders[i] = &Spider{
			Pos:         den,
			Personality: p,
			dir:         maze.Dirs[rng.IntN(len(maze.Dirs))],
			waitTicks:   i * releaseInterval,
		}
	}
	return spiders
}

func (s *Spider) sendHome(den maze.Point) {
	s.Pos = den
	s.Scared = false
	s.waitTicks = respawnDelay
}

func (g *Game) frightenSpiders() {
	for _, s := range g.Spiders {
		s.Scared = true
		s.dir = s.dir.Opposite()
	}
}

func (g *Game) calmSpiders() {
	for _, s := range g.Spiders {
		s.Scared = false
	}
}

func (g *Game) reverseSpiders() {
	for _, s := range g.Spiders {
		s.dir = s.dir.Opposite()
	}
}

func (g *Game) moveSpiders() {
	for _, s := range g.Spiders {
		if s.waitTicks > 0 {
			s.waitTicks--
			continue
		}
		if g.isResting(s) {
			continue
		}
		s.dir = g.chooseSpiderDir(s)
		if next, ok := g.Maze.Step(s.Pos, s.dir); ok {
			s.Pos = next
		}
	}
}

func (g *Game) isResting(s *Spider) bool {
	if s.Scared {
		return g.tick%scaredRestEvery == 0
	}
	return g.tick%spiderRestEvery == 0
}

// chooseSpiderDir never reverses unless it is the only way out; reversals
// otherwise happen only on mode changes.
func (g *Game) chooseSpiderDir(s *Spider) maze.Dir {
	options := make([]maze.Dir, 0, len(maze.Dirs))
	for _, d := range g.Maze.Exits(s.Pos) {
		if d != s.dir.Opposite() {
			options = append(options, d)
		}
	}
	if len(options) == 0 {
		return s.dir.Opposite()
	}
	target, hasTarget := g.spiderTarget(s)
	if !hasTarget {
		return options[g.rng.IntN(len(options))]
	}
	return g.closestOption(s.Pos, options, target)
}

// spiderTarget returns the cell a spider aims for, or false if it wanders.
func (g *Game) spiderTarget(s *Spider) (maze.Point, bool) {
	if s.Scared {
		return maze.Point{}, false
	}
	if g.scattering {
		return g.scatterCorner(s.Personality), true
	}
	switch s.Personality {
	case Hunter:
		return g.Firefly, true
	case Ambusher:
		dx, dy := g.heading.Delta()
		return maze.Point{X: g.Firefly.X + ambushLead*dx, Y: g.Firefly.Y + ambushLead*dy}, true
	default:
		return maze.Point{}, false
	}
}

func (g *Game) scatterCorner(p Personality) maze.Point {
	m := g.Maze
	switch p {
	case Hunter:
		return maze.Point{X: m.W - 2, Y: 1}
	case Ambusher:
		return maze.Point{X: 1, Y: 1}
	default:
		return maze.Point{X: m.W - 2, Y: m.H - 2}
	}
}

func (g *Game) closestOption(from maze.Point, options []maze.Dir, target maze.Point) maze.Dir {
	best := options[0]
	bestDistance := -1
	ties := 0
	for _, d := range options {
		next, _ := g.Maze.Step(from, d)
		distance := manhattan(next, target)
		switch {
		case bestDistance < 0 || distance < bestDistance:
			best, bestDistance, ties = d, distance, 1
		case distance == bestDistance:
			ties++
			if g.rng.IntN(ties) == 0 {
				best = d
			}
		}
	}
	return best
}

func manhattan(a, b maze.Point) int {
	return abs(a.X-b.X) + abs(a.Y-b.Y)
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
