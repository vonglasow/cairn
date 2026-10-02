package gui

// The window as a value, and every gesture as a function from one to another.
//
// Gio is immediate mode: there is no loop handing back a new model the way
// bubbletea does, and a widget keeping its own state is the path of least
// resistance. So the state a decision turns on is kept here, changed here, and
// only then drawn — the widgets hold what they must and decide nothing. That is
// what lets a scenario click on a card and read what happened with no window
// anywhere, and letting it slip is how this surface would become the one the
// gates cannot reach.

import (
	"fmt"
	"strconv"
	"strings"

	"cairn/internal/domain"
	"cairn/internal/engine"
	"cairn/internal/reading"
	"cairn/internal/trace"
)

// Screen is what has the window.
type Screen int

const (
	// OnTheBoard is the board itself, which is where a window opens.
	OnTheBoard Screen = iota
	// OnDetail is one task, opened to read what a card only marked.
	OnDetail
	// OnForm is the modal over the board, adding or editing.
	OnForm
	// OnSearch is the board narrowed to what a search reaches, while it is
	// still being typed.
	OnSearch
	// OnAsking is the one question this window asks: deleting loses work, and
	// losing work without being asked is how work is lost.
	OnAsking
	// OnHelp is what everything does, which is the one thing here that has to
	// be read rather than pointed at.
	OnHelp
	// OnChoosing is the list of objectives a task may serve, open over the
	// form. Its own screen and not a state of the form, because what a key
	// does here is not what it does in a field: space ticks, and enter — which
	// saves everywhere else — is what opened this.
	OnChoosing
	// TooSmallToDraw is a window too small to say anything useful in.
	TooSmallToDraw
)

// Tab is which of the two is being looked at.
type Tab int

const (
	// Tasks is the work itself, and is where a window opens.
	Tasks Tab = iota
	// Objectives is what the work is for.
	Objectives
)

// Held is a card being dragged: which task, where it was picked up, and where
// the pointer is now.
type Held struct {
	Number int
	From   Point
	At     Point
	// Moved says the pointer has gone far enough for this to be a drag rather
	// than a click that wobbled.
	Moved bool
}

// Grip is how far a pointer travels before a click becomes a drag. Without it,
// a shaky hand moves a task every time it chooses one.
const Grip = 4

// Window is everything the desktop board is, at one moment.
type Window struct {
	work reading.Work

	width, height int
	snap          reading.Snapshot
	version       int64

	screen Screen
	tab    Tab
	chosen int
	// chosenGoal says which of the two kinds the number belongs to. A number
	// alone is not enough: cairn numbers tasks and objectives apart, so task 1
	// and objective 1 both exist, and a board that kept only the number opened
	// the task whenever an objective was pointed at.
	chosenGoal bool

	scroll [Columns]int
	// goals is how far down the objectives somebody is. Its own number and not
	// a column's: the objectives are one list across the whole board, and the
	// four columns are four.
	goals int
	// history is how far down a task's history somebody is, counted in lines
	// and not in acts: an act wraps onto as many lines as it needs, and a
	// history that jumped an act at a time would skip past what it wrapped.
	history int
	// saying is how far the description on an open task has been scrolled,
	// counted in lines the way the history is.
	saying int
	// What a wheel sent that did not come to a whole line, kept until the
	// turns after it make one.
	spareHistory, spareSaying, spareChoosing int
	// choosingFor is the field the open list is choosing for, and filtering
	// what is being typed into it to narrow that list.
	choosingFor string
	filtering   Written
	// choosingFrom is the first row of the list in sight. A list of everything
	// a task may hang under is as long as the board, and a panel is not.
	choosingFrom int
	// copying is what the window has been asked to put on the clipboard, and
	// pasting that it has asked for what is on it. The window is a value and
	// cannot execute anything: it says what it wants and the loop does it.
	copying string
	pasting bool
	// holdingText says a press landed in a field and the pointer has not been
	// let go: dragging from there holds a run, which is how text is selected
	// with a mouse everywhere else.
	holdingText bool
	// serving is how far the work under an open objective has been scrolled,
	// counted in lines the way the history is: a task wraps onto as many lines
	// as its title needs, and a list that jumped a task at a time would skip
	// past what it wrapped.
	serving int
	// draggingSaying says the pointer has taken hold of the bar beside the
	// description, and draggingServing the one beside what hangs under a task.
	// All of them are pulled the way the history's is.
	draggingSaying, draggingServing bool
	// draggingChoosing says the same of the list's own bar. Its thumb is sized
	// never to be too small to take hold of, and taking hold of it did nothing
	// at all — two of the window's four bars answered a drag, and two invited
	// one and stayed where they were.
	draggingChoosing bool
	// draggingBar says the pointer has taken hold of the bar beside the
	// history. Told apart from a held card, because a panel is over the board
	// and the board's cards are not there to be taken hold of.
	draggingBar bool

	held Held
	// pressed says a press is still waiting for its release, and on which
	// screen it happened. A release belongs to the press it follows: if what
	// is on screen changed in between — a double click opening a task — then
	// the release is not a click on a screen the press never saw.
	pressed   bool
	pressedOn Screen
	acts      []string
	form      Form
	// looking is what a search has been given so far, and where its caret
	// stands. A search narrows what is drawn and changes nothing, so it lives
	// here and never reaches the store.
	looking Written
	// archived says what was taken off the board is drawn with the rest,
	// dimmed. Nothing is asked of the store either way.
	archived bool
	// over is the offer the pointer is resting on, and whether it is resting
	// on one at all. Nothing in the window used to answer a pointer before it
	// was pressed, so "Delete" read exactly like a label — the one thing that
	// says a word can be clicked was missing from the whole surface.
	over    string
	hovered bool
	// question is which of the two things the one question screen is asking
	// about. There used to be only one.
	question Question
	// wasScreen is what was open when the window became too small to draw it,
	// so growing back gives it back rather than an empty board.
	wasScreen Screen
	// ticked is the objectives the open list has in, by number, and choosing
	// is where in that list the keyboard sits. They live on the window and not
	// in the form because the list is a screen of its own; the form is given
	// what they came to only when the list closes.
	ticked   map[int]bool
	choosing int
	// underAt is which part of the work the keyboard is on in the box beneath
	// an open task's fields. Put back to the first whenever another thing is
	// opened: a cursor kept from the last one would point at a part of it that
	// is not there any more.
	underAt int

	err error
}

// New builds a window over whatever holds the work.
func New(work reading.Work, width, height int) Window {
	w := Window{work: work, width: width, height: height}
	return w.sized()
}

// Resized answers the window this one becomes at another size.
func (w Window) Resized(width, height int) Window {
	w.width, w.height = width, height
	return w.sized()
}

// sized settles which screen a size means. A window too small says what it
// needs rather than drawing something nobody can use; growing back draws again.
func (w Window) sized() Window {
	// An if-chain and not a tagless switch, for the reason test/harness.go
	// gives: a switch with no subject puts its conditions outside every block
	// coverage sees, so a wrong one would go out unjudged.
	if TooSmall(w.width, w.height) {
		// What was open is remembered rather than closed. A window dragged
		// narrow for a moment used to come back as an empty board, taking a
		// form somebody had been filling in with it.
		if w.screen != TooSmallToDraw {
			w.wasScreen = w.screen
		}
		w.screen = TooSmallToDraw
		return w
	}
	if w.screen == TooSmallToDraw {
		w.screen = w.wasScreen
	}
	return w
}

// Layout is every size this window draws from.
func (w Window) Layout() Layout { return LayoutFor(w.width, w.height) }

// Screen says what has the window.
func (w Window) Screen() Screen { return w.screen }

