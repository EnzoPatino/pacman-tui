package ui

import (
	"math"

	"github.com/gdamore/tcell/v2"

	"github.com/EnzoPatino/pacman-tui/internal/game"
	"github.com/EnzoPatino/pacman-tui/internal/maze"
)

const (
	fireflyGlyph = 'C'

	legPhaseTicks  = 2 // ticks between leg positions
	wingPhaseTicks = 1 // ticks between wing beats
	eyePhaseTicks  = 4 // ticks between the drifter's eye movements
	pulsePeriod    = 6 // ticks in one brightness cycle of the firefly
	blinkPeriod    = 37
	blinkTicks     = 2
	blinkOffset    = 11 // keeps robots from blinking at the same time

	bodyMask = 'b' // a sprite cell drawn as ink on a solid body color
)

var spriteInk = rgb{14, 10, 22}

// sprite is a small picture centered on a maze cell. Spaces are transparent, except
// where mask marks a cell as part of the solid body.
type sprite struct {
	rows []string
	mask []string
}

// canvas paints one game on a frame at one scale.
type canvas struct {
	f                *frame
	g                *game.Game
	scale            scale
	originX, originY int
}

func newCanvas(f *frame, g *game.Game, s scale) canvas {
	drawnW := (g.Maze.W-1)*s.cellW + 1
	drawnH := (g.Maze.H-1)*s.cellH + 1
	return canvas{
		f:       f,
		g:       g,
		scale:   s,
		originX: (f.w - drawnW) / 2,
		originY: hudRows + (f.h-hudRows-footerRows-drawnH)/2,
	}
}

// position returns the terminal position of a cell's center.
func (c canvas) position(p maze.Point) (x, y int) {
	return c.originX + p.X*c.scale.cellW, c.originY + p.Y*c.scale.cellH
}

func (c canvas) paintSprite(p maze.Point, s sprite, color rgb) {
	centerX, centerY := c.position(p)
	top := centerY - len(s.rows)/2
	limb := color.fg().Bold(true)
	body := tcell.StyleDefault.Background(color.styleColor()).Foreground(spriteInk.styleColor()).Bold(true)
	for rowIndex, row := range s.rows {
		runes, mask := []rune(row), []rune(s.mask[rowIndex])
		left := centerX - len(runes)/2
		for i, r := range runes {
			switch {
			case mask[i] == bodyMask:
				c.f.set(left+i, top+rowIndex, r, body)
			case r != ' ':
				c.f.set(left+i, top+rowIndex, r, limb)
			}
		}
	}
}

func (c canvas) paintRobot(s *game.Robot) {
	tick := c.g.Tick()
	color := robotColor(c.g, s)
	c.paintSprite(s.Pos, robotSprite(robotFace(s, tick), tick, c.scale.large), color)
	if c.scale.large {
		x, y := c.position(s.Pos)
		eyeColor := rgb{255, 255, 255}.fg().Bold(true)
		c.f.set(x-1, y, '●', eyeColor)
		c.f.set(x+1, y, '●', eyeColor)
	}
}

func (c canvas) paintFirefly() {
	tick := c.g.Tick()
	if c.g.Caught {
		c.paintSprite(c.g.Firefly, caughtSprite(c.scale.large), caughtColor)
		return
	}
	c.paintSprite(c.g.Firefly, fireflySprite(c.g.Heading(), tick, c.scale.large), fireflyBody)
}

// pulse returns a brightness from 0 to 1 that rises and falls over pulsePeriod.
func pulse(tick int) float64 {
	return 0.5 + 0.5*math.Sin(2*math.Pi*float64(tick)/pulsePeriod)
}

// pair is a robot's two eyes.
type pair struct{ left, right rune }

// robotFace gives each personality its own eyes, so robots can be told apart
// even on terminals with few colors.
func robotFace(s *game.Robot, tick int) pair {
	if s.Scared {
		if (tick/legPhaseTicks)%2 == 0 {
			return pair{'o', 'o'}
		}
		return pair{'O', 'O'}
	}
	if (tick+int(s.Personality)*blinkOffset)%blinkPeriod < blinkTicks {
		return pair{'-', '-'}
	}
	switch s.Personality {
	case game.Hunter:
		return pair{'ò', 'ó'}
	case game.Ambusher:
		return pair{'◕', '◕'}
	default:
		if (tick/eyePhaseTicks)%2 == 0 {
			return pair{'°', 'o'}
		}
		return pair{'o', '°'}
	}
}

// robotSprite draws a ghost with a rounded cap, eyes and a moving skirt.
func robotSprite(face pair, tick int, large bool) sprite {
	if !large {
		return sprite{rows: []string{" ◉ "}, mask: []string{"   "}}
	}
	feet := "▀▄▀▄▀"
	bodyRow := "█████"
	if (tick/legPhaseTicks)%2 == 1 {
		feet = "▄▀▄▀▄"
		bodyRow = "██ ██"
	}
	return sprite{
		rows: []string{" ▄█▄ ", bodyRow, feet},
		mask: []string{"     ", "     ", "     "},
	}
}

// fireflySprite draws Pac-Man as a round, open mouth facing the direction of travel.
func fireflySprite(heading maze.Dir, tick int, large bool) sprite {
	mouth := fireflyGlyph
	if tick%2 == 1 {
		mouth = 'c'
	}
	if !large {
		return sprite{rows: []string{" " + string(mouth) + " "}, mask: []string{"   "}}
	}
	rows := [3][]rune{[]rune("     "), []rune("     "), []rune("     ")}
	mask := [3][]rune{[]rune("     "), []rune("     "), []rune("     ")}
	// A five by three yellow disk with a wedge cut toward the direction of travel.
	// Background-colored cells make the Pac-Man silhouette legible in any font.
	shape := [3]string{" bbb ", "bb   ", " bbb "} // mouth opens right
	switch heading {
	case maze.Left:
		shape[1] = "   bb"
	case maze.Up:
		shape = [3]string{"bb bb", "bbbbb", " bbb "}
	case maze.Down:
		shape = [3]string{" bbb ", "bbbbb", "bb bb"}
	}
	for y, row := range shape {
		for x, cell := range row {
			if cell == bodyMask {
				mask[y][x] = bodyMask
			}
		}
	}
	rowIndex, col := lanternCell(heading)
	rows[rowIndex][col] = mouth
	return sprite{
		rows: []string{string(rows[0]), string(rows[1]), string(rows[2])},
		mask: []string{string(mask[0]), string(mask[1]), string(mask[2])},
	}
}

func lanternCell(heading maze.Dir) (row, col int) {
	switch heading {
	case maze.Up:
		return 0, 2
	case maze.Right:
		return 1, 3
	case maze.Down:
		return 2, 2
	default:
		return 1, 1
	}
}

func caughtSprite(large bool) sprite {
	if !large {
		return sprite{rows: []string{" ✕ "}, mask: []string{"   "}}
	}
	return sprite{
		rows: []string{"     ", "  ✕  ", "     "},
		mask: []string{"     ", "     ", "     "},
	}
}
