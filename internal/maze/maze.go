// Package maze generates and navigates braided grid mazes: every floor cell is
// reachable, none is a dead end, and some rows wrap around horizontally.
package maze

import "math/rand/v2"

// Minimum dimensions that leave room for the den and a playable maze.
const (
	MinWidth  = 15
	MinHeight = 11
)

const (
	tunnelCount     = 2
	extraLoopChance = 0.12
	denHalfWidth    = 2
	denHalfHeight   = 1
)

// Unreachable marks cells with no path in the result of Distances.
const Unreachable = -1

// Point is a cell coordinate.
type Point struct{ X, Y int }

// Dir is a movement direction.
type Dir int

// The four movement directions, ordered so that adding 2 gives the opposite.
const (
	Up Dir = iota
	Right
	Down
	Left
)

// Dirs lists every direction.
var Dirs = [4]Dir{Up, Right, Down, Left}

// Opposite returns the reverse direction.
func (d Dir) Opposite() Dir { return (d + 2) % 4 }

// Delta returns the coordinate change of one step in this direction.
func (d Dir) Delta() (dx, dy int) {
	switch d {
	case Up:
		return 0, -1
	case Right:
		return 1, 0
	case Down:
		return 0, 1
	default:
		return -1, 0
	}
}

// Maze is an immutable grid of walls and floor.
type Maze struct {
	W, H  int
	Den   Point
	walls []bool
}

// Generate builds a random maze. Even dimensions are reduced by one because the
// carving algorithm needs odd ones; the caller guarantees the Min dimensions.
func Generate(width, height int, rng *rand.Rand) *Maze {
	w, h := oddFloor(width), oddFloor(height)
	m := &Maze{W: w, H: h, Den: Point{w/2 | 1, h/2 | 1}, walls: make([]bool, w*h)}
	for i := range m.walls {
		m.walls[i] = true
	}
	m.carvePassages(rng)
	m.removeDeadEnds(rng)
	m.openExtraLoops(rng)
	m.carveDen()
	m.openTunnels(rng)
	return m
}

func oddFloor(n int) int {
	if n%2 == 0 {
		return n - 1
	}
	return n
}

// carvePassages runs a randomized depth-first search over the odd cells,
// producing a spanning tree of corridors.
func (m *Maze) carvePassages(rng *rand.Rand) {
	start := Point{1, 1}
	m.open(start)
	stack := []Point{start}
	for len(stack) > 0 {
		current := stack[len(stack)-1]
		dir, ok := m.randomUnvisitedDir(current, rng)
		if !ok {
			stack = stack[:len(stack)-1]
			continue
		}
		dx, dy := dir.Delta()
		next := Point{current.X + 2*dx, current.Y + 2*dy}
		m.open(Point{current.X + dx, current.Y + dy})
		m.open(next)
		stack = append(stack, next)
	}
}

func (m *Maze) randomUnvisitedDir(p Point, rng *rand.Rand) (Dir, bool) {
	for _, i := range rng.Perm(len(Dirs)) {
		dx, dy := Dirs[i].Delta()
		next := Point{p.X + 2*dx, p.Y + 2*dy}
		if m.isInterior(next) && m.IsWall(next) {
			return Dirs[i], true
		}
	}
	return 0, false
}

// removeDeadEnds opens a wall at every corridor end so that chases never get
// trapped. Dead ends can only sit on odd cells, the nodes of the spanning tree.
func (m *Maze) removeDeadEnds(rng *rand.Rand) {
	for y := 1; y < m.H; y += 2 {
		for x := 1; x < m.W; x += 2 {
			p := Point{x, y}
			if len(m.Exits(p)) >= 2 {
				continue
			}
			m.openTowardsNeighbor(p, rng)
		}
	}
}

func (m *Maze) openTowardsNeighbor(p Point, rng *rand.Rand) {
	for _, i := range rng.Perm(len(Dirs)) {
		dx, dy := Dirs[i].Delta()
		wall := Point{p.X + dx, p.Y + dy}
		if m.isInterior(Point{p.X + 2*dx, p.Y + 2*dy}) && m.IsWall(wall) {
			m.open(wall)
			return
		}
	}
}

