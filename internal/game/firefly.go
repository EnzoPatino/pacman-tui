package game

import "github.com/EnzoPatino/pacman-tui/internal/maze"

const (
	safetyMargin = 1  // extra steps the firefly wants over any spider
	panicRange   = 7  // spider distance at which nectar becomes the priority
	huntRange    = 12 // farthest scared spider worth chasing
	boldAfter    = 60 // idle ticks after which the firefly stops avoiding spiders
)

func (g *Game) moveFirefly() {
	dir := g.chooseFireflyDir()
	if next, ok := g.Maze.Step(g.Firefly, dir); ok {
		g.Firefly = next
		g.heading = dir
		g.lastVisit[g.Maze.Index(next)] = g.tick
	}
}

// chooseFireflyDir picks the most urgent objective that can be reached without
// a spider getting there first: eat scared spiders, grab nectar when threatened,
// collect sparks. When no objective is safely reachable, the firefly flees. If
// fleeing goes on for too long it takes the risk, so a round can never stall.
func (g *Game) chooseFireflyDir() maze.Dir {
	threat := g.threatField()
	safe := func(p maze.Point, steps int) bool { return g.isSafe(threat, p, steps) }
	if g.idleTicks > boldAfter {
		anywhere := func(maze.Point, int) bool { return true }
		if dir, ok := g.Maze.FirstStepTo(g.Firefly, anywhere, g.holds(Spark)); ok {
			return dir
		}
	}

	if g.glowTicks > 0 {
		huntable := func(p maze.Point, steps int) bool {
			return steps <= min(huntRange, g.glowTicks) && safe(p, steps)
		}
		if dir, ok := g.Maze.FirstStepTo(g.Firefly, huntable, g.hasScaredSpider); ok {
			return dir
		}
	}
	if d := threat[g.Maze.Index(g.Firefly)]; d != maze.Unreachable && d <= panicRange {
		if dir, ok := g.Maze.FirstStepTo(g.Firefly, safe, g.holds(Nectar)); ok {
			return dir
		}
	}
	if dir, ok := g.Maze.FirstStepTo(g.Firefly, safe, g.holds(Spark)); ok {
		return dir
	}
	return g.fleeDir(threat)
}

// threatField returns the distance from every cell to the nearest dangerous
// spider.
func (g *Game) threatField() []int {
	sources := make([]maze.Point, 0, len(g.Spiders))
	for _, s := range g.Spiders {
		if !s.Scared {
			sources = append(sources, s.Pos)
		}
	}
	return g.Maze.Distances(sources...)
}

func (g *Game) isSafe(threat []int, p maze.Point, steps int) bool {
	d := threat[g.Maze.Index(p)]
	return d == maze.Unreachable || steps+safetyMargin < d
}

func (g *Game) hasScaredSpider(p maze.Point) bool {
	for _, s := range g.Spiders {
		if s.Scared && s.Pos == p {
			return true
		}
	}
	return false
}

func (g *Game) holds(item Item) func(maze.Point) bool {
	return func(p maze.Point) bool { return g.ItemAt(p) == item }
}

// fleeDir moves away from the closest spider, preferring the least recently
// visited cell so the firefly does not shuttle between the same two cells.
func (g *Game) fleeDir(threat []int) maze.Dir {
	best, bestDanger, bestVisit := g.heading, -1, 0
	for _, dir := range g.Maze.Exits(g.Firefly) {
		next, _ := g.Maze.Step(g.Firefly, dir)
		danger := threat[g.Maze.Index(next)]
		if danger == maze.Unreachable {
			danger = g.Maze.W * g.Maze.H
		}
		visit := g.lastVisit[g.Maze.Index(next)]
		if danger > bestDanger || (danger == bestDanger && visit < bestVisit) {
			best, bestDanger, bestVisit = dir, danger, visit
		}
	}
	return best
}
