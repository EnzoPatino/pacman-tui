package ui

import (
	"fmt"
	"math"

	"github.com/gdamore/tcell/v2"

	"github.com/EnzoPatino/pacman-tui/internal/game"
	"github.com/EnzoPatino/pacman-tui/internal/maze"
)

const (
	hudRows    = 3 // title, status and a blank line
	footerRows = 2 // a blank line and the footer

	maxMazeWidth  = 81
	maxMazeHeight = 41

	titleRow  = 0
	statusRow = 1

	wideStatusColumns = 90
	statusBarWidth    = 20

	ambientLight  = 0.72
	lightRadius   = 10.0
	glowRadius    = 17.0
	lightLevels   = 6.0
	glowWarnTicks = 15 // scared robots flicker when the glow is about to end
)

// scale is how large the maze and its characters are drawn. Terminal cells are
// about twice as tall as they are wide, so both scales look square.
type scale struct {
	cellW, cellH int
	marginX      int // room for a sprite inside a side tunnel
	large        bool
}

var (
	largeScale = scale{cellW: 4, cellH: 2, marginX: 3, large: true}
	smallScale = scale{cellW: 2, cellH: 1, marginX: 2}
)

// fit is a scale and the maze size that fills a terminal with it.
type fit struct {
	scale
	mazeW, mazeH int
}

func (s scale) fitIn(columns, rows int) (fit, bool) {
	drawableW := columns - 2*s.marginX - 1
	drawableH := rows - hudRows - footerRows - 1
	if drawableW < 0 || drawableH < 0 {
		return fit{}, false
	}
	w := min(drawableW/s.cellW+1, maxMazeWidth)
	h := min(drawableH/s.cellH+1, maxMazeHeight)
	return fit{s, w, h}, w >= maze.MinWidth && h >= maze.MinHeight
}

// fitTerminal picks the largest scale at which a whole maze fits.
func fitTerminal(columns, rows int) (fit, bool) {
	if f, ok := largeScale.fitIn(columns, rows); ok {
		return f, true
	}
	return smallScale.fitIn(columns, rows)
}

type rgb struct{ r, g, b float64 }

func (c rgb) mix(other rgb, amount float64) rgb {
	return rgb{
		c.r + (other.r-c.r)*amount,
		c.g + (other.g-c.g)*amount,
		c.b + (other.b-c.b)*amount,
	}
}

func (c rgb) styleColor() tcell.Color {
	return tcell.NewRGBColor(int32(c.r), int32(c.g), int32(c.b))
}

func (c rgb) fg() tcell.Style {
	return tcell.StyleDefault.Foreground(c.styleColor())
}

// The arcade palette uses blue walls, white dots, yellow power pellets and
// distinct ghost colors.
var (
	black       = rgb{0, 0, 0}
	neonMagenta = rgb{35, 75, 255}
	neonViolet  = rgb{35, 75, 255}
	sparkDim    = rgb{245, 245, 245}
	sparkLit    = rgb{255, 255, 255}
	nectarDim   = rgb{255, 200, 30}
	nectarLit   = rgb{255, 225, 45}
	denPattern  = rgb{35, 75, 255}
	fireflyBody = rgb{255, 220, 0}
	fireflyGlow = rgb{255, 235, 40}
	caughtColor = rgb{255, 220, 0}
	scaredColor = rgb{35, 90, 255}
	warnColor   = rgb{255, 255, 255}
	mutedColor  = rgb{185, 185, 185}
	titleColor  = rgb{255, 220, 0}
	barSparks   = rgb{255, 255, 255}
	barGlow     = rgb{35, 90, 255}
	barEmpty    = rgb{70, 70, 70}
)

var robotColors = map[game.Personality]rgb{
	game.Hunter:   {255, 35, 35},
	game.Ambusher: {255, 150, 205},
	game.Drifter:  {0, 230, 255},
}

// renderer draws frames and sends only what changed to the screen.
type renderer struct {
	screen  tcell.Screen
	current frame
	shown   frame
}

func newRenderer(screen tcell.Screen) *renderer { return &renderer{screen: screen} }

// redraw repaints a screen whose layout changed, leaving nothing of the old one.
func (r *renderer) redraw(g *game.Game) {
	r.shown = frame{}
	r.draw(g)
	r.screen.Sync()
}

func (r *renderer) draw(g *game.Game) {
	columns, rows := r.screen.Size()
	r.current.reset(columns, rows)
	if g == nil {
		paintTooSmallNotice(&r.current)
	} else {
		paintGame(&r.current, g)
	}
	r.flush()
	r.screen.Show()
}

