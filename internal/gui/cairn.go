package gui

import (
	"cairn/internal/board"
	"cairn/internal/engine/item"
	"cairn/internal/gui/geom"
	"cairn/internal/gui/motion"
	"cairn/internal/store/files"
	"image"
	"sync"
	"time"
)

const (
	stoneGap    = 0
	stoneWobble = 2
	stoneLeast  = 40
)

// Born is the window told the moment of its first frame, which is when the board starts arriving.
func (w Window) Born(at time.Time) Window {
	if w.born.IsZero() {
		w.born, w.Now = at, at
	}
	return w
}

// LessMotion is the window told whether the system asks for less motion, in which case the board arrives at once.
func (w Window) LessMotion(asked bool) Window {
	w.lessMotion = asked
	return w
}

// Introducing is the window playing the intro named, as the intro chooser plays the one it is on.
func (w Window) Introducing(name string) Window {
	w.introPlayed = name
	return w
}

// IntroShown is the intro the board arrives by: the one the intro chooser is on, the one played, the one the configuration names, or the default.
func (w Window) IntroShown() string {
	switch {
	case w.Screen == board.OnConfiguration && w.IntrosOpen:
		return files.Intros()[w.IntroOn]
	case w.introPlayed != "":
		return w.introPlayed
	case w.Configured.Window != nil && w.Configured.Window.Intro != "":
		return w.Configured.Window.Intro
	}
	return files.DefaultIntro
}

// Intro is how far the board has arrived, from nought to one, and whether it is arriving at all: not before the first frame, not once arrived, not once anything was pressed, not when the system asks for less motion, and not on a board holding no task.
func (w Window) Intro() (float64, bool) {
	if w.born.IsZero() || w.introOver || w.lessMotion || len(w.Showing().Cards) == 0 {
		return 0, false
	}
	progress := float64(w.Now.Sub(w.born)) / float64(motion.IntroLasts(w.IntroShown(), w.Stage()))
	if progress >= 1 {
		return 1, false
	}
	return max(0, progress), true
}

// Scene is the board at this instant of its intro, and whether one is playing; the still board when none is.
func (w Window) Scene() (motion.Scene, bool) {
	if _, on := w.Intro(); !on || w.Tab != board.Tasks {
		return motion.Blank(motion.Stage{}), false
	}
	return motion.Staging(w.IntroShown(), w.Now.Sub(w.born), w.Stage()), true
}

func (w Window) interrupted() Window {
	w.introOver = true
	return w
}

func (w Window) replaying() Window {
	w.born, w.introOver = w.Now, false
	return w
}

// Stage is the board an intro is played on: its columns' and lanes' grounds, every card seen on them in the order the board draws them, the card in focus — or the first in progress, or the first — to start from, and its focus mark.
func (w Window) Stage() motion.Stage {
	l := w.Layout()
	s := motion.Stage{Wide: w.width, High: w.height, Lead: -1}
	dealt, _ := board.Deal(w.Showing().Cards, w.Archived)
	current := -1
	for i, column := range dealt {
		s.Grounds = append(s.Grounds, geom.Box{X: l.Head[i].X, Y: l.Head[i].Y, W: l.Head[i].W, H: l.Head[i].H + l.Column[i].H})
		if column.Name == string(item.Current) {
			current = len(s.Grounds) - 1
		}
	}
	for lane := range board.Stalled {
		s.Grounds = append(s.Grounds, LaneGround(l, lane))
	}
	rows := map[int]int{}
	for _, p := range w.seenPlaced() {
		ground := w.groundOf(l, p.Within)
		staged := motion.Staged{Slot: p.Slot, Ground: ground, Row: rows[ground], Priority: priorityOf(p.Card), Focus: p.Card.InFocus, Letters: len([]rune(p.Card.Title))}
		rows[ground]++
		if staged.Focus && s.Lead < 0 {
			s.Lead = len(s.Cards)
			s.Mark, _ = FocusMark(p.Card, PartsOf(p.Slot).Title)
		}
		s.Cards = append(s.Cards, staged)
	}
	for i, c := range s.Cards {
		if s.Lead < 0 && c.Ground == current {
			s.Lead = i
		}
	}
	if s.Lead < 0 && len(s.Cards) > 0 {
		s.Lead = 0
	}
	return s
}

func (w Window) groundOf(l geom.Layout, within geom.Box) int {
	for i, column := range l.Column {
		if column == within {
			return i
		}
	}
	for lane := range board.Stalled {
		if l.Lane[lane] == within {
			return len(l.Column) + lane
		}
	}
	return 0
}

// Placed is a card where the board draws it: its slot, and the column or lane it is seen through.
type Placed struct {
	Card      board.Card
	Slot      geom.Box
	Within    geom.Box
	Reference bool
	Says      string
}

