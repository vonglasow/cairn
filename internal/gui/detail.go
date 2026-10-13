package gui

import (
	"cairn/internal/engine"
	"fmt"
)

// Ghost is the card in mid-air under the pointer, and whether there is one.
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

// Landing is the slot the card would drop into, and whether it is over one at all.
func (w Window) Landing() (Box, bool) {
	if !w.Dragging() {
		return Box{}, false
	}
	l := w.Layout()
	if lane, on := l.LaneUnder(w.held.At); on {
		return l.Lane[lane], true
	}
	for i := range engine.Columns {
		if !l.Column[i].In(w.held.At) {
			continue
		}
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

const landingHigh = 3

// Acts is the history of the task the detail is open on, oldest first.
func (w Window) Acts() []string { return w.acts }

// Lines is the history as the lines it is drawn on, wrapped to the room the panel gives it.
func (w Window) Lines() []string {
	inside := Inside(BesideBox(w.formPanel()))
	wrap := func(said string) Lines { return w.wrapping(said, inside, small) }
	if w.chosenGoal {
		return wrap("an objective has no trace of its own").Said
	}
	return Told(w.acts, wrap)
}

const (
	historyHead = "TRACE"
	servingHead = "WHAT SERVES IT"
	hangingHead = "WHAT HANGS UNDER IT"
)

// PanelHead is what the right column is called.
func (w Window) PanelHead() string { return historyHead }

// ServingSays is what that box reads when there is no card to draw in it.
const ServingSays = "nothing serves it yet"

// Told is how many of those lines there are, and Rows how many are shown at once.
func (w Window) Told() int { return len(w.Lines()) }

// Rows is how many lines of history the panel shows at once.
func (w Window) Rows() int { return BesideRows(w.formPanel()) }

// History is the first line of the history that is shown.
func (w Window) History() int { return w.history }

// ChoosingFromRow is the first row of the open list in sight, and ChoosingRoom how many of them are in sight at once.
func (w Window) ChoosingFromRow() int { return w.choosingFrom }

// ChoosingRoom is how many rows the open list shows at once.
func (w Window) ChoosingRoom() int { return ChoosingRows(w.ChoosingBox()) }

func (w Window) settledChoosing(want int) int {
	return min(max(0, len(w.Choosing())-w.ChoosingRoom()), max(0, want))
}

// Saying is how far the description on an open task has been scrolled.
func (w Window) Saying() int { return w.settledSaying(w.saying) }

// SaidLines is the description as it is laid out in its box, and SaidRows how many of them are in sight.
func (w Window) SaidLines() []string {
	_, _, lines, _, _ := w.sayingOn()
	return lines.Said
}

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

func (w Window) settledHistory(want int) int {
	return min(max(0, w.Told()-w.Rows()), max(0, want))
}

func (w Window) settledServing(want int) int {
	box := BesideBox(w.formPanel())
	return min(max(0, len(w.Under())*SubHeight-box.H), max(0, want))
}

// ServingRows is how many tiles that box shows at once.
func (w Window) ServingRows() int {
	return max(1, BesideBox(w.formPanel()).H/SubHeight)
}

// ServingCardAt is where one of those tiles is drawn, counting the scroll — the same reckoning a column makes of its cards.
func (w Window) ServingCardAt(at int) Box {
	box := BesideBox(w.formPanel())
	return Box{X: box.X, Y: box.Y + at*SubHeight - w.serving, W: box.W, H: SubHeight}
}

func (w Window) panel() Box {
	switch w.screen {
	case OnForm:
		return w.formPanel()
	case OnChoosing:
		return w.ChoosingBox()
	case OnAsking:
		return w.QuestionBox()
	case OnHelp:
		return w.HelpBox()
	case OnCalendar:
		return w.CalendarBox()
	case OnMenu:
		return w.MenuBox()
	case OnBackups:
		return w.BackupsBox()
	}
	return Over(w.width, w.height)
}

func (w Window) formPanel() Box {
	if w.form.Adding {
		return FormOver(w.width, w.height, w.form.Fields())
	}
	return OpenOver(w.width, w.height, w.form.Fields())
}

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
		return true
	}
	return false
}

// ServingAt is how far that box has been scrolled.
func (w Window) ServingAt() int { return w.serving }

// Reading is the work as it was last read, which is what everything draws from.
func (w Window) Reading() engine.Snapshot { return w.snap }

// Says is what a window too small to draw in says instead.
func (w Window) Says() string {
	if w.screen != TooSmallToDraw {
		return ""
	}
	return fmt.Sprintf(
		"This window is %d by %d. The board needs %d by %d.",
		w.width, w.height, Narrowest, Shortest)
}

func (w Window) scrollingSaying(p Point, by int) Window {
	frame, bar, _, _, ok := w.sayingOn()
	if !ok || !frame.In(p) && !bar.In(p) {
		return w
	}
	lines, left := turning(w.spareSaying, by)
	w.saying, w.spareSaying = w.settledSaying(w.saying+lines), left
	return w
}

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
