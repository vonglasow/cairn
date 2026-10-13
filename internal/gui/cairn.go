package gui

// The window opens on its board being built the way a cairn is, or an old
// wall: each card arrives as stones — irregular, fitted in courses, never
// bricks — falling into its place one card after another, and gives way to
// the card once they have landed. What is built is the content itself;
// nothing is drawn that is not a card, and a board with no task opens on
// nothing being built. It lasts a breath, starts on the first frame a screen
// draws — never in a scenario, which draws none — and the first key or click
// ends it and still does what it does. Drawn rather than shaded: Gio runs no
// shader of a program's own. See docs/gui.md.

import (
	"cairn/internal/engine"
	"time"
)

const (
	// arriveOver is the share of the breath over which the cards start to
	// fall, one after another; falling is how long a card's stones take to
	// land, and settling how long they take to give way to the card.
	arriveOver = 0.45
	falling    = 0.35
	settling   = 0.2
	// stoneGap is what is left between two stones; stoneWobble how far a
	// corner of one is drawn in from its course, at most.
	stoneGap    = 1
	stoneWobble = 4
	// stoneLeast is the narrowest a stone is cut.
	stoneLeast = 48
)

// IntroLasts is how long the board takes to build.
func IntroLasts() time.Duration { return 1800 * time.Millisecond }

// Born is the window told the moment of its first frame, which is when the
// board starts building. Told once: a later frame is not a birth.
func (w Window) Born(at time.Time) Window {
	if w.born.IsZero() {
		w.born, w.now = at, at
	}
	return w
}

// Intro is how far the board is built, from nought to one, and whether it is
// being built at all: not before the first frame, not once built, not once
// anything was pressed, and not on a board holding no task.
func (w Window) Intro() (float64, bool) {
	if w.born.IsZero() || w.introOver || len(w.Showing().Cards) == 0 {
		return 0, false
	}
	progress := float64(w.now.Sub(w.born)) / float64(IntroLasts())
	if progress >= 1 {
		return 1, false
	}
	return max(0, progress), true
}

// interrupted ends the building: somebody has something to do.
func (w Window) interrupted() Window {
	w.introOver = true
	return w
}

// Arriving is how far the k-th of n cards has come at a share of the breath:
// how far its stones have fallen, slowing as they land, and how far they
// have given way to the card since. The first falls first; all have arrived
// by the end.
func Arriving(progress float64, k, n int) (fall, fade float64) {
	start := arriveOver * float64(k) / float64(max(1, n))
	share := min(1, max(0, (progress-start)/falling))
	fall = 1 - (1-share)*(1-share)
	fade = min(1, max(0, (progress-start-falling)/settling))
	return fall, fade
}

// Placed is a card where the board draws it: its slot, and the column or
// lane it is seen through.
type Placed struct {
	Card   engine.Card
	Slot   Box
	Within Box
}

// Placed is every card the board draws, the columns' then the lanes', in
// the order they arrive.
func (w Window) Placed() []Placed {
	l := w.Layout()
	var out []Placed
	dealt, _ := engine.Deal(w.Showing().Cards, w.archived)
	for i, column := range dealt {
		for at, c := range column.Cards {
			out = append(out, Placed{Card: c, Slot: l.CardAt(i, at, w.scroll[i]), Within: l.Column[i]})
		}
	}
	for lane := range engine.Stalled {
		for at, c := range w.inLane(lane) {
			out = append(out, Placed{Card: c, Slot: l.LaneCardAt(lane, at, w.lanes[lane]), Within: l.Lane[lane]})
		}
	}
	return out
}

// seen says whether any of a slot shows through what it is seen through.
func (p Placed) seen() bool {
	return p.Slot.Y < p.Within.Y+p.Within.H && p.Slot.Y+p.Slot.H > p.Within.Y
}

// StonesOf is a card built of stones: courses across it, each cut into
// stones of their own widths, every corner drawn in a little from its
// course and every edge given a middle that is not quite on the line — the
// same stones every time for the same card, and other stones for another.
func StonesOf(panel Box, seed int) [][]Point {
	r := stoneRand(uint32(seed)*2654435761 + 1)
	courses := 2 + r.below(2)
	var stones [][]Point
	top := panel.Y
	for c := range courses {
		bottom := panel.Y + panel.H*(c+1)/courses
		if c < courses-1 {
			bottom += r.between(-panel.H/(4*courses), panel.H/(4*courses))
		}
		left := panel.X
		for left < panel.X+panel.W {
			right := cut(left, panel.X+panel.W, stoneLeast+r.below(max(1, panel.W/2)))
			stones = append(stones, r.stone(Box{X: left, Y: top, W: right - left, H: bottom - top}))
			left = right
		}
		top = bottom
	}
	return stones
}

// cut is where a stone wanted so wide from left ends in a course ending at
// end: never past it, and at it when what would be left is too narrow to be
// a stone of its own.
func cut(left, end, wanted int) int {
	right := min(end, left+wanted)
	if end-right < stoneLeast {
		return end
	}
	return right
}

// Dropped is how far above its place a card's stones are, having fallen so
// far: from above the window, at nought, to their place, at one.
func Dropped(panel Box, fall float64) int { return -int(float64(panel.Y+panel.H) * (1 - fall)) }

// Fallen is a stone moved down by drop.
func Fallen(stone []Point, drop int) []Point {
	moved := make([]Point, len(stone))
	for i, p := range stone {
		moved[i] = Point{X: p.X, Y: p.Y + drop}
	}
	return moved
}

// stoneRand is a small generator of its own, so the same card is built of
// the same stones on every machine and every run.
type stoneRand uint32

func (r *stoneRand) next() uint32 {
	*r = *r*1664525 + 1013904223
	return uint32(*r >> 8)
}

// below is a number from nought up to, not including, n.
func (r *stoneRand) below(n int) int { return int(r.next() % uint32(max(1, n))) }

// between is a number from lo to hi, both included.
func (r *stoneRand) between(lo, hi int) int { return lo + r.below(hi-lo+1) }

// stone is one stone cut from a cell of its course: six corners, each drawn
// in from the cell by a little of its own, the gap between stones left round
// it, and never outside the cell.
func (r *stoneRand) stone(cell Box) []Point {
	in := func() int { return stoneGap + 1 + r.below(stoneWobble) }
	x0, x1 := cell.X+in(), cell.X+cell.W-1-in()
	y0, y1 := cell.Y+in(), cell.Y+cell.H-1-in()
	mid := cell.X + cell.W/2 + r.between(-cell.W/6, cell.W/6)
	return []Point{
		{X: x0, Y: y0},
		{X: mid, Y: cell.Y + stoneGap + r.below(stoneWobble)},
		{X: x1, Y: cell.Y + in()},
		{X: cell.X + cell.W - 1 - in(), Y: y1},
		{X: mid + r.between(-cell.W/8, cell.W/8), Y: cell.Y + cell.H - 1 - stoneGap - r.below(stoneWobble)},
		{X: cell.X + in(), Y: cell.Y + cell.H - 1 - in()},
	}
}