// Chosen is the task picked out, and zero when none is.
func (w Window) Chosen() int { return w.chosen }

// Wrong is what went wrong, if anything did.
func (w Window) Wrong() error { return w.err }

// Dragging says whether a card is in hand and has travelled far enough to be a
// drag rather than a click that wobbled.
func (w Window) Dragging() bool { return w.held.Number != 0 && w.held.Moved }

// Tab is which of the two is being looked at.
func (w Window) Tab() Tab { return w.tab }

// Shown is everything the task the detail is open on carries, read off the same
// description a form asks from. A card has room for a title and a number; this
// is the rest of it, which is what a detail is for.
func (w Window) Shown() []reading.Field {
	if w.screen != OnDetail {
		return nil
	}
	if w.chosenGoal {
		for _, g := range w.snap.Goals {
			if g.Number == w.chosen {
				return reading.GoalFields(g)
			}
		}
		return nil
	}
	for _, c := range w.snap.Cards {
		if c.Number == w.chosen {
			return reading.TaskFields(c)
		}
	}
	return nil
}

// ChosenGoal says whether what is chosen is an objective rather than a task.
func (w Window) ChosenGoal() bool { return w.chosenGoal }

// Serving is the work under the objective that is open, by title and by where
// each task stands. Empty unless an objective is what was opened: a task is
// served by nothing, and the question does not arise.
func (w Window) Serving() []reading.Card {
	if !w.chosenGoal {
		return nil
	}
	return reading.Serving(w.snap.Cards, w.chosen)
}

// Under is what the box beneath the fields holds: the work that serves an open
// objective, or the work hanging under an open task. One box and one question
// — what is under this — asked of whichever kind of thing was opened.
func (w Window) Under() []reading.Card {
	if w.chosenGoal {
		return reading.Serving(w.snap.Cards, w.chosen)
	}
	return reading.Hanging(w.snap.Cards, w.chosen)
}

// UnderHead is what that box is called, which depends on what is open: an
// objective is served, a task is hung under, and calling both by one word
// would make the box say less than it knows.
func (w Window) UnderHead() string {
	if w.chosenGoal {
		return servingHead
	}
	return hangingHead
}

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
	wide := room(Inside(HistoryBox(Over(w.width, w.height))))
	// An objective has no history of its own — the trace is kept of tasks. It
	// says so rather than leaving the frame empty, which reads as a panel that
	// failed to load. The column itself stays the history's: it is what a task
	// opens onto, and taking it for an objective would cost every task the
	// thing this panel was built for.
	if w.chosenGoal {
		return Wrapped("an objective has no history of its own", wide).Said
	}
	return Told(w.acts, wide)
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
		return frame, bar, Wrapped(w.FieldSays(fields[at]), room(Inside(frame))), rows, true
	case OnForm:
		fields := w.form.Fields()
		at := w.form.On()
		if !fields[at].Long {
			return Box{}, Box{}, Lines{}, 0, false
		}
		_, value := FieldRow(box, FieldTop(box, fields, at))
		rows = RowsOf(box, fields, at)
		frame, bar = Saying(value, rows)
		return frame, bar, Wrapped(w.form.Writing().Reads(), room(Inside(frame))), rows, true
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
	into := string(domain.Todo)
	if !reading.Finished(on) {
		into = string(domain.Done)
	}
	asked := engine.Input{Given: map[string]string{
		"number": strconv.Itoa(on.Number),
		"state":  into,
	}}
	if _, err := w.work.Ask(engine.TaskEdit, asked); err != nil {
		w.err = err
		return w
	}
	return w.Look()
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
	if _, err := w.work.Ask(engine.TaskMove, engine.Input{Given: given}); err != nil {
		w.err = err
		return w
	}
	return w.Look()
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

// Click chooses whatever is under the pointer, and nothing when there is
// nothing there.
func (w Window) Click(p Point) Window {
	// The bar answers whatever is on screen: it is what closes a panel, so it
	// is the one thing a panel does not swallow.
	if offer, on := w.offerAt(p); on {
		return w.Choose(offer)
	}
	// A list answers before the form does. It is drawn over one, and the boxes
	// underneath are still where they were — a click meant for a row used to
	// fall through and move the keyboard to whatever field was behind it.
	if w.screen == OnChoosing {
		return w.clickingList(p)
	}
	// A form's own fields answer, because that is how the keyboard gets to one
	// with a mouse.
	if at, on := w.fieldAt(p); on {
		w.form = w.form.onto(at)
		// A field a list answers opens it. There is nothing to type there, so
		// a click that only moved the keyboard onto it left a pointer with
		// nowhere to go and nothing saying a list was behind it.
		if chosenFromAList(w.form.Fields()[at].Arg) {
			return w.choosingList()
		}
		// A field holding one of a closed set moves to the next of them. Its
		// value changed by the arrows alone, so a column or a priority could
		// not be set with a mouse at all.
		if len(w.form.Fields()[at].Choices) > 0 {
			return w.cycle(1)
		}
		return w.CaretInto(w.reaching(at, p))
	}
	// Off a panel is out of it, where getting out costs nothing. A hand on the
	// mouse should not have to reach for the keyboard to put something away.
	//
	// A form is the exception and the only one: it holds what somebody typed
	// and has not saved, and a click landing anywhere outside would throw that
	// away without asking. esc is still there, and the bar offers it.
	if w.closesOnAClickOff() && !Over(w.width, w.height).In(p) {
		return w.Choose("Close")
	}
	// A search is not a panel over the board: the board is still there, still
	// showing what the search left, and a card under the pointer is a card.
	// Making it wait for enter meant closing the box to reach what the box had
	// just found. Over the box itself a click belongs to what is being typed.
	if w.screen == OnSearch {
		if Searching(w.Layout()).In(p) {
			return w
		}
	} else if w.screen != OnTheBoard {
		// And a panel over the board does have the window. Letting a click
		// through to the board behind it is how an open detail came to show
		// something nobody asked it for.
		return w
	}
	if tab, on := w.tabAt(p); on {
		w.tab, w.chosen, w.chosenGoal = tab, 0, false
		return w
	}
	w.chosen, w.chosenGoal = w.at(p), w.tab == Objectives
	return w
}

// Showing is the work narrowed to what a search reaches, which is what
// everything draws from. With no search it is the whole reading.
//
// The narrowing is done here and not asked of the store: a search is a way of
// looking, and a board that asked the store to search would be a board that
// waits to draw.
func (w Window) Showing() reading.Snapshot {
	if w.looking.Empty() {
		return w.snap
	}
	found := reading.Snapshot{Focus: w.snap.Focus, Version: w.snap.Version}
	// What a board shows of a search: a part that matched answers as the work
	// it is part of, since a sub-task has no card of its own here to appear on.
	found.Cards = reading.Found(w.snap.Cards, w.looking.Reads(), w.archived)
	for _, g := range w.snap.Goals {
		if reading.ReachesGoal(g, w.looking.Reads()) {
			found.Goals = append(found.Goals, g)
		}
	}
	return found
}

// Looking is what has been typed into the search so far.
func (w Window) Looking() string { return w.looking.Reads() }

// Searched is the search as a written text, caret and all.
func (w Window) Searched() Written { return w.looking }

// other is the page somebody is not on. Both tab and shift+tab come here.
//
// Written as "the other one" and not as a walk by a given number of pages.
// There are two, so which way round a walk goes is a difference nothing can
// see — and written as a walk it was arithmetic no test could ever judge, which
// the mutation gate said out loud. A third page would make it a walk again, and
// would need this rewritten rather than extended.
//
// Changing page lets go of what was chosen: a card chosen on the page you left
// is a card the arrows would walk from on the page you are on.
func (w Window) other() Window {
	if w.tab == Tasks {
		w.tab = Objectives
	} else {
		w.tab = Tasks
	}
	w.chosen, w.chosenGoal = 0, false
	return w
}