// Placed is every card the board draws, the columns' then the lanes', in the order they arrive.
func (w Window) Placed() []Placed {
	l := w.Layout()
	var out []Placed
	dealt, _ := board.Deal(w.Showing().Cards, w.Archived)
	for i, column := range dealt {
		for at, c := range column.Cards {
			out = append(out, Placed{Card: c, Slot: l.CardAt(i, at, w.scroll[i]), Within: l.Column[i]})
		}
	}
	for lane := range board.Stalled {
		for at, c := range w.InLane(lane) {
			out = append(out, Placed{Card: c, Slot: l.LaneCardAt(lane, at, w.lanes[lane]), Within: l.Lane[lane]})
		}
	}
	return out
}

func (p Placed) seen() bool {
	return p.Slot.Y < p.Within.Y+p.Within.H && p.Slot.Y+p.Slot.H > p.Within.Y
}

// StonesOf is a card built of stones: courses across it, each cut into stones of their own widths, every corner drawn in a little from its course and every edge given a middle that is not quite on the line — the same stones every time for the same card, and other stones for another.
func StonesOf(panel geom.Box, seed int) [][]geom.Point {
	r := stoneRand(uint32(seed)*2654435761 + 1)
	courses := 2 + r.below(2)
	var stones [][]geom.Point
	top := panel.Y
	for c := range courses {
		bottom := panel.Y + panel.H*(c+1)/courses
		if c < courses-1 {
			bottom += r.between(-panel.H/(4*courses), panel.H/(4*courses))
		}
		left := panel.X
		for left < panel.X+panel.W {
			right := cut(left, panel.X+panel.W, stoneLeast+r.below(max(1, panel.W/2)))
			stones = append(stones, r.stone(geom.Box{X: left, Y: top, W: right - left, H: bottom - top}))
			left = right
		}
		top = bottom
	}
	return stones
}

func cut(left, end, wanted int) int {
	right := min(end, left+wanted)
	if end-right < stoneLeast {
		return end
	}
	return right
}

// Dropped is how far above its place a card's stones are, having fallen so far: from above the window, at nought, to their place, at one.
func Dropped(panel geom.Box, fall float64) int { return -int(float64(panel.Y+panel.H) * (1 - fall)) }

// Fallen is a stone moved down by drop.
func Fallen(stone []geom.Point, drop int) []geom.Point {
	moved := make([]geom.Point, len(stone))
	for i, p := range stone {
		moved[i] = geom.Point{X: p.X, Y: p.Y + drop}
	}
	return moved
}

type stoneRand uint32

func (r *stoneRand) next() uint32 {
	*r = *r*1664525 + 1013904223
	return uint32(*r >> 8)
}

func (r *stoneRand) below(n int) int { return int(r.next() % uint32(max(1, n))) }

func (r *stoneRand) between(lo, hi int) int { return lo + r.below(hi-lo+1) }

func (r *stoneRand) stone(cell geom.Box) []geom.Point {
	in := func() int { return stoneGap + r.below(stoneWobble) }
	x0, x1 := cell.X+in(), cell.X+cell.W-1-in()
	y0, y1 := cell.Y+in(), cell.Y+cell.H-1-in()
	mid := cell.X + cell.W/2 + r.between(-cell.W/6, cell.W/6)
	return []geom.Point{
		{X: x0, Y: y0},
		{X: mid, Y: cell.Y + stoneGap + r.below(stoneWobble)},
		{X: x1, Y: cell.Y + in()},
		{X: cell.X + cell.W - 1 - in(), Y: y1},
		{X: mid + r.between(-cell.W/8, cell.W/8), Y: cell.Y + cell.H - 1 - stoneGap - r.below(stoneWobble)},
		{X: cell.X + in(), Y: cell.Y + cell.H - 1 - in()},
	}
}

const grainSide = 128

// StoneGrain is the grain every stone is painted with over its tone, the same on every run: a mottle a little lighter here and there, and pits a little darker, as a dressed stone has.
var StoneGrain = sync.OnceValue(func() *image.RGBA {
	grain := image.NewRGBA(image.Rect(0, 0, grainSide, grainSide))
	r := stoneRand(2026)
	lattice := make([]float64, (grainSide/grainCell+2)*(grainSide/grainCell+2))
	for i := range lattice {
		lattice[i] = float64(r.below(1000)) / 1000
	}
	across := grainSide/grainCell + 2
	at := func(x, y int) float64 { return lattice[y*across+x] }
	for y := range grainSide {
		for x := range grainSide {
			cx, cy := x/grainCell, y/grainCell
			fx, fy := float64(x%grainCell)/grainCell, float64(y%grainCell)/grainCell
			top := at(cx, cy)*(1-fx) + at(cx+1, cy)*fx
			bottom := at(cx, cy+1)*(1-fx) + at(cx+1, cy+1)*fx
			mottle := top*(1-fy) + bottom*fy
			offset := grain.PixOffset(x, y)
			if r.below(100) < grainPits {
				grain.Pix[offset+3] = uint8(40 + r.below(50))
				continue
			}
			light := uint8(max(0, mottle-0.45) * 120)
			grain.Pix[offset], grain.Pix[offset+1], grain.Pix[offset+2], grain.Pix[offset+3] = light, light, light, light
		}
	}
	return grain
})

const (
	grainCell = 16
	grainPits = 9
)