// flush avoids writing unchanged cells: tcell's own comparison is defeated by
// clearing, and every write has a cost on a screensaver that runs for hours.
func (r *renderer) flush() {
	cur, prev := &r.current, &r.shown
	sameSize := prev.w == cur.w && prev.h == cur.h && len(prev.cells) == len(cur.cells)
	for i, cell := range cur.cells {
		if !sameSize || cell != prev.cells[i] {
			r.screen.SetContent(i%cur.w, i/cur.w, cell.r, nil, cell.style)
		}
	}
	r.current, r.shown = r.shown, r.current
}

func paintTooSmallNotice(f *frame) {
	smallest := smallScale
	minColumns := (maze.MinWidth-1)*smallest.cellW + 1 + 2*smallest.marginX
	minRows := (maze.MinHeight-1)*smallest.cellH + 1 + hudRows + footerRows
	message := fmt.Sprintf("Terminal too small (need at least %dx%d)", minColumns, minRows)
	f.centered(f.h/2, segment{message, mutedColor.fg()})
}

func paintGame(f *frame, g *game.Game) {
	fitted, _ := fitTerminal(f.w, f.h)
	c := newCanvas(f, g, fitted.scale)
	c.paintHeader()
	c.paintMaze()
	for _, s := range g.Robots {
		c.paintRobot(s)
	}
	c.paintFirefly()
	c.paintBanner()
	f.centered(f.h-1, segment{statusFooter(g), mutedColor.fg()})
}

func (c canvas) paintHeader() {
	titleStyle := titleColor.fg().Bold(true)
	c.f.centered(titleRow, segment{"P A C - M A N", titleStyle})

	c.f.centered(statusRow, c.statusLine()...)
}

// statusLine always shows the score and lives. Only wide terminals have
// room for the high score and the progress bars.
func (c canvas) statusLine() []segment {
	muted := mutedColor.fg()
	value := titleColor.fg().Bold(true)
	wide := c.f.w >= wideStatusColumns
	label := "1UP "
	if !wide {
		label = "SCORE "
	}
	status := []segment{{label, muted}, {fmt.Sprintf("%06d", c.g.Score()), value}}
	if wide {
		status = append(status, segment{"   HI ", muted}, segment{fmt.Sprintf("%06d", c.g.HighScore()), value})
	}
	status = append(status,
		segment{"   LIVES ", muted}, segment{repeatRune('●', c.g.Lives()), fireflyBody.fg().Bold(true)},
	)
	if !wide {
		return status
	}
	status = append(status, segment{"   ", muted})
	status = append(status, barSegments(c.g.SparkProgress(), statusBarWidth, barSparks)...)
	if c.g.GlowTicks() > 0 {
		status = append(status, segment{"   POWER ", muted})
		status = append(status, barSegments(c.g.GlowFraction(), statusBarWidth/2, barGlow)...)
	}
	return status
}

// paintBanner announces a new level or the end of a game over the maze's center.
func (c canvas) paintBanner() {
	text := ""
	switch {
	case c.g.GameOver():
		text = "G A M E   O V E R"
	case c.g.Ready():
		text = fmt.Sprintf("READY!   L E V E L   %d", c.g.Level())
	default:
		return
	}
	_, row := c.position(maze.Point{X: 0, Y: c.g.Maze.H / 2})
	style := tcell.StyleDefault.Background(spriteInk.styleColor()).Foreground(titleColor.styleColor()).Bold(true)
	c.f.centered(row, segment{"   " + text + "   ", style})
}

func (c canvas) paintMaze() {
	m := c.g.Maze
	for y := range m.H {
		for x := range m.W {
			p := maze.Point{X: x, Y: y}
			if m.IsWall(p) {
				c.paintWall(p)
			} else {
				c.paintFloor(p)
			}
		}
	}
}

// lightAt returns how strongly the firefly brightens a cell, from 0 to 1.
func lightAt(g *game.Game, p maze.Point) float64 {
	if g.SparksLeft() == 0 {
		return 1 // the whole maze flares up when it is cleared
	}
	radius := lightRadius
	if g.GlowTicks() > 0 {
		radius = glowRadius
	}
	distance := math.Hypot(float64(p.X-g.Firefly.X), float64(p.Y-g.Firefly.Y))
	falloff := max(1-distance/radius, 0)
	light := ambientLight + (1-ambientLight)*falloff*falloff
	// Discrete bands let more cells keep their color between frames, which
	// reduces what has to be sent to the terminal.
	return math.Round(light*lightLevels) / lightLevels
}

func robotColor(g *game.Game, s *game.Robot) rgb {
	if !s.Scared {
		return robotColors[s.Personality]
	}
	if ticks := g.GlowTicks(); ticks < glowWarnTicks && ticks%4 < 2 {
		return warnColor
	}
	return scaredColor
}
