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
	spareHistory, spareSaying int
	// serving is how far the work under an open objective has been scrolled,
	// counted in lines the way the history is: a task wraps onto as many lines
	// as its title needs, and a list that jumped a task at a time would skip
	// past what it wrapped.
	serving int
	// draggingSaying says the pointer has taken hold of the bar beside the
	// description, which is pulled the way the history's is.
	draggingSaying bool
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
		w.screen = TooSmallToDraw
		return w
	}
	if w.screen == TooSmallToDraw {
		w.screen = OnTheBoard
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

// Saying is how far the description on an open task has been scrolled.
func (w Window) Saying() int { return w.saying }

// SaidLines is the description as it is laid out in its box, and SaidRows how
// many of them are in sight.
func (w Window) SaidLines() []string {
	at, long := LongField(w.Shown())
	if !long {
		return nil
	}
	fields := w.Shown()
	return Wrapped(w.FieldSays(fields[at]), room(SayingBox(Over(w.width, w.height), fields, at))).Said
}

// SaidRows is how many lines of the description its box holds.
func (w Window) SaidRows() int { return longRows }

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
	// A form's own fields answer, because that is how the keyboard gets to one
	// with a mouse.
	if at, on := w.fieldAt(p); on {
		w.form = w.form.onto(at)
		_, value := FieldRow(Over(w.width, w.height), FieldTop(Over(w.width, w.height), w.form.Fields(), at))
		return w.CaretInto(into(value, p))
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
func Heads(w Window) string {
	switch w.screen {
	case OnHelp:
		return "What everything does"
	case OnAsking:
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
	case "edit":
		return "e"
	case "delete":
		return "d"
	case "save", "choose":
		return "enter"
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
		{Name: "e", Value: "edit what is chosen, on the board or on an open task"},
		{Name: "f", Value: "take up what is chosen, or let it go — several at once is fine"},
		{Name: "d", Value: "delete what is chosen — it asks first"},
		{Name: "x", Value: "take what is chosen off the board, or put it back"},
		{Name: "⇧A", Value: "what was archived, on the board or off it — it reads faintly"},
		{Name: "⌘N", Value: "add"},
		{Name: "/ or ⌘F", Value: "search — it reaches the description and the reason"},
		{Name: "⏎ in a search", Value: "keep what it narrowed and go back to the board"},
		{Name: "tab", Value: "the other page — shift+tab the one before"},
		{Name: "tab, ↑↓ in a form", Value: "walk its fields"},
		{Name: "←→", Value: "change a field that has a closed set"},
		{Name: "⏎", Value: "save the form"},
		{Name: "esc", Value: "back out of whatever is open"},
	}
}

// Offers is what the bar says can be done here. It changes with what is on
// screen and with what is chosen, because an offer that does nothing is worse
// than no offer.
func (w Window) Offers() []string {
	switch w.screen {
	case OnChoosing:
		return []string{"Close"}
	case OnForm:
		// The one field a list answers says so on the bar. enter is Save's key
		// everywhere else in a form, and a bar that went on offering Save here
		// while enter opened a list would be a bar that lies — which is the
		// one thing this bar exists not to do. Save comes back the moment the
		// keyboard leaves this field.
		if w.form.Fields()[w.form.On()].Arg == "objective" {
			return []string{"Choose", "Cancel"}
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
		return []string{"Delete", "Cancel"}
	case OnHelp:
		return []string{"Close"}
	case OnSearch:
		// Nothing: the whole bar is the search box, and an offer drawn over it
		// would be a word sitting in the middle of what is being typed.
		return nil
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

// offerAt says which offer a point lands on, in the bar along the bottom or on
// the panel that has the window. Both say the same things and both answer,
// because a person clicks what they are looking at.
func (w Window) offerAt(p Point) (string, bool) {
	l := w.Layout()
	panel := Over(w.width, w.height)
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
	if w.screen != OnTheBoard {
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
	if w.draggingSaying {
		at, _ := LongField(w.Shown())
		track := SayingBar(Over(w.width, w.height), w.Shown(), at)
		w.saying = w.settledSaying(Reached(track, len(w.SaidLines()), w.SaidRows(), p))
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
	if w.draggingBar || w.draggingSaying {
		w.draggingBar, w.draggingSaying = false, false
		return w
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
		if at, long := LongField(w.Shown()); long && SayingBox(Over(w.width, w.height), w.Shown(), at).In(p) {
			lines, left := turning(w.spareSaying, by)
			w.saying, w.spareSaying = w.settledSaying(w.saying+lines), left
		}
		return w
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
	// The two spellings of a key the bar has no room to write twice.
	switch name {
	case "ctrl+n":
		return w.Key("cmd+n")
	case "cmd+f", "ctrl+f":
		return w.Key("/")
	case "esc":
		// Escape out of a search closes it and clears what it narrowed. Out of
		// anything else the bar offers it as Close or Cancel, and it got there
		// through the offers above.
		if w.screen == OnSearch {
			w.screen, w.looking = OnTheBoard, Written{}
			return w
		}
	}
	if w.screen == OnChoosing {
		return w.inChoosing(name)
	}
	if w.screen == OnForm {
		return w.inForm(name)
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
	switch name {
	case "tab", "down":
		w.form = w.form.onto(w.form.On() + 1)
	case "shift+tab", "up":
		w.form = w.form.onto(w.form.On() - 1)
	case "backspace":
		return w.back()
	case "left":
		return w.walks(-1)
	case "right":
		return w.walks(1)
	case "enter", "return":
		return w.Save()
	}
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
	column, at := w.whereChosen()
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
	// An objective is opened on the work under it and not on a history: it has
	// none of its own, and the number it carries belongs to a task as well —
	// asking for the trace of that number answered with a task nobody had
	// chosen, which is the wrong answer said confidently.
	w.underAt = 0
	if w.chosenGoal {
		w.acts, w.history, w.serving, w.saying = nil, 0, 0, 0
		w.spareHistory, w.spareSaying = 0, 0
		return w
	}
	shown, err := w.work.Ask(engine.TraceShow, reading.Numbered(w.chosen))
	if err != nil {
		w.err = err
		return w
	}
	w.acts, w.history = nil, 0
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
