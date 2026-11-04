package ui

import (
	"fmt"
	"math"

	"github.com/gdamore/tcell/v2"

	"github.com/EnzoPatino/pacman-tui/internal/game"
	"github.com/EnzoPatino/pacman-tui/internal/maze"
)

const (
	wallGlyph    = '▒'
	sparkGlyph   = '·'
	nectarGlyph  = '✿'
	fireflyGlyph = '✦'
	caughtGlyph  = '✕'
	spiderGlyph  = 'ж'

	ambientLight  = 0.18
	lightRadius   = 9.0
	glowRadius    = 16.0
	lightLevels   = 5.0
	glowWarnTicks = 15 // scared spiders flicker when the glow is about to end
)

type rgb struct{ r, g, b float64 }

func (c rgb) mix(other rgb, amount float64) rgb {
	return rgb{
		c.r + (other.r-c.r)*amount,
		c.g + (other.g-c.g)*amount,
		c.b + (other.b-c.b)*amount,
	}
}

func (c rgb) style() tcell.Style {
	return tcell.StyleDefault.Foreground(tcell.NewRGBColor(int32(c.r), int32(c.g), int32(c.b)))
}

// The night-forest palette: objects are dim until the firefly's light hits them.
var (
	mossDark    = rgb{38, 58, 48}
	mossLit     = rgb{112, 190, 120}
	sparkDark   = rgb{92, 86, 50}
	sparkLit    = rgb{250, 225, 130}
	nectarDark  = rgb{120, 62, 104}
	nectarLit   = rgb{255, 140, 205}
	fireflyBody = rgb{214, 255, 110}
	fireflyGlow = rgb{255, 255, 225}
	caughtColor = rgb{255, 90, 80}
	scaredColor = rgb{140, 190, 255}
	warnColor   = rgb{235, 240, 255}
)

var spiderColors = map[game.Personality]rgb{
	game.Hunter:   {238, 86, 86},
	game.Ambusher: {188, 128, 255},
	game.Drifter:  {255, 158, 72},
}

func draw(screen tcell.Screen, g *game.Game) {
	screen.Clear()
	if g == nil {
		drawTooSmallNotice(screen)
	} else {
		drawGame(screen, g)
	}
	screen.Show()
}

func drawTooSmallNotice(screen tcell.Screen) {
	columns, rows := screen.Size()
	message := fmt.Sprintf("Terminal too small (need %dx%d)", maze.MinWidth*cellColumns, maze.MinHeight)
	runes := []rune(message)
	x := max((columns-len(runes))/2, 0)
	for i, r := range runes {
		screen.SetContent(x+i, rows/2, r, nil, tcell.StyleDefault)
	}
}

func drawGame(screen tcell.Screen, g *game.Game) {
	columns, rows := screen.Size()
	originX := (columns - g.Maze.W*cellColumns) / 2
	originY := (rows - g.Maze.H) / 2
	put := func(p maze.Point, glyph rune, style tcell.Style) {
		screen.SetContent(originX+p.X*cellColumns, originY+p.Y, glyph, nil, style)
	}

	for y := range g.Maze.H {
		for x := range g.Maze.W {
			p := maze.Point{X: x, Y: y}
			light := lightAt(g, p)
			switch {
			case g.Maze.IsWall(p):
				style := mossDark.mix(mossLit, light).style()
				put(p, wallGlyph, style)
				screen.SetContent(originX+x*cellColumns+1, originY+y, wallGlyph, nil, style)
			case g.ItemAt(p) == game.Spark:
				put(p, sparkGlyph, sparkDark.mix(sparkLit, light).style())
			case g.ItemAt(p) == game.Nectar:
				put(p, nectarGlyph, nectarDark.mix(nectarLit, light).style())
			}
		}
	}
	for _, s := range g.Spiders {
		put(s.Pos, spiderGlyph, spiderColor(g, s).style().Bold(true))
	}
	if g.Caught {
		put(g.Firefly, caughtGlyph, caughtColor.style().Bold(true))
		return
	}
	body := fireflyBody
	if g.GlowTicks() > 0 {
		body = fireflyGlow
	}
	put(g.Firefly, fireflyGlyph, body.style().Bold(true))
}

// lightAt returns how strongly the firefly illuminates a cell, from 0 to 1.
func lightAt(g *game.Game, p maze.Point) float64 {
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

func spiderColor(g *game.Game, s *game.Spider) rgb {
	if !s.Scared {
		return spiderColors[s.Personality]
	}
	if ticks := g.GlowTicks(); ticks < glowWarnTicks && ticks%4 < 2 {
		return warnColor
	}
	return scaredColor
}