// working takes up the chosen task, or lets it go when it is already being
// worked on. Both name the task: several can be in focus at once, so letting
// go of nothing in particular would let go of every one of them.
func (w Window) working() Window {
	op := engine.TaskFocus
	for _, c := range w.snap.Cards {
		if c.Number == w.chosen && c.InFocus {
			op = engine.TaskUnfocus
		}
	}
	if _, err := w.work.Ask(op, reading.Numbered(w.chosen)); err != nil {
		w.err = err
		return w
	}
	return w.Look()
}

// chosenIsArchived says whether the card chosen is one off the board.
func (w Window) chosenIsArchived() bool {
	for _, c := range w.snap.Cards {
		if c.Number == w.chosen {
			return c.Archived
		}
	}
	return false
}

// puttingAway takes the chosen card off the board, or brings it back.
func (w Window) puttingAway() Window {
	op := engine.TaskArchive
	if w.chosenIsArchived() {
		op = engine.TaskUnarchive
	}
	if _, err := w.work.Ask(op, reading.Numbered(w.chosen)); err != nil {
		w.err = err
		return w
	}
	return w.Look()
}

// LaneSays is what a lane of the band writes above itself: what it is, and how
// many it holds. The four columns have counted since the beginning; the three
// lanes did not, so the one run of cards that always wants acting on was the
// one nobody could count without counting.
func (w Window) LaneSays(kind string) string {
	_, stalled := reading.Deal(w.Showing().Cards, w.archived)
	held := 0
	for _, c := range stalled {
		if c.State == kind {
			held++
		}
	}
	return kind + "  " + strconv.Itoa(held)
}

// Holds is how much a page holds, which is what its tab says beside its name.
// Read off what is showing, so a search that narrows the board narrows what the
// tabs claim as well.
func (w Window) Holds(tab Tab) int {
	if tab == Objectives {
		return len(w.Showing().Goals)
	}
	return len(w.Showing().Cards)
}

// Pages is what the two tabs are called, in the order they are drawn. Here
// and not in the drawing, because which page somebody is on is a thing the
// window knows and the drawing only shows.
var Pages = []string{"TASKS", "OBJECTIVES"}

// Heads is what the panel that is open says it is, written across its top.
// Nothing when no panel is open: the board is its own heading.
func Heads(w Window) string { return HeadsOn(w, w.screen) }

// HeadsOn is what a given panel says it is, whether or not it is the one on
// top. A panel drawn under another is still drawn, and asking the window what
// it was on made the form under an open list read the list's own heading.
func HeadsOn(w Window, screen Screen) string {
	switch screen {
	case OnChoosing:
		// What is being chosen, and what for. A list with no heading is a list
		// somebody has to remember the way into.
		what := "objectives for"
		if w.OneOnly() {
			what = "a parent for"
		}
		return what + " " + w.formKind() + " " + strconv.Itoa(w.form.Number)
	case OnHelp:
		return "What everything does"
	case OnAsking:
		if w.question == aboutLeaving {
			return "Throw away what you typed?"
		}
		if w.question == aboutSweeping {
			return "Take every task in " + strings.ToUpper(w.sweeps()) + " off the board?"
		}
		return "Delete " + w.kind() + " " + strconv.Itoa(w.chosen) + "?"
	case OnDetail:
		// The whole title, on the line beside the number. It is the one line
		// on the panel that was empty, it holds ninety-nine characters at the
		// smallest window this board draws in, and the title is what somebody
		// opened the task to read — the field below it cuts at thirty.
		return headed(w.kind()+" "+strconv.Itoa(w.chosen), w.titled())
	case OnForm:
		if w.form.Adding {
			return "New " + w.formKind()
		}
		return w.formKind() + " " + strconv.Itoa(w.form.Number)
	}
	return ""
}

// kind is the word for what is chosen, and formKind for what a form is about.
// Two of them because a form can be open over nothing chosen at all.
func (w Window) kind() string {
	if w.chosenGoal {
		return "objective"
	}
	return "task"
}

// titled is the title of whatever is open, read off the fields it shows so a
// task and an objective answer the same way.
func (w Window) titled() string {
	for _, field := range w.Shown() {
		if field.Arg == "title" {
			return field.Value
		}
	}
	return ""
}

// headed joins a head to what it is about, and is the head alone when there is
// nothing to join.
func headed(head, about string) string {
	if about == "" {
		return head
	}
	return head + " · " + about
}

func (w Window) formKind() string {
	if w.form.ForObjective {
		return "objective"
	}
	return "task"
}

// Shortcut is the key that does an offer without the pointer. Every offer has
// one: the bar is what somebody reads to find out what a key does, and an
// offer with no key is a word the bar cannot teach anything about.
//
// Delete is d and not the delete key, because that is what the terminal board
// has always used and one of the two had to give.
func Shortcut(offer string) string {
	switch strings.ToLower(offer) {
	case "new":
		return "cmd+n"
	case "search":
		return "/"
	case "help":
		return "?"
	case "archived":
		return "shift+a"
	case "archive", "restore":
		return "x"
	case "take":
		return "enter"
	case "keep":
		// Not space, which a list with a line to type into cannot have: a
		// title has spaces in it, so the filter needs them and the shortcut
		// cannot. Taking a row and being done with the list are two acts, and
		// the second gets the modifier.
		return "cmd+enter"
	case "edit":
		return "e"
	case "delete":
		return "d"
	case "agree":
		// Not d, which is what opened the question. A quick dd — which is how
		// a deletion is typed in the grammar this product borrows — used to
		// remove a task before a word of the question had been read.
		return "enter"
	case "choose":
		return "enter"
	case "save":
		// One key, and not enter. A form has two acts and enter is spent on
		// the other: it opens the list a field answers with, and it breaks a
		// line in a field with room for one. ⌘⏎ saves wherever the keyboard
		// is, which is what the bar says.
		return "cmd+enter"
	case "done":
		return "space"
	case "close", "cancel", "clear":
		return "esc"
	}
	return ""
}

// Pretty is how the bar writes a key, which is not how a key arrives. The two
// are told apart on purpose: a bar that wrote "cmd+n" would be reading out a
// name meant for a program, and a key looked up by "⌘N" would never be found.
func Pretty(key string) string {
	switch key {
	case "cmd+n":
		return "⌘N"
	case "cmd+enter":
		return "⌘⏎"
	case "shift+a":
		return "⇧A"
	case "enter":
		return "⏎"
	}
	return key
}

