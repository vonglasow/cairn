package gui

import (
	"fmt"
	"strconv"

	"cairn/internal/engine"
	"cairn/internal/reading"
)

// What the window shows of one task opened: what hangs under it and what it
// serves, its history, and what it says — and how the pointer and the keys
// walk each of them.

// Ghost is the card in mid-air under the pointer, and whether there is one. A
// drop with nothing shown moving is a drop nobody can aim.
func (w Window) Ghost() (Box, bool) {
	if !w.Dragging() {
		return Box{}, false
	}
	// A tile in hand is a tile, not a card: the same ghost, sized to what is
	// being carried.
	if w.screen == OnDetail {
		box := ServingBox(Over(w.width, w.height), w.Shown())
		return Box{X: box.X, Y: w.held.At.Y - SubHeight/2, W: box.W, H: SubHeight}, true
	}
	l := w.Layout()
	return Box{
		X: w.held.At.X - l.Column[0].W/2,
		Y: w.held.At.Y - CardHeight/2,
		W: l.Column[0].W,
		H: CardHeight,
	}, true
}

// Landing is the slot the card would drop into, and whether it is over one at
// all. Drawn as a line rather than guessed at by the hand.
func (w Window) Landing() (Box, bool) {
	if !w.Dragging() {
		return Box{}, false
	}
	// Where a thing may land depends on what is open, and nothing else about
	// the gesture does. Asked of the board while a part was in hand, this drew
	// the board's own line under the panel — for a drop that could never reach
	// a column.
	if w.screen == OnDetail {
		return w.landingUnder()
	}
	l := w.Layout()
	if lane, on := l.LaneUnder(w.held.At); on {
		return l.Lane[lane], true
	}
	for i := range Columns {
		if !l.Column[i].In(w.held.At) {
			continue
		}
		// Snapped to the grid on screen, and the column's scroll does not come
		// into it: the line is drawn where the hand is, and the hand is looking
		// at what is shown. It was added and taken away again here, which is a
		// sum that says nothing — found by mutation, which could not tell the
		// two halves apart because they cancel.
		at := (w.held.At.Y - l.Column[i].Y) / CardHeight
		return Box{
			X: l.Column[i].X,
			Y: l.Column[i].Y + at*CardHeight,
			W: l.Column[i].W,
			H: landingHigh,
		}, true
	}
	return Box{}, false
}

// landingHigh is how thick the line saying where a card would land is drawn.
const landingHigh = 3

// Acts is the history of the task the detail is open on, oldest first.
func (w Window) Acts() []string { return w.acts }

// Lines is the history as the lines it is drawn on, wrapped to the room the
// panel gives it.
func (w Window) Lines() []string {
	inside := Inside(HistoryBox(Over(w.width, w.height)))
	wrap := func(said string) Lines { return w.wrapping(said, inside, small) }
	// An objective has no history of its own — the trace is kept of tasks. It
	// says so rather than leaving the frame empty, which reads as a panel that
	// failed to load. The column itself stays the history's: it is what a task
	// opens onto, and taking it for an objective would cost every task the
	// thing this panel was built for.
	if w.chosenGoal {
		return wrap("an objective has no history of its own").Said
	}
	return Told(w.acts, wrap)
}

// servingHead names the box under the fields, where an objective says what
// serves it. PanelHead names the right column, which is the history's whatever
// is open.
const (
	historyHead = "HISTORY"
	servingHead = "WHAT SERVES IT"
	hangingHead = "WHAT HANGS UNDER IT"
)

// PanelHead is what the right column is called.
func (w Window) PanelHead() string { return historyHead }

// ServingSays is what that box reads when there is no card to draw in it. An
// objective nothing serves yet says so: an empty frame reads as a box that
// failed to load rather than as a question nobody has answered.
const ServingSays = "nothing serves it yet"

// Told is how many of those lines there are, and Rows how many are shown at
// once. The two together are what says whether there is anything to scroll.
func (w Window) Told() int { return len(w.Lines()) }

// Rows is how many lines of history the panel shows at once.
func (w Window) Rows() int { return HistoryRows(Over(w.width, w.height)) }

// History is the first line of the history that is shown.
func (w Window) History() int { return w.history }

// ChoosingFromRow is the first row of the open list in sight, and ChoosingRoom
// how many of them are in sight at once.
func (w Window) ChoosingFromRow() int { return w.choosingFrom }

// ChoosingRoom is how many rows the open list shows at once.
func (w Window) ChoosingRoom() int { return ChoosingRows(ChoosingOver(w.width, w.height)) }

func (w Window) settledChoosing(want int) int {
	return min(max(0, len(w.Choosing())-w.ChoosingRoom()), max(0, want))
}

// Saying is how far the description on an open task has been scrolled.
func (w Window) Saying() int { return w.settledSaying(w.saying) }

