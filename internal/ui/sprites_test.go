package ui

import (
	"testing"

	"github.com/EnzoPatino/pacman-tui/internal/game"
	"github.com/EnzoPatino/pacman-tui/internal/maze"
)

func TestPersonalitiesHaveDifferentFaces(t *testing.T) {
	const awakeTick = 20 // no robot blinks on this tick
	faces := map[pair]bool{}
	for _, p := range []game.Personality{game.Hunter, game.Ambusher, game.Drifter} {
		faces[robotFace(&game.Robot{Personality: p}, awakeTick)] = true
	}
	if len(faces) != 3 {
		t.Fatalf("got %d distinct faces, want 3", len(faces))
	}
}

func TestScaredRobotsLookScaredWhateverTheirPersonality(t *testing.T) {
	hunter := robotFace(&game.Robot{Personality: game.Hunter, Scared: true}, 0)
	drifter := robotFace(&game.Robot{Personality: game.Drifter, Scared: true}, 0)
	if hunter != drifter {
		t.Fatalf("scared faces differ: %q, %q", hunter, drifter)
	}
}

func TestSpritesHaveTheSizeTheirScaleLeavesRoomFor(t *testing.T) {
	face := pair{'o', 'o'}
	sizes := []struct {
		name   string
		sprite sprite
		w, h   int
	}{
		{"large robot", robotSprite(face, 0, true), 5, 3},
		{"large robot, other pose", robotSprite(face, legPhaseTicks, true), 5, 3},
		{"small robot", robotSprite(face, 0, false), 3, 1},
		{"large firefly", fireflySprite(maze.Left, 0, true), 5, 3},
		{"small firefly", fireflySprite(maze.Left, 0, false), 3, 1},
		{"large caught", caughtSprite(true), 5, 3},
		{"small caught", caughtSprite(false), 3, 1},
	}
	for _, s := range sizes {
		if len(s.sprite.rows) != s.h || len(s.sprite.mask) != s.h {
			t.Errorf("%s: %d rows and %d mask rows, want %d", s.name, len(s.sprite.rows), len(s.sprite.mask), s.h)
		}
		for i, row := range s.sprite.rows {
			if got := len([]rune(row)); got != s.w {
				t.Errorf("%s row %d: width %d, want %d", s.name, i, got, s.w)
			}
			if got := len([]rune(s.sprite.mask[i])); got != s.w {
				t.Errorf("%s mask %d: width %d, want %d", s.name, i, got, s.w)
			}
		}
	}
}

func TestFireflyLanternMovesTowardItsHeading(t *testing.T) {
	want := map[maze.Dir][2]int{
		maze.Up:    {0, 2},
		maze.Right: {1, 3},
		maze.Down:  {2, 2},
		maze.Left:  {1, 1},
	}
	for dir, cell := range want {
		s := fireflySprite(dir, 0, true)
		if got := []rune(s.rows[cell[0]])[cell[1]]; got != fireflyGlyph {
			t.Errorf("heading %v: got %q at %v, want the lantern", dir, got, cell)
		}
	}
}

func TestSpritesAnimate(t *testing.T) {
	face := pair{'o', 'o'}
	if robotSprite(face, 0, true).rows[1] == robotSprite(face, legPhaseTicks, true).rows[1] {
		t.Error("robot arms do not move")
	}
	if fireflySprite(maze.Left, 0, true).rows[1] == fireflySprite(maze.Left, wingPhaseTicks, true).rows[1] {
		t.Error("firefly wings do not beat")
	}
}