// Helping is what everything does, as a person reads it.
func (w Window) Helping() []reading.Field {
	return []reading.Field{
		{Name: "click", Value: "choose a card"},
		{Name: "double-click", Value: "open it, and see everything it carries"},
		{Name: "drag", Value: "move it — the column it lands in is its new state"},
		{Name: "drag into a lane", Value: "blocked, watching or paused"},
		{Name: "wheel", Value: "scroll what is under the pointer — a column, the objectives, a history"},
		{Name: "↑↓ under a task", Value: "walk what hangs under it — space marks one done, or not"},
		{Name: "wheel on a history", Value: "scroll it — the bar beside it can be pulled too"},
		{Name: "^D ^U on an open task", Value: "walk its description half a box at a time — the wheel and the bar do too"},
		{Name: "e", Value: "edit what is chosen, on the board or on an open task"},
		{Name: "f", Value: "take up what is chosen, or let it go — several at once is fine"},
		{Name: "d", Value: "delete what is chosen — it asks first"},
		{Name: "x", Value: "take what is chosen off the board, or put it back"},
		{Name: "⇧X", Value: "take every task in that card's column off the board — it asks first"},
		{Name: "←→↑↓", Value: "walk the board, the band included"},
		{Name: "⇧←→↑↓", Value: "move what is chosen — sideways changes its column"},
		{Name: "⏎ on the board", Value: "open what is chosen"},
		{Name: "⏎ on a question", Value: "agree to it"},
		{Name: "⏎ in a form", Value: "a new line in the description, the next field anywhere else — it never saves"},
		{Name: "↑↓ in the description", Value: "walk its lines — tab leaves it, and the wheel and the bar scroll it"},
		{Name: "⌘⏎ or ⌘S in a form", Value: "save, wherever the keyboard is"},
		{Name: "⌘A ⌘C ⌘X ⌘V", Value: "hold all of a field, copy, cut, paste"},
		{Name: "⇧←→ in a field", Value: "hold a run of text — dragging across it does the same"},
		{Name: "⏎ in a list", Value: "take the row · ⌘⏎ keeps the list · j and k walk it"},
		{Name: "⇧A", Value: "what was archived, on the board or off it — it reads faintly"},
		{Name: "⌘N", Value: "add"},
		{Name: "/ or ⌘F", Value: "search — it reaches the description and the reason"},
		{Name: "⏎ in a search", Value: "keep what it narrowed and go back to the board"},
		{Name: "tab", Value: "the other page — shift+tab the one before"},
		{Name: "tab, ↑↓ in a form", Value: "walk its fields — inside the description ↑↓ walk its lines"},
		{Name: "←→", Value: "change a field that has a closed set"},
		{Name: "esc", Value: "back out of whatever is open"},
	}
}

// Offers is what the bar says can be done here. It changes with what is on
// screen and with what is chosen, because an offer that does nothing is worse
// than no offer.
func (w Window) Offers() []string {
	switch w.screen {
	case OnChoosing:
		// Take and Done, not Choose twice. Space takes what the keyboard is
		// on, which is the gesture a hand already has for a list; enter is
		// done with the whole list, which is a different act and says so.
		return []string{"Take", "Keep", "Cancel"}
	case OnForm:
		// The one field a list answers says so on the bar. enter is Save's key
		// everywhere else in a form, and a bar that went on offering Save here
		// while enter opened a list would be a bar that lies — which is the
		// one thing this bar exists not to do. Save comes back the moment the
		// keyboard leaves this field.
		// Save is always offered. It used to give way to Choose on a field a
		// list answers, and with the keyboard left on that field after the
		// list closed there was no way to save at all — the one act a form
		// exists for, unreachable for having chosen something.
		if chosenFromAList(w.form.Fields()[w.form.On()].Arg) {
			return []string{"Choose", "Save", "Cancel"}
		}
		return []string{"Save", "Cancel"}
	case OnDetail:
		// Done only where there is a part to mark. The bar is where a key is
		// looked for, so it says what space does exactly where space does it,
		// and says nothing of it on a task nothing hangs under.
		if len(w.Under()) > 0 && !w.chosenGoal {
			return []string{"Done", "Edit", "Close"}
		}
		return []string{"Edit", "Close"}
	case OnAsking:
		return []string{"Agree", "Cancel"}
	case OnHelp:
		return []string{"Close"}
	case OnSearch:
		// One word, at the right of the bar, where the box no longer reaches.
		// Offering nothing left the pointer with no way out of a search and no
		// way to act on the card it had just found.
		return []string{"Close"}
	}
	offers := []string{"New", "Search", "Archived", "Help"}
	if !w.looking.Empty() {
		offers = append(offers, "Clear")
	}
	if w.chosen != 0 {
		// Archive and Restore are the same key on the same card: the bar names
		// which way it goes, so nobody has to work out which it will be.
		away := "Archive"
		if w.chosenIsArchived() {
			away = "Restore"
		}
		offers = append(offers, "Edit", away, "Delete")
	}
	return offers
}

// clickingList is what a click does inside an open list: a row is taken the
// way the keyboard takes it, and the line that narrows it takes the caret.
// Anywhere else inside the panel it does nothing, which is what a click on a
// panel's own ground has always done.
func (w Window) clickingList(p Point) Window {
	box := ChoosingOver(w.width, w.height)
	// A click puts the keyboard on the row; it does not take it. Taking is
	// space, and a double-click — a single click that took would make choosing
	// which row to read impossible without choosing it.
	if at, on := ChoosingAtPoint(box, len(w.Choosing()), w.choosingFrom, p); on {
		w.choosing = at
		return w
	}
	if looking := ChoosingLooking(box); looking.In(p) {
		w.filtering = w.filtering.CaretAt(into(Inside(looking), p))
	}
	return w
}

// chosenFromAList says whether a field is answered by picking from a list
// rather than by typing. Both are numbers somebody would otherwise have to go
// and look up, which is the whole of why the list exists.
func chosenFromAList(arg string) bool { return arg == "objective" || arg == "parent" }

// offerAt says which offer a point lands on, in the bar along the bottom or on
// the panel that has the window. Both say the same things and both answer,
// because a person clicks what they are looking at.
func (w Window) offerAt(p Point) (string, bool) {
	l := w.Layout()
	// The panel the offers are drawn on, which for a list is its own and not
	// the one everything else uses: drawn in one place and listened for in
	// another, three words looked like buttons and answered nowhere.
	panel := Over(w.width, w.height)
	if w.screen == OnChoosing {
		panel = ChoosingOver(w.width, w.height)
	}
	for i, offer := range w.Offers() {
		if OfferAt(l, i).In(p) {
			return offer, true
		}
		if w.screen != OnTheBoard && PanelOfferAt(panel, i).In(p) {
			return offer, true
		}
	}
	return "", false
}

// Question is what the one question screen is asking about.
type Question int

// The two things worth asking before doing: both lose work, and work lost
// without being asked is work somebody has to remember they lost.
const (
	aboutDeleting Question = iota
	aboutLeaving
	aboutSweeping
)

// sweeps is the column a sweep would take off the board: the one the chosen
// card stands in. Nothing when nothing is chosen, which is when there is no
// column to name and so nothing to ask about.
func (w Window) sweeps() string {
	if w.chosen == 0 || w.chosenGoal {
		return ""
	}
	state := w.stateOf(w.chosen)
	if reading.Stopped(state) {
		return ""
	}
	return state
}

// sweeping takes every task in that column off the board.
func (w Window) sweeping() Window {
	state := w.sweeps()
	w.screen, w.question = OnTheBoard, aboutDeleting
	if state == "" {
		return w
	}
	_, err := w.work.Ask(engine.TaskArchive, engine.Input{Given: map[string]string{"state": state}})
	if err != nil {
		w.err = err
		return w
	}
	return w.Look()
}

// Hovering answers the window the pointer resting at a point leaves behind.
func (w Window) Hovering(p Point) Window {
	w.over, w.hovered = w.offerAt(p)
	return w
}

// Away is the window the pointer having left it leaves behind.
func (w Window) Away() Window {
	w.over, w.hovered = "", false
	return w
}

// Hovered is the offer the pointer is resting on.
func (w Window) Hovered() (string, bool) { return w.over, w.hovered }

// Lighting says whether an offer is the one the pointer is resting on, which
// is the one thing in the window that says a word can be clicked. Asked here
// rather than worked out in the drawing, so a scenario can put the question.
func (w Window) Lighting(offer string) bool { return w.hovered && w.over == offer }

// goalAt says which objective a point lands on in the list.
func (w Window) goalAt(p Point) int {
	l := w.Layout()
	here := reading.Sorted(w.Showing().Goals)
	for i := range here {
		if GoalAt(l, i, w.goals).In(p) {
			return here[i].Number
		}
	}
	return 0
}

