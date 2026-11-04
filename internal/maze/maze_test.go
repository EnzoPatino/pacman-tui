package maze

import (
	"math/rand/v2"
	"testing"
)

var testSizes = [][2]int{{MinWidth, MinHeight}, {16, 12}, {41, 21}, {101, 61}}

func forEachMaze(t *testing.T, check func(t *testing.T, m *Maze)) {
	t.Helper()
	for _, size := range testSizes {
		for seed := range uint64(25) {
			check(t, Generate(size[0], size[1], rand.New(rand.NewPCG(seed, 1))))
		}
	}
}

func floorCells(m *Maze) []Point {
	var cells []Point
	for y := range m.H {
		for x := range m.W {
			if p := (Point{x, y}); !m.IsWall(p) {
				cells = append(cells, p)
			}
		}
	}
	return cells
}

func TestEveryFloorCellIsReachableFromTheDen(t *testing.T) {
	forEachMaze(t, func(t *testing.T, m *Maze) {
		dist := m.Distances(m.Den)
		for _, p := range floorCells(m) {
			if dist[m.Index(p)] == Unreachable {
				t.Fatalf("%dx%d: cell %v is unreachable", m.W, m.H, p)
			}
		}
	})
}

func TestThereAreNoDeadEnds(t *testing.T) {
	forEachMaze(t, func(t *testing.T, m *Maze) {
		for _, p := range floorCells(m) {
			if n := len(m.Exits(p)); n < 2 {
				t.Fatalf("%dx%d: cell %v has %d exits", m.W, m.H, p, n)
			}
		}
	})
}

func TestTunnelsWrapAroundAndOtherBordersAreClosed(t *testing.T) {
	forEachMaze(t, func(t *testing.T, m *Maze) {
		tunnels := 0
		for y := range m.H {
			left, right := Point{0, y}, Point{m.W - 1, y}
			if m.IsWall(left) != m.IsWall(right) {
				t.Fatalf("%dx%d: row %d has a one-sided tunnel", m.W, m.H, y)
			}
			if m.IsWall(left) {
				continue
			}
			tunnels++
			if next, ok := m.Step(left, Left); !ok || next != right {
				t.Fatalf("%dx%d: row %d does not wrap: got %v, %v", m.W, m.H, y, next, ok)
			}
		}
		if tunnels != tunnelCount {
			t.Fatalf("%dx%d: got %d tunnels, want %d", m.W, m.H, tunnels, tunnelCount)
		}
		for x := range m.W {
			if !m.IsWall(Point{x, 0}) || !m.IsWall(Point{x, m.H - 1}) {
				t.Fatalf("%dx%d: top or bottom border is open at column %d", m.W, m.H, x)
			}
		}
	})
}

func TestEvenDimensionsAreReducedToOdd(t *testing.T) {
	m := Generate(16, 12, rand.New(rand.NewPCG(1, 1)))
	if m.W != 15 || m.H != 11 {
		t.Fatalf("got %dx%d, want 15x11", m.W, m.H)
	}
}

func TestFirstStepToFollowsShortestPathAndRespectsPassable(t *testing.T) {
	m := Generate(41, 21, rand.New(rand.NewPCG(3, 1)))
	start := Point{1, 1}
	target := Point{m.W - 2, m.H - 2}
	dist := m.Distances(target)
	anywhere := func(Point, int) bool { return true }
	isTarget := func(p Point) bool { return p == target }

	dir, ok := m.FirstStepTo(start, anywhere, isTarget)
	if !ok {
		t.Fatal("no path found")
	}
	next, ok := m.Step(start, dir)
	if !ok || dist[m.Index(next)] != dist[m.Index(start)]-1 {
		t.Fatalf("first step %v does not get closer to the target", dir)
	}

	nowhere := func(Point, int) bool { return false }
	if _, ok := m.FirstStepTo(start, nowhere, isTarget); ok {
		t.Fatal("found a path through impassable cells")
	}
}