// SaidLines is the description as it is laid out in its box, and SaidRows how
// many of them are in sight.
func (w Window) SaidLines() []string {
	_, _, lines, _, _ := w.sayingOn()
	return lines.Said
}

// sayingOn is the long field the window is showing a scroll of — the
// description on an open task, or the long field the keyboard is on in the
// form — as one thing: its frame, its bar, the lines it is laid out on, and how
// many of them are in sight.
//
// One answer for both screens, and the lines are laid out on the width they
// are drawn on. The end of the scroll used to be worked out from the text laid
// out on a width a few points wider than the one it was drawn on, so what was
// drawn ran longer than what was counted and the last lines of a long
// description could never be brought into the box.
func (w Window) sayingOn() (frame, bar Box, lines Lines, rows int, ok bool) {
	box := Over(w.width, w.height)
	switch w.screen {
	case OnDetail:
		fields := w.Shown()
		at, long := LongField(fields)
		if !long {
			return Box{}, Box{}, Lines{}, 0, false
		}
		frame, bar = SayingBox(box, fields, at), SayingBar(box, fields, at)
		rows = LongRows(box, fields)
		return frame, bar, w.wrapping(w.FieldSays(fields[at]), Inside(frame), reads), rows, true
	case OnForm:
		fields := w.form.Fields()
		at := w.form.On()
		if !fields[at].Long {
			return Box{}, Box{}, Lines{}, 0, false
		}
		var value Box
		_, value, rows = FormPlace(box, fields, at)
		frame, bar = Saying(value, rows)
		return frame, bar, w.wrapping(w.form.Writing().Reads(), Inside(frame), reads), rows, true
	}
	return Box{}, Box{}, Lines{}, 0, false
}

// SaidRows is how many lines of the description its box holds.
func (w Window) SaidRows() int {
	_, _, _, rows, _ := w.sayingOn()
	return rows
}

func (w Window) settledSaying(want int) int {
	return min(max(0, len(w.SaidLines())-w.SaidRows()), max(0, want))
}

// settledHistory keeps the history between its two ends: never above its first
// line, never past the last screenful of it.
func (w Window) settledHistory(want int) int {
	return min(max(0, w.Told()-w.Rows()), max(0, want))
}

// settledServing keeps the work under an objective between its two ends, the
// way a column's scroll is kept between its own — in pixels, because these are
// cards of one height and not lines of text that wrap.
func (w Window) settledServing(want int) int {
	box := ServingBox(Over(w.width, w.height), w.Shown())
	return min(max(0, len(w.Under())*SubHeight-box.H), max(0, want))
}

// ServingRows is how many tiles that box shows at once.
func (w Window) ServingRows() int {
	return max(1, ServingBox(Over(w.width, w.height), w.Shown()).H/SubHeight)
}

// ServingCardAt is where one of those tiles is drawn, counting the scroll —
// the same reckoning a column makes of its cards.
func (w Window) ServingCardAt(at int) Box {
	box := ServingBox(Over(w.width, w.height), w.Shown())
	return Box{X: box.X, Y: box.Y + at*SubHeight - w.serving, W: box.W, H: SubHeight}
}

// UnderAt is which of them the keyboard is on.
func (w Window) UnderAt() int { return w.underAt }

// walkingUnder moves the keyboard through the box and brings what it lands on
// into sight, the way walking a column does. A cursor that could leave the
// frame would be a cursor somebody loses.
func (w Window) walkingUnder(by int) Window {
	under := w.Under()
	if len(under) == 0 {
		return w
	}
	w.underAt = min(len(under)-1, max(0, w.underAt+by))
	top := w.underAt * SubHeight
	w.serving = min(w.serving, top)
	// max rather than a comparison written out: the two say the same thing,
	// and the written-out one generates a boundary mutant no test can kill —
	// its branches part company only where they agree.
	w.serving = max(w.serving, top+SubHeight-w.ServingRows()*SubHeight)
	return w
}

// marking turns the part the keyboard is on done, or back. One key and one
// act: this is not a form, and a tile is one fact somebody is either changing
// or not, so it reaches the store at once.
func (w Window) marking() Window {
	// Only what hangs under a task. The work under an objective is tasks with
	// columns of their own — one of them may be current, or blocked — and
	// "not done" for those is not todo. A part of a task has nothing else to
	// be, which is what makes one key enough here and not there.
	if w.chosenGoal {
		return w
	}
	// Walked to rather than indexed: whatever moved the cursor keeps it inside,
	// so an index needs a guard that guards against nothing — and a clamp
	// written to satisfy one hides its own arithmetic behind the min.
	var on reading.Card
	var found bool
	for at, c := range w.Under() {
		if at == w.underAt {
			on, found = c, true
		}
	}
	if !found {
		return w
	}
	return w.asking(engine.TaskEdit, engine.Input{Given: map[string]string{
		"number": strconv.Itoa(on.Number),
		"state":  reading.PartState(!reading.Finished(on)),
	}})
}