// tabAt says which tab a point falls on, and whether it falls on one at all.
// The strip is wider than the two tabs together, so clicking beside them is
// clicking nothing.
func (w Window) tabAt(p Point) (Tab, bool) {
	l := w.Layout()
	for _, tab := range []Tab{Tasks, Objectives} {
		if TabAt(l, int(tab)).In(p) {
			return tab, true
		}
	}
	return w.tab, false
}

// DoubleClick opens the detail of the card under the pointer, which is where
// the description and the history are.
func (w Window) DoubleClick(p Point) Window {
	// In a list it takes the row and is done with it, which is the gesture a
	// hand already has: one click says which row, two say that one.
	if w.screen == OnChoosing {
		box := ChoosingOver(w.width, w.height)
		if at, on := ChoosingAtPoint(box, len(w.Choosing()), w.choosingFrom, p); on {
			w.choosing = at
			w = w.taking()
			if w.screen == OnChoosing {
				return w.keeping()
			}
		}
		return w
	}
	// The same two screens a single click answers on, for the same reason: a
	// search leaves the board showing, so a card under the pointer is a card.
	if w.screen != OnTheBoard && w.screen != OnSearch {
		return w
	}
	number := w.at(p)
	if number == 0 {
		return w
	}
	w.chosen, w.chosenGoal = number, w.tab == Objectives
	return w.opening()
}

// PressAt takes hold of whatever is under the pointer, without deciding yet
// whether this is a click or a drag.
func (w Window) PressAt(p Point) Window {
	w.pressed, w.pressedOn = true, w.screen
	if w.screen == OnDetail {
		w.draggingBar = HistoryBar(Over(w.width, w.height)).In(p)
		w.draggingServing = ServingBar(Over(w.width, w.height), w.Shown()).In(p)
		if at, long := LongField(w.Shown()); long {
			w.draggingSaying = SayingBar(Over(w.width, w.height), w.Shown(), at).In(p)
		}
		// A tile is taken hold of the way a card is: solidary with its parent
		// does not mean fixed among its sisters, and the order is what says
		// which part is wanted first.
		if at, on := w.tileUnder(p); on {
			w.held = Held{Number: w.Under()[at].Number, From: p, At: p}
			w.underAt = at
		}
		return w
	}
	// A press in a field is where a run of text may be held from, so the caret
	// goes where the press landed and not to the field's end: the end is where
	// somebody arriving at a field types, and this one is not arriving.
	if w.screen == OnChoosing {
		w.draggingChoosing = ChoosingBar(ChoosingOver(w.width, w.height)).In(p)
		return w
	}
	if w.screen == OnForm {
		// The bar beside the long field the keyboard is on is taken hold of,
		// and a press on it is not a press in the text.
		if _, bar, _, _, long := w.sayingOn(); long && bar.In(p) {
			w.draggingSaying = true
			return w
		}
		if at, on := w.fieldAt(p); on {
			w.form = w.form.onto(at)
			w.form = w.form.Writes(w.form.Writing().CaretAt(w.reaching(at, p)))
			w.holdingText = true
			return w.inSightOfTheCaret()
		}
		return w
	}
	if w.screen != OnTheBoard {
		return w
	}
	// Nothing is taken hold of on the objectives page. An objective and a task
	// can carry the same number, and what is dragged is let go into a column —
	// which only a task has. Dragging one used to move the other.
	if w.tab == Objectives {
		return w
	}
	w.held = Held{Number: w.at(p), From: p, At: p}
	return w
}

// MoveTo follows the pointer, and past Grip says the card is being dragged.
func (w Window) MoveTo(p Point) Window {
	if w.draggingBar {
		track := HistoryBar(Over(w.width, w.height))
		w.history = w.settledHistory(Reached(track, w.Told(), w.Rows(), p))
		return w
	}
	// A pointer held down in a field holds a run of its text.
	if w.holdingText && w.screen == OnForm {
		if at, on := w.fieldAt(p); on && at == w.form.On() {
			w.form = w.form.Writes(w.form.Writing().Selects(w.reaching(at, p)))
		}
		return w
	}
	if w.draggingSaying {
		_, track, lines, rows, _ := w.sayingOn()
		w.saying = w.settledSaying(Reached(track, len(lines.Said), rows, p))
		return w
	}
	if w.draggingChoosing {
		box := ChoosingOver(w.width, w.height)
		w.choosingFrom = w.settledChoosing(
			Reached(ChoosingBar(box), len(w.Choosing()), ChoosingRows(box), p))
		return w
	}
	if w.draggingServing {
		// Counted in tiles and settled in pixels, since that box is scrolled
		// by pixels and not by lines.
		track := ServingBar(Over(w.width, w.height), w.Shown())
		room := max(1, ServingBox(Over(w.width, w.height), w.Shown()).H/SubHeight)
		w.serving = w.settledServing(Reached(track, len(w.Under()), room, p) * SubHeight)
		return w
	}
	if w.held.Number == 0 {
		return w
	}
	w.held.At = p
	if abs(p.X-w.held.From.X) > Grip || abs(p.Y-w.held.From.Y) > Grip {
		w.held.Moved = true
	}
	return w
}

// LetGo drops the card, or lets a click be a click.
func (w Window) LetGo(p Point) Window {
	if w.draggingBar || w.draggingSaying || w.draggingServing || w.draggingChoosing {
		w.draggingBar, w.draggingSaying = false, false
		w.draggingServing, w.draggingChoosing = false, false
		return w
	}
	if w.holdingText {
		// Let go where it was let go. A run held with the mouse is finished by
		// lifting it, and the click that would follow would collapse the run
		// it just made.
		w.holdingText = false
		if w.form.Writing().Holding() {
			return w
		}
	}
	// A release with no press behind it on this screen is not a click here.
	// A double click opens a task on the press, and the release that follows
	// landed on the board it had just left — closing what it had just opened,
	// for anyone whose pointer was on the part of a card the panel does not
	// cover.
	waiting, was := w.pressed, w.pressedOn
	w.pressed = false
	if !waiting || was != w.screen {
		w.held = Held{}
		return w
	}
	if w.screen == OnDetail {
		if w.held.Moved {
			return w.droppedUnder(p)
		}
		// Nothing was dragged, so this is a click, and it goes where every
		// other click goes. Answered here instead, it left the panel deaf to
		// the mouse — which is the fault the comment below was written about,
		// made again on a screen it did not name.
		w.held = Held{}
		return w.Click(p)
	}
	held := w.held
	w.held = Held{}
	if held.Number != 0 && held.Moved {
		return w.dropped(held.Number, p)
	}
	// Everything else is a click, including a press that took hold of nothing.
	// A tab is not a card and neither is an offer in the bar: making this the
	// card's business is what left both unreachable in a real window, where
	// every click arrives as a press and then a release.
	return w.Click(p)
}

// turning adds what a wheel sent to what the turns before it left over, and
// answers whole lines and the new leftover. A trackpad sends a few pixels at a
// time: taken one by one they each round to nothing and are thrown away, and
// added up they are a line like any other.
func turning(spare, by int) (lines, left int) {
	total := spare + by
	return total / lineHigh, total % lineHigh
}

