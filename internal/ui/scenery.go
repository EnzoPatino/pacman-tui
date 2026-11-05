package ui

import (
	"github.com/EnzoPatino/pacman-tui/internal/game"
	"github.com/EnzoPatino/pacman-tui/internal/maze"
)

const (
	sparkGlyph      = '·'
	largeSparkGlyph = '·'
	twinkleGlyph    = '✧'
	nectarGlyph     = '●'
	nectarAltGlyph  = '◉'
	denPatternGlyph = ' '
	trailGlyph      = '·'

	twinklePeriod     = 47 // on average, one spark in this many is twinkling
	twinklePhaseTicks = 2
	nectarBloomTicks  = 4 // ticks between the nectar's two flower shapes
	nectarPulseTicks  = 2 // slows the nectar's pulse relative to the firefly's
	nectarMinBright   = 0.45
	trailLength       = 7
	trailBrightAge    = 2
	wallColorBands    = 16

	progressBarFull   = '━'
	progressBarEmpty  = '─'
	progressBarMin    = 6
	progressBarMax    = 24
	progressBarColumn = 5 // one bar column for every this many terminal columns
)

// The bits of a wall's connections to its neighbors.
const (
	linkUp = 1 << iota
	linkRight
	linkDown
	linkLeft
)

// wallJoints maps connections to the box-drawing character that joins them.
var wallJoints = [16]rune{
	0:                                        '█',
	linkUp:                                   '╹',
	linkRight:                                '╺',
	linkDown:                                 '╻',
	linkLeft:                                 '╸',
	linkUp | linkDown:                        '┃',
	linkLeft | linkRight:                     '━',
	linkUp | linkRight:                       '╰',
	linkRight | linkDown:                     '╭',
	linkDown | linkLeft:                      '╮',
	linkUp | linkLeft:                        '╯',
	linkUp | linkRight | linkDown:            '┣',
	linkRight | linkDown | linkLeft:          '┳',
	linkUp | linkDown | linkLeft:             '┫',
	linkUp | linkRight | linkLeft:            '┻',
	linkUp | linkRight | linkDown | linkLeft: '╋',
}

// wallLinks reports which neighbors of a wall cell are walls too.
func wallLinks(m *maze.Maze, p maze.Point) int {
	links := 0
	isWall := func(x, y int) bool {
		return x >= 0 && y >= 0 && x < m.W && y < m.H && m.IsWall(maze.Point{X: x, Y: y})
	}
	if isWall(p.X, p.Y-1) {
		links |= linkUp
	}
	if isWall(p.X+1, p.Y) {
		links |= linkRight
	}
	if isWall(p.X, p.Y+1) {
		links |= linkDown
	}
	if isWall(p.X-1, p.Y) {
		links |= linkLeft
	}
	return links
}

// paintWall draws walls as thin lines from one cell's center to the next, which
// leaves corridors wide enough for the large sprites.
func (c canvas) paintWall(p maze.Point) {
	x, y := c.position(p)
	style := c.wallColor(p).fg().Bold(true)
	links := wallLinks(c.g.Maze, p)
	c.f.set(x, y, wallJoints[links], style)
	if links&linkRight != 0 {
		for i := 1; i < c.scale.cellW; i++ {
			c.f.set(x+i, y, '━', style)
		}
	}
	if links&linkDown != 0 {
		for i := 1; i < c.scale.cellH; i++ {
			c.f.set(x, y+i, '┃', style)
		}
	}
}

// wallColor keeps the classic saturated blue maze walls.
func (c canvas) wallColor(p maze.Point) rgb {
	return rgb{70, 110, 255}
}

func (c canvas) paintFloor(p maze.Point) {
	x, y := c.position(p)
	tick := c.g.Tick()
	switch c.g.ItemAt(p) {
	case game.Spark:
		glyph, color := spark(p, tick, lightAt(c.g, p), c.scale.large)
		c.f.set(x, y, glyph, color.fg())
	case game.Nectar:
		glyph, color := nectar(tick, lightAt(c.g, p))
		style := color.fg().Bold(true)
		c.f.set(x, y, glyph, style)
	default:
		if age, visited := c.g.TicksSinceVisit(p); visited && age < trailLength {
			glyph, color := trailSpark(age)
			c.f.set(x, y, glyph, color.fg())
		} else if c.g.Maze.InDen(p) && isDenPattern(p) {
			c.f.set(x, y, denPatternGlyph, denPattern.fg())
		}
	}
}

// cellHash gives every cell a stable pseudo-random value, so that sparks twinkle
// at different times.
func cellHash(p maze.Point) uint32 {
	h := uint32(p.X)*73856093 ^ uint32(p.Y)*19349663
	h ^= h >> 13
	h *= 1274126177
	return h ^ h>>16
}

func spark(p maze.Point, tick int, light float64, large bool) (rune, rgb) {
	if (cellHash(p)+uint32(tick/twinklePhaseTicks))%twinklePeriod == 0 {
		return twinkleGlyph, sparkLit
	}
	return sparkGlyph, sparkLit
}

func trailSpark(age int) (rune, rgb) {
	color := fireflyGlow.mix(fireflyBody.mix(black, 0.6), float64(age)/trailLength)
	if age <= trailBrightAge {
		return trailGlyph, color
	}
	return sparkGlyph, color
}

// nectar always stands out, unlike sparks, because the firefly's life depends on it.
func nectar(tick int, light float64) (rune, rgb) {
	if (tick/nectarBloomTicks)%2 == 1 {
		return nectarAltGlyph, nectarLit
	}
	return nectarGlyph, nectarLit
}

func isDenPattern(p maze.Point) bool { return (p.X+p.Y)%2 == 0 }

func progressBarWidth(columns int) int {
	return min(max(columns/progressBarColumn, progressBarMin), progressBarMax)
}

// progressBar returns the filled and empty parts of a bar. A fraction above zero
// always shows at least one filled column, so progress is never invisible.
func progressBar(fraction float64, width int) (filled, empty string) {
	count := int(fraction * float64(width))
	if fraction > 0 && count == 0 {
		count = 1
	}
	count = min(count, width)
	return repeatRune(progressBarFull, count), repeatRune(progressBarEmpty, width-count)
}

func repeatRune(r rune, count int) string {
	runes := make([]rune, count)
	for i := range runes {
		runes[i] = r
	}
	return string(runes)
}

func barSegments(fraction float64, width int, color rgb) []segment {
	filled, empty := progressBar(fraction, width)
	return []segment{
		{filled, color.fg().Bold(true)},
		{empty, barEmpty.fg()},
	}
}

func statusFooter(g *game.Game) string {
	const controls = "PAC-MAN  ·  q / esc para salir"
	switch {
	case g.GameOver():
		return "GAME OVER"
	case g.Caught:
		return "CAUGHT"
	case g.SparksLeft() == 0:
		return "MAZE CLEARED  ·  nuevo nivel"
	case g.GlowTicks() > 0:
		return "PODER ACTIVO  ·  " + controls
	default:
		return controls
	}
}
