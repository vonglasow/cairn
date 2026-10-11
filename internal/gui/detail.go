package gui

import (
	"fmt"

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
	inside := Inside(BesideBox(Over(w.width, w.height)))
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
func (w Window) Rows() int { return BesideRows(Over(w.width, w.height)) }

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
	switch w.screen {
	case OnForm:
		at, long := w.sayingAt()
		if !long {
			return Box{}, Box{}, Lines{}, 0, false
		}
		fields := w.form.Fields()
		var value Box
		_, value, rows = FormPlace(w.formBox(), fields, at)
		frame, bar = Saying(value, rows)
		said := fields[at].Value
		if at == w.form.On() {
			said = w.form.Writing().Reads()
		}
		return frame, bar, w.wrapping(said, Inside(frame), reads), rows, true
	}
	return Box{}, Box{}, Lines{}, 0, false
}

// sayingAt is which long field of a form the wheel and the bar move: the one
// the keyboard is on, or else the first, which is what a task is opened to
// read. A description reached only once the keyboard had been put on it was
// a description somebody opening a task to read could not scroll.
func (w Window) sayingAt() (int, bool) {
	fields := w.form.Fields()
	if len(fields) == 0 {
		return 0, false
	}
	if fields[w.form.On()].Long {
		return w.form.On(), true
	}
	for i, field := range fields {
		if field.Long {
			return i, true
		}
	}
	return 0, false
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
	box := BesideBox(Over(w.width, w.height))
	return min(max(0, len(w.Under())*SubHeight-box.H), max(0, want))
}

// ServingRows is how many tiles that box shows at once.
func (w Window) ServingRows() int {
	return max(1, BesideBox(Over(w.width, w.height)).H/SubHeight)
}

// ServingCardAt is where one of those tiles is drawn, counting the scroll —
// the same reckoning a column makes of its cards.
func (w Window) ServingCardAt(at int) Box {
	box := BesideBox(Over(w.width, w.height))
	return Box{X: box.X, Y: box.Y + at*SubHeight - w.serving, W: box.W, H: SubHeight}
}

// closesOnAClickOff says whether what is open goes away when a click lands
// outside it. Everything that can be left without losing anything: the detail,
// the help, the question about deleting, the list of objectives.
//
// Not the form, and not the board or a search, which have nothing to close.
// panel is the box the screen that has the window is drawn in: what its
// offers sit along the foot of, and what a click off it lands outside of.
func (w Window) panel() Box {
	switch w.screen {
	case OnForm:
		return w.formPanel()
	case OnChoosing:
		return ChoosingOver(w.width, w.height)
	case OnAsking:
		return w.QuestionBox()
	case OnHelp:
		return w.HelpBox()
	}
	return Over(w.width, w.height)
}

// formPanel is the panel a form is drawn in. One being added is cut to what it
// holds; one opened on something already there takes the whole panel, its
// fields down the left and, down the right, what was done to a task or the
// work serving an objective.
func (w Window) formPanel() Box {
	if w.form.Adding {
		return FormOver(w.width, w.height, w.form.Fields())
	}
	return Over(w.width, w.height)
}

// formBox is the box a form's fields are laid out in: the whole of a form
// being added, the left column of one opened on something already there.
func (w Window) formBox() Box {
	if w.form.Adding {
		return w.formPanel()
	}
	return FormFields(w.formPanel())
}

func (w Window) closesOnAClickOff() bool {
	switch w.screen {
	case OnHelp, OnAsking, OnChoosing:
		return true
	case OnForm:
		// A form nobody typed into loses nothing by going: it is a task
		// opened to look at as much as to change. One somebody typed into
		// stays, because a click landing anywhere outside would throw away
		// what was typed without asking.
		return !w.form.Changed()
	}
	return false
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
