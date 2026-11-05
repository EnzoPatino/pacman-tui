package ui

import "github.com/gdamore/tcell/v2"

type glyph struct {
	r     rune
	style tcell.Style
}

// frame is an off-screen picture of the terminal. Painting here instead of on the
// screen lets later layers overwrite earlier ones freely: only the final result of
// each cell is compared with the previous frame and sent to the terminal.
type frame struct {
	w, h  int
	cells []glyph
}

func (f *frame) reset(w, h int) {
	f.w, f.h = w, h
	if len(f.cells) != w*h {
		f.cells = make([]glyph, w*h)
	}
	for i := range f.cells {
		f.cells[i] = glyph{' ', tcell.StyleDefault}
	}
}

// set ignores positions outside the frame, so callers need not clip.
func (f *frame) set(x, y int, r rune, style tcell.Style) {
	if x >= 0 && y >= 0 && x < f.w && y < f.h {
		f.cells[y*f.w+x] = glyph{r, style}
	}
}

// segment is a run of text in one style within a line.
type segment struct {
	text  string
	style tcell.Style
}

// centered writes a line in the middle of a row.
func (f *frame) centered(y int, segments ...segment) {
	length := 0
	for _, s := range segments {
		length += len([]rune(s.text))
	}
	x := max((f.w-length)/2, 0)
	for _, s := range segments {
		for _, r := range s.text {
			f.set(x, y, r, s.style)
			x++
		}
	}
}