// Scroll moves the column under the pointer, and no other.
func (w Window) Scroll(p Point, by int) Window {
	// A panel stops the wheel. Over a form, the help or the question it used to
	// reach the board behind, and scrolled something nobody could see. What
	// it reaches on a form is the long field the keyboard is on, the way it
	// reaches the description on an open task.
	if w.screen == OnForm {
		return w.scrollingSaying(p, by)
	}
	if w.screen == OnHelp || w.screen == OnAsking {
		return w
	}
	if w.screen == OnChoosing {
		box := ChoosingOver(w.width, w.height)
		if ChoosingRun(box).In(p) || ChoosingBar(box).In(p) {
			lines, left := turning(w.spareChoosing, by)
			w.choosingFrom, w.spareChoosing = w.settledChoosing(w.choosingFrom+lines), left
		}
		return w
	}
	if w.screen == OnDetail {
		// The panel is over the board, so the wheel over it moves the history
		// and never a column behind it.
		if HistoryBox(Over(w.width, w.height)).In(p) {
			lines, left := turning(w.spareHistory, by)
			w.history, w.spareHistory = w.settledHistory(w.history+lines), left
		}
		if ServingBox(Over(w.width, w.height), w.Shown()).In(p) {
			w.serving = w.settledServing(w.serving + by)
		}
		return w.scrollingSaying(p, by)
	}
	l := w.Layout()
	if w.tab == Objectives {
		// One grid across the four columns' worth of width, so anywhere over
		// the board is over the objectives.
		if GoalsRun(l).In(p) {
			w.goals = w.settledGoals(w.goals + by)
		}
		return w
	}
	for i := range Columns {
		if !l.Column[i].In(p) {
			continue
		}
		w.scroll[i] = w.settled(i, w.scroll[i]+by)
		return w
	}
	return w
}

// Key is the few flat shortcuts this board keeps. There are no modes: what a
// person points at is what happens, and these are the things a pointer is bad
// at saying.
func (w Window) Key(name string) Window {
	// A key does what the bar offers, where the bar offers it, and nothing
	// where the bar says nothing. One rule and not a list of guards: the bar
	// used to offer Edit on an open task while e did nothing there, and a key
	// that adds used to work over an open form and throw away what was in it.
	//
	// Read off the offers rather than written out beside them, so an offer
	// added to one screen cannot arrive without its key.
	if name != "ctrl+n" && name != "cmd+f" && name != "ctrl+f" {
		for _, offer := range w.Offers() {
			if Shortcut(offer) == name {
				return w.Choose(offer)
			}
		}
	}
	// Taking a task up is not offered on the bar: the bar has room for what
	// changes a task, and this changes what somebody is working on. It is in
	// the help, under the same key the terminal board uses.
	if name == "f" && w.screen == OnTheBoard && w.chosen != 0 && !w.chosenGoal {
		return w.working()
	}
	// The whole column off the board at once. Not on the bar either: the bar
	// has room for what one card does, and this does what a column does. The
	// engine has swept by a narrowing all along and only the command line ever
	// asked it to.
	if name == "shift+x" && w.screen == OnTheBoard && w.tab == Tasks && w.sweeps() != "" {
		w.screen, w.question = OnAsking, aboutSweeping
		return w
	}
	// Half a box at a time through the description on an open task: the
	// terminal board's keys for it, taken here too so the same fingers work.
	if w.screen == OnDetail && (name == "cmd+d" || name == "cmd+u") {
		_, _, _, rows, ok := w.sayingOn()
		if !ok {
			return w
		}
		by := max(1, rows/2)
		if name == "cmd+u" {
			by = -by
		}
		w.saying = w.settledSaying(w.saying + by)
		return w
	}
	// The two spellings of a key the bar has no room to write twice.
	switch name {
	case "ctrl+n":
		return w.Key("cmd+n")
	case "cmd+f", "ctrl+f":
		return w.Key("/")
	}
	// Escape is not here: every screen that has a way out offers it as Close
	// or Cancel, and Shortcut maps both to esc — so it is served by the loop
	// above before it ever reaches this far. The branch that used to close a
	// search here was unreachable, and the only thing that noticed was
	// mutation, which found arithmetic no test could run.
	if w.screen == OnChoosing {
		return w.inChoosing(name)
	}
	if w.screen == OnForm {
		return w.inForm(name).inSightOfTheCaret()
	}
	if w.screen == OnTheBoard {
		switch name {
		case "tab", "shift+tab":
			return w.other()
		case "left":
			return w.walking(-1, 0)
		case "right":
			return w.walking(1, 0)
		case "up":
			return w.walking(0, -1)
		case "down":
			return w.walking(0, 1)
		case "enter", "return":
			return w.opening()
		case "shift+left":
			return w.shifting(-1, 0)
		case "shift+right":
			return w.shifting(1, 0)
		case "shift+up":
			return w.shifting(0, -1)
		case "shift+down":
			return w.shifting(0, 1)
		}
	}
	if w.screen == OnDetail {
		// A hand on the keyboard reads a history the same way it reads a
		// board: with the arrows, without reaching for the mouse. Over an
		// objective they walk the work under it instead — that is the list
		// there is to read, and the history column carries one line saying an
		// objective has none.
		// The box takes the arrows wherever there is one, because it is what
		// somebody acts on: a tile is marked done, and a line of history is
		// read rather than changed. The history keeps the wheel over it and
		// the bar beside it.
		if len(w.Under()) > 0 {
			switch name {
			case "down":
				return w.walkingUnder(1)
			case "up":
				return w.walkingUnder(-1)
			case " ", "space":
				return w.marking()
			}
			return w
		}
		switch name {
		case "down":
			w.history = w.settledHistory(w.history + 1)
		case "up":
			w.history = w.settledHistory(w.history - 1)
		}
		return w
	}
	if w.screen == OnSearch {
		switch name {
		case "backspace":
			w.looking = w.looking.Back()
		case "left":
			return w.walks(-1)
		case "right":
			return w.walks(1)
		case "enter", "return":
			// The box closes and what it narrowed stays narrowed. A search
			// that vanished the moment it was asked for would be a search
			// nobody could act on.
			w.screen = OnTheBoard
		}
	}
	return w
}