// landingUnder is the line a tile would land on, snapped to the slot the hand
// is over — the same reckoning the board makes of a column, in the one place a
// part may go.
func (w Window) landingUnder() (Box, bool) {
	box := ServingBox(Over(w.width, w.height), w.Shown())
	if !box.In(w.held.At) {
		return Box{}, false
	}
	at := (w.held.At.Y - box.Y) / SubHeight
	return Box{X: box.X, Y: box.Y + at*SubHeight, W: box.W, H: landingHigh}, true
}

// closesOnAClickOff says whether what is open goes away when a click lands
// outside it. Everything that can be left without losing anything: the detail,
// the help, the question about deleting, the list of objectives.
//
// Not the form, and not the board or a search, which have nothing to close.
func (w Window) closesOnAClickOff() bool {
	switch w.screen {
	case OnDetail, OnHelp, OnAsking, OnChoosing:
		return true
	}
	return false
}

// tileUnder says which part of the work a point falls on, if any.
func (w Window) tileUnder(p Point) (int, bool) {
	box := ServingBox(Over(w.width, w.height), w.Shown())
	if !box.In(p) {
		return 0, false
	}
	at := (p.Y - box.Y + w.serving) / SubHeight
	if at < 0 || at >= len(w.Under()) {
		return 0, false
	}
	return at, true
}

// droppedUnder is a tile let go: it takes the place it was let go on, and what
// was there moves down. Where it lands is reckoned the way a column reckons a
// card — the slot the pointer is over, which is first or behind the one above.
//
// Let go anywhere but in the box, nothing happens. A part has no column to
// fall into and its parent is where it belongs, so there is nowhere else for
// it to go.
func (w Window) droppedUnder(p Point) Window {
	held := w.held
	w.held = Held{}
	box := ServingBox(Over(w.width, w.height), w.Shown())
	if !box.In(p) || held.Number == 0 {
		return w
	}
	at := (p.Y - box.Y + w.serving) / SubHeight
	under := w.Under()
	// Let go on its own place, or on the edge just below it, nothing is asked:
	// both name where it already is, and a move to where a thing already sits
	// is a line of history saying nothing happened.
	for i, c := range under {
		if c.Number == held.Number && (at == i || at == i+1) {
			return w
		}
	}
	// Written with ifs rather than a switch, for the reason internal/engine
	// gives where it does the same: a tagless switch puts its conditions
	// outside any covered block, so a mutation of one reports as unreachable
	// and the gate cannot judge it. Measured here — six of them at once.
	given := map[string]string{"number": strconv.Itoa(held.Number)}
	if at <= 0 {
		given["first"] = "true"
	} else if at-1 < len(under) {
		given["after"] = strconv.Itoa(under[at-1].Number)
	} else {
		given["last"] = "true"
	}
	return w.asking(engine.TaskMove, engine.Input{Given: given})
}

// ServingAt is how far that box has been scrolled.
func (w Window) ServingAt() int { return w.serving }

// Reading is the work as it was last read, which is what everything draws from.
func (w Window) Reading() reading.Snapshot { return w.snap }

// Says is what a window too small to draw in says instead. It names both sizes
// it needs, because "too small" without a number is something nobody can act
// on.
func (w Window) Says() string {
	if w.screen != TooSmallToDraw {
		return ""
	}
	return fmt.Sprintf(
		"This window is %d by %d. The board needs %d by %d.",
		w.width, w.height, Narrowest, Shortest)
}

// scrollingSaying is the wheel over a long field, on whichever screen has one.
func (w Window) scrollingSaying(p Point, by int) Window {
	frame, bar, _, _, ok := w.sayingOn()
	if !ok || !frame.In(p) && !bar.In(p) {
		return w
	}
	lines, left := turning(w.spareSaying, by)
	w.saying, w.spareSaying = w.settledSaying(w.saying+lines), left
	return w
}

// walksLines moves the caret a line up or down inside a long field, keeping as
// far into the line as it was, and brings it into sight. At either end it
// stays: leaving the field is what tab is for.
func (w Window) walksLines(by int) Window {
	_, _, lines, _, ok := w.sayingOn()
	if !ok {
		return w
	}
	writing := w.form.Writing()
	line, column := lines.At(writing.At)
	to := min(len(lines.Said)-1, max(0, line+by))
	at := lines.From[to] + min(column, len([]rune(lines.Said[to])))
	w.form = w.form.Writes(writing.CaretAt(at))
	return w.inSightOfTheCaret()
}