// openExtraLoops removes a few walls between odd cells to add escape routes.
func (m *Maze) openExtraLoops(rng *rand.Rand) {
	for y := 1; y < m.H-1; y++ {
		for x := 1; x < m.W-1; x++ {
			separatesCells := (x%2 == 0) != (y%2 == 0)
			if separatesCells && rng.Float64() < extraLoopChance {
				m.open(Point{x, y})
			}
		}
	}
}

func (m *Maze) carveDen() {
	for y := m.Den.Y - denHalfHeight; y <= m.Den.Y+denHalfHeight; y++ {
		for x := m.Den.X - denHalfWidth; x <= m.Den.X+denHalfWidth; x++ {
			m.open(Point{x, y})
		}
	}
}

// openTunnels opens both side walls on random odd rows. Odd rows always have
// floor at the columns next to the border, so tunnels connect to the maze.
func (m *Maze) openTunnels(rng *rand.Rand) {
	oddRows := (m.H - 1) / 2
	for _, i := range rng.Perm(oddRows)[:tunnelCount] {
		y := 2*i + 1
		m.open(Point{0, y})
		m.open(Point{m.W - 1, y})
	}
}

func (m *Maze) open(p Point) { m.walls[m.Index(p)] = false }

func (m *Maze) isInterior(p Point) bool {
	return p.X > 0 && p.X < m.W-1 && p.Y > 0 && p.Y < m.H-1
}

// Index maps a point to its offset in per-cell slices such as those returned by
// Distances.
func (m *Maze) Index(p Point) int { return p.Y*m.W + p.X }

// IsWall reports whether p is a wall.
func (m *Maze) IsWall(p Point) bool { return m.walls[m.Index(p)] }

// InDen reports whether p lies inside the central den.
func (m *Maze) InDen(p Point) bool {
	return abs(p.X-m.Den.X) <= denHalfWidth && abs(p.Y-m.Den.Y) <= denHalfHeight
}

// Step returns the cell one move away from p, wrapping through side tunnels. It
// reports false when the move is blocked.
func (m *Maze) Step(p Point, d Dir) (Point, bool) {
	dx, dy := d.Delta()
	next := Point{(p.X + dx + m.W) % m.W, p.Y + dy}
	if next.Y < 0 || next.Y >= m.H || m.IsWall(next) {
		return Point{}, false
	}
	return next, true
}

// Exits lists the directions that can be taken from p.
func (m *Maze) Exits(p Point) []Dir {
	exits := make([]Dir, 0, len(Dirs))
	for _, d := range Dirs {
		if _, ok := m.Step(p, d); ok {
			exits = append(exits, d)
		}
	}
	return exits
}

// Distances returns, for every cell, the number of steps to the closest source,
// or Unreachable.
func (m *Maze) Distances(sources ...Point) []int {
	dist := make([]int, m.W*m.H)
	for i := range dist {
		dist[i] = Unreachable
	}
	queue := make([]Point, 0, len(dist))
	for _, s := range sources {
		dist[m.Index(s)] = 0
		queue = append(queue, s)
	}
	for head := 0; head < len(queue); head++ {
		current := queue[head]
		for _, d := range Dirs {
			next, ok := m.Step(current, d)
			if ok && dist[m.Index(next)] == Unreachable {
				dist[m.Index(next)] = dist[m.Index(current)] + 1
				queue = append(queue, next)
			}
		}
	}
	return dist
}

// FirstStepTo runs a breadth-first search from start and returns the first move
// on a shortest path to the closest cell accepted by goal. The search only enters
// cells accepted by passable, which also receives the steps taken to reach them.
func (m *Maze) FirstStepTo(start Point, passable func(p Point, steps int) bool, goal func(Point) bool) (Dir, bool) {
	type node struct {
		pos   Point
		first Dir
		steps int
	}
	seen := make([]bool, m.W*m.H)
	seen[m.Index(start)] = true
	queue := make([]node, 0, 64)
	queue = append(queue, node{pos: start})
	for head := 0; head < len(queue); head++ {
		current := queue[head]
		for _, d := range Dirs {
			next, ok := m.Step(current.pos, d)
			if !ok || seen[m.Index(next)] || !passable(next, current.steps+1) {
				continue
			}
			first := current.first
			if current.steps == 0 {
				first = d
			}
			if goal(next) {
				return first, true
			}
			seen[m.Index(next)] = true
			queue = append(queue, node{next, first, current.steps + 1})
		}
	}
	return 0, false
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