// inForm is what a key does while the form has the window. Everything here is
// about one field, because a form is a column of fields and the keyboard is on
// one of them.
func (w Window) inForm(name string) Window {
	long := w.form.Fields()[w.form.On()].Long
	switch name {
	case "tab":
		w.form = w.form.onto(w.form.On() + 1)
	case "shift+tab":
		w.form = w.form.onto(w.form.On() - 1)
	case "down", "up":
		// Inside the description the arrows walk its lines, so every one of
		// them can be reached. Down used to leave for the next field, and the
		// only way through a long text was a character at a time — which is
		// why it read as cut off. Tab is what leaves it.
		by := 1
		if name == "up" {
			by = -1
		}
		if long {
			return w.walksLines(by)
		}
		w.form = w.form.onto(w.form.On() + by)
	case "backspace":
		return w.back()
	case "cmd+a":
		w.form = w.form.Writes(w.form.Writing().All())
	case "cmd+c":
		w.copying = w.form.Writing().Holds()
	case "cmd+x":
		w.copying = w.form.Writing().Holds()
		w.form = w.form.Writes(w.form.Writing().Back())
	case "cmd+v":
		// Asked for here and answered later: a clipboard is read by the
		// system, which takes as long as it takes, and the answer comes back
		// as an event like any other.
		w.pasting = true
	case "cmd+s":
		// The terminal's save, which the window takes too: a terminal cannot
		// be sent ⌘, so the board saves on ^S, and the same fingers should
		// work in both.
		return w.Save()
	case "shift+left":
		return w.holding(-1)
	case "shift+right":
		return w.holding(1)
	case "left":
		return w.walks(-1)
	case "right":
		return w.walks(1)
	case "enter", "return", "shift+enter", "shift+return":
		// A line where a line can be written, and on to the next field
		// everywhere else — the same on both boards. It never saves: saving
		// is a chord of its own, so it cannot happen by reaching for the key
		// that breaks a line.
		if long {
			w.form = w.form.Writes(w.form.Writing().Types("\n"))
			return w.inSightOfTheCaret()
		}
		w.form = w.form.onto(w.form.On() + 1)
	}
	return w
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

// inSightOfTheCaret scrolls the long field the keyboard is on just far enough
// for the caret's line to be in the box. The wheel moves the box on its own; the
// caret moving brings it back.
func (w Window) inSightOfTheCaret() Window {
	_, _, lines, rows, ok := w.sayingOn()
	if !ok {
		return w
	}
	line, _ := lines.At(w.form.Writing().At)
	w.saying = w.settledSaying(inSight(line, 1, rows, w.saying))
	return w
}

// holding walks the caret and leaves the anchor, which is how a run is held
// from the keyboard.
func (w Window) holding(by int) Window {
	writing := w.form.Writing()
	w.form = w.form.Writes(writing.Selects(writing.At + by))
	return w
}

// Copying is what the window wants put on the clipboard, and whether it wants
// anything at all.
func (w Window) Copying() (string, bool) { return w.copying, w.copying != "" }

// Copied says the loop has done it.
func (w Window) Copied() Window {
	w.copying = ""
	return w
}

// Pasting says the window has asked for what is on the clipboard.
func (w Window) Pasting() bool { return w.pasting }

// Pasted hands over what was on it. Empty is still an answer: the clipboard
// held nothing, and the window has stopped waiting either way.
func (w Window) Pasted(text string) Window {
	w.pasting = false
	if text == "" {
		return w
	}
	if w.screen == OnChoosing {
		w.filtering, w.choosing, w.choosingFrom = w.filtering.Types(text), 0, 0
		return w
	}
	if w.screen != OnForm {
		return w
	}
	w.form = w.form.Writes(w.form.Writing().Types(text))
	return w
}

// walking moves what is chosen by the arrows: across the columns, or down the
// cards of one. A hand on the keyboard should not have to reach for the mouse
// to look at the next task.
//
// It stops at the edges rather than wrapping round, and walks past a column
// with nothing in it rather than stopping in one — an empty column is not a
// place to be.
func (w Window) walking(across, down int) Window {
	if w.tab == Objectives {
		return w.walkingGoals(across, down)
	}
	// The band is walked like the columns. A task that is blocked, watching or
	// paused used to be beyond the arrows altogether — and those are the ones
	// that want a decision.
	if lane, at, inBand := w.whereInBand(); inBand {
		return w.walkingBand(lane, at, across, down)
	}
	column, at := w.whereChosen()
	if down > 0 {
		// Off the bottom of a column is into the lane beneath it. The backlog
		// stands alone, full height, with no lane under it.
		if here := w.inColumn(column); column > 0 && at == len(here)-1 {
			if lane := w.inLane(column - 1); len(lane) > 0 {
				w.chosen, w.chosenGoal = lane[0].Number, false
				return w
			}
		}
	}
	if across != 0 {
		for next := column + across; next >= 0 && next < Columns; next += across {
			here := w.inColumn(next)
			if len(here) == 0 {
				continue
			}
			at = min(at, len(here)-1)
			w.chosen, w.chosenGoal = here[at].Number, false
			return w.following(next, at)
		}
		return w
	}
	here := w.inColumn(column)
	if len(here) == 0 {
		return w
	}
	at = min(len(here)-1, max(0, at+down))
	w.chosen, w.chosenGoal = here[at].Number, false
	return w.following(column, at)
}

// shifting moves the chosen card itself rather than the choice: sideways it
// changes column, up and down it changes place among the ones it sits with.
// The terminal has done this from the beginning; the window had only the drag,
// so the central gesture of the product wanted a mouse.
func (w Window) shifting(across, down int) Window {
	if w.tab == Objectives || w.chosen == 0 {
		return w
	}
	column, at := w.whereChosen()
	if across != 0 {
		next := column + across
		if next < 0 || next >= Columns {
			return w
		}
		dealt, _ := reading.Deal(w.Showing().Cards, w.archived)
		if _, err := w.work.Ask(engine.TaskEdit, engine.Input{Given: map[string]string{
			"number": strconv.Itoa(w.chosen), "state": dealt[next].State}}); err != nil {
			w.err = err
			return w
		}
		return w.Look()
	}
	here := w.inColumn(column)
	to := at + down
	if to < 0 || to >= len(here) {
		return w
	}
	given := map[string]string{"number": strconv.Itoa(w.chosen)}
	if to == 0 {
		given["first"] = "true"
	} else {
		given["after"] = strconv.Itoa(here[to].Number)
	}
	if _, err := w.work.Ask(engine.TaskMove, engine.Input{Given: given}); err != nil {
		w.err = err
		return w
	}
	return w.Look()
}

// whereInBand says which lane the chosen task sits in, and where in it.
func (w Window) whereInBand() (lane, at int, ok bool) {
	for i := range Lanes {
		for j, c := range w.inLane(i) {
			if c.Number == w.chosen {
				return i, j, true
			}
		}
	}
	return 0, 0, false
}

// walkingBand moves inside the band, and back up into the column a lane sits
// under. The lanes are one row, so down walks along a lane and up leaves it.
func (w Window) walkingBand(lane, at, across, down int) Window {
	if down < 0 && at == 0 {
		if here := w.inColumn(lane + 1); len(here) > 0 {
			last := len(here) - 1
			w.chosen, w.chosenGoal = here[last].Number, false
			return w.following(lane+1, last)
		}
		return w
	}
	if across != 0 {
		for next := lane + across; next >= 0 && next < len(Lanes); next += across {
			if here := w.inLane(next); len(here) > 0 {
				w.chosen, w.chosenGoal = here[min(at, len(here)-1)].Number, false
				return w
			}
		}
		return w
	}
	here := w.inLane(lane)
	w.chosen, w.chosenGoal = here[min(len(here)-1, max(0, at+down))].Number, false
	return w
}

// following moves a column so that the card just chosen is in sight. Walking
// used to move the choice and leave the column where it was, which chose a
// card nobody could see — and a card chosen out of sight is a card whose
// title, whose state and whose place are all unknown to whoever chose it.
func (w Window) following(column, at int) Window {
	run := w.Layout().Column[column]
	w.scroll[column] = w.settled(column, inSight(at, CardHeight, run.H, w.scroll[column]))
	return w
}

// followingGoal is the same for the objectives, which are one list rather than
// four columns and so have one number of their own.
func (w Window) followingGoal(at int) Window {
	run := GoalsRun(w.Layout())
	w.goals = w.settledGoals(inSight(at/GoalsAcross(), CardHeight, run.H, w.goals))
	return w
}

// settledGoals keeps the objectives between their two ends.
func (w Window) settledGoals(want int) int {
	run := GoalsRun(w.Layout())
	return min(max(0, GoalRows(len(w.Showing().Goals))*CardHeight-run.H), max(0, want))
}

// walkingGoals is the same down a list, which is what the objectives are.
func (w Window) walkingGoals(across, down int) Window {
	here := reading.Sorted(w.Showing().Goals)
	if len(here) == 0 {
		return w
	}
	at := 0
	for i, g := range here {
		if g.Number == w.chosen {
			at = i
		}
	}
	// Across moves to the next objective, down to the one a row below: the
	// objectives are a grid, and an arrow means in a grid what it looks like
	// it means.
	at = min(len(here)-1, max(0, at+across+down*GoalsAcross()))
	w.chosen, w.chosenGoal = here[at].Number, true
	return w.followingGoal(at)
}

// WhereChosen is where what is chosen is drawn and the run of the window it is
// drawn in, so that a test can ask whether one is inside the other.
func (w Window) WhereChosen() (box, run Box, ok bool) {
	l := w.Layout()
	if w.tab == Objectives {
		for i, g := range reading.Sorted(w.Showing().Goals) {
			if g.Number == w.chosen {
				return GoalAt(l, i, w.goals), GoalsRun(l), true
			}
		}
		return Box{}, Box{}, false
	}
	dealt, _ := reading.Deal(w.Showing().Cards, w.archived)
	for i, held := range dealt {
		for j, c := range held.Cards {
			if c.Number == w.chosen {
				return l.CardAt(i, j, w.scroll[i]), l.Column[i], true
			}
		}
	}
	return Box{}, Box{}, false
}

// whereChosen is which column what is chosen stands in and how far down it, and
// the first card of the first column that has one when nothing is chosen — an
// arrow on an untouched board has to begin somewhere.
func (w Window) whereChosen() (column, at int) {
	// What is walked is what is showing: under a search, the arrows reach
	// what the search left and nothing behind it.
	dealt, _ := reading.Deal(w.Showing().Cards, w.archived)
	for i, held := range dealt {
		for j, c := range held.Cards {
			if c.Number == w.chosen {
				return i, j
			}
		}
	}
	for i, held := range dealt {
		if len(held.Cards) > 0 {
			return i, 0
		}
	}
	return 0, 0
}

// opening is what enter does: it opens whatever the arrows walked to, the way a
// double-click opens whatever the pointer landed on.
func (w Window) opening() Window {
	if w.chosen == 0 {
		return w
	}
	w.screen = OnDetail
	// Whatever is opened is opened at its top: its description, the work under
	// it, its history. The description's scroll was left where the last task or
	// the form had put it, so a short description opened after a long one had
	// been read to its end showed an empty box until the wheel moved.
	w.underAt, w.history, w.serving, w.saying = 0, 0, 0, 0
	w.spareHistory, w.spareSaying = 0, 0
	// An objective is opened on the work under it and not on a history: it has
	// none of its own, and the number it carries belongs to a task as well —
	// asking for the trace of that number answered with a task nobody had
	// chosen, which is the wrong answer said confidently.
	if w.chosenGoal {
		w.acts = nil
		return w
	}
	shown, err := w.work.Ask(engine.TraceShow, reading.Numbered(w.chosen))
	if err != nil {
		w.err = err
		return w
	}
	w.acts = nil
	for _, act := range shown.Acts {
		w.acts = append(w.acts, line(act))
	}
	return w
}

// Wonder asks the store whether anybody else has written to it, and reads the
// whole thing again only when the answer has moved. Every listing is two
// questions and a redraw; the version is one integer.
func (w Window) Wonder() Window {
	version, err := w.work.Version()
	if err != nil {
		w.err = err
		return w
	}
	if version == w.version {
		return w
	}
	return w.Look()
}

// Look reads the work again and draws from what came back.
func (w Window) Look() Window {
	found, err := reading.Look(w.work)
	if err != nil {
		w.err = err
		return w
	}
	w.snap, w.version, w.err = found, found.Version, nil
	for i := range Columns {
		w.scroll[i] = w.settled(i, w.scroll[i])
	}
	return w
}

// at says which task is under a point, in a column or in one of the band's
// lanes, and zero when there is none.
func (w Window) at(p Point) int {
	// The objectives tab is a list across the whole board, and its rows lie
	// over the columns behind them. Asking the columns there answers whichever
	// task happens to sit at the same height, which is not what was pointed at.
	if w.tab == Objectives {
		return w.goalAt(p)
	}
	l := w.Layout()
	if column, _, found := l.Under(p, 0); found {
		here := w.inColumn(column)
		index := (p.Y - l.Column[column].Y + w.scroll[column]) / CardHeight
		if index < len(here) {
			return here[index].Number
		}
		return 0
	}
	if lane, found := l.LaneUnder(p); found {
		here := w.inLane(lane)
		// A lane is not scrolled: it holds what is not moving, two rows of it,
		// and a third row is a sign to go and unblock something rather than a
		// sign to scroll.
		index := (p.Y - l.Lane[lane].Y) / CardHeight
		if index < len(here) {
			return here[index].Number
		}
	}
	return 0
}

// dropped is what a card let go somewhere means. A drop says two things at
// once — which column, and which place among the cards there — and the engine
// is told them in that order, because it refuses to do both in one act.
//
// There is no transaction between the two. If the second is refused, the task
// has changed column and not place, and the board says so when it looks again:
// visible rather than silent.
func (w Window) dropped(number int, p Point) Window {
	l := w.Layout()
	state, place, found := w.where(l, p)
	if !found {
		return w
	}
	if state != w.stateOf(number) {
		asked := engine.Input{Given: map[string]string{
			"number": strconv.Itoa(number),
			"state":  state,
		}}
		if _, err := w.work.Ask(engine.TaskEdit, asked); err != nil {
			w.err = err
			return w
		}
	}
	if len(place) > 0 {
		given := map[string]string{"number": strconv.Itoa(number)}
		for name, value := range place {
			given[name] = value
		}
		if _, err := w.work.Ask(engine.TaskMove, engine.Input{Given: given}); err != nil {
			w.err = err
			return w
		}
	}
	return w.Look()
}

func (w Window) where(l Layout, p Point) (state string, place map[string]string, found bool) {
	if lane, in := l.LaneUnder(p); in {
		return Lanes[lane], nil, true
	}
	for i := range Columns {
		if !l.Column[i].In(p) {
			continue
		}
		here := w.inColumn(i)
		at := (p.Y - l.Column[i].Y + w.scroll[i]) / CardHeight
		if at <= 0 {
			return reading.StateAt(i), map[string]string{"first": "true"}, true
		}
		if at-1 < len(here) {
			return reading.StateAt(i), map[string]string{"after": strconv.Itoa(here[at-1].Number)}, true
		}
		return reading.StateAt(i), map[string]string{"last": "true"}, true
	}
	return "", nil, false
}

// inColumn and inLane are the cards of one column and of one lane, in the order
// the engine keeps them.
func (w Window) inColumn(column int) []reading.Card {
	dealt, _ := reading.Deal(w.Showing().Cards, w.archived)
	return dealt[column].Cards
}

func (w Window) inLane(lane int) []reading.Card {
	_, stalled := reading.Deal(w.Showing().Cards, w.archived)
	var here []reading.Card
	for _, c := range stalled {
		if c.State == Lanes[lane] {
			here = append(here, c)
		}
	}
	return here
}

// stateOf is the column a task stands in now.
func (w Window) stateOf(number int) string {
	for _, c := range w.snap.Cards {
		if c.Number == number {
			return c.State
		}
	}
	return ""
}

// settled keeps a column's scroll between its two ends: never above the first
// card, never past what there is to see.
func (w Window) settled(column, want int) int {
	l := w.Layout()
	most := max(0, len(w.inColumn(column))*CardHeight-l.Column[column].H)
	return min(most, max(0, want))
}

// line is one act of a task's history, written the way a person reads it. Read
// off the act's own membrane rather than out of its fields, which is the only
// way in.
func line(act trace.Act) string {
	told := trace.OfAct(act)
	written := told.At + "  " + told.Did
	if told.Field != "" {
		written += " " + told.Field
	}
	if told.Became != "" {
		written += " → " + told.Became
	}
	if told.Through != "" {
		written += "  (" + told.Through + ")"
	}
	return written
}

// abs is the distance between two points on one axis. Written without a branch:
// a branch here has a boundary at zero that both sides agree on, so a wrong
// comparison would go out unjudged.
func abs(n int) int { return max(n, -n) }
