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
	"time"

	"cairn/internal/reading"
	"cairn/internal/trace"
)

// Screen is what has the window.
type Screen int

const (
	// OnTheBoard is the board itself, which is where a window opens.
	OnTheBoard Screen = iota
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
	// OnCalendar is the calendar a deadline is chosen on, over the form.
	OnCalendar
	// OnMenu is what a right click offers, beside the pointer.
	OnMenu
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
	// lanes is the same for the three lanes of the band. They were left
	// unscrolled once, on the grounds that a third stopped task is a sign to
	// go and unblock something — and a lane holding more than two then hid
	// the rest where neither the wheel nor the arrows could reach them.
	lanes [3]int
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
	spareHistory, spareSaying, spareChoosing, spareHelp int
	// helpFrom is the first line of the help in sight, see help.go.
	helpFrom int
	// measure is the face text is measured with; nil until one is given.
	measure Measure
	// keyboard is what keys are written for, see keyboard.go.
	keyboard Keyboard
	// now is the moment the window was last told it is, and day the day the
	// calendar's keyboard is on: see calendar.go.
	now time.Time
	day time.Time
	// menu is what a right click offered, menuAt where it was clicked and
	// menuOn the line the keyboard is on: see menu.go.
	menu   []string
	menuAt Point
	menuOn int
	// paused is what the open form held when the window last looked at the
	// store, see autosave.go.
	paused []string
	// growing is what a run held with the pointer grows by, and anchor the run
	// it grows from: what the press, the double or the triple click held.
	growing Unit
	anchor  Picked
	// choosingFor is the field the open list is choosing for, and filtering
	// what is being typed into it to narrow that list.
	choosingFor string
	filtering   Written
	// choosingFrom is the first row of the list in sight. A list of everything
	// a task may hang under is as long as the board, and a panel is not.
	choosingFrom int
	// labelsOffered is every label the open list of labels offers, numbered by
	// where it stands here: those some task or objective carries, and those
	// created since the list opened, last.
	labelsOffered []string
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
	// picked is the run of text held on an open task, and pickingText that a
	// press began in a piece of text and is still down.
	picked      Picked
	pickingText bool
	// wheelOn is the screen the turn of the wheel under way began on, and
	// wheelAt when its last part arrived: the rest of a turn belongs to what it
	// began on, and never to what opens under it while it is still coming.
	wheelOn Screen
	wheelAt time.Time
	// question is which of the two things the one question screen is asking
	// about. There used to be only one.
	question Question
	// draft is the last form thrown away with something typed in it, set
	// aside so the form of the same task can bring it back. esc twice leaves
	// a form without a question, and this is what keeps that from losing work.
	draft Form
	// taken is what is taken together, gathered with ⌘ and stretched with ⇧,
	// in the order it was taken; gathering and stretching say which of the
	// two keys is held while the pointer acts.
	taken                 []int
	gathering, stretching bool
	// wasScreen is what was open when the window became too small to draw it,
	// so growing back gives it back rather than an empty board.
	wasScreen Screen
	// ticked is the objectives the open list has in, by number, and choosing
	// is where in that list the keyboard sits. They live on the window and not
	// in the form because the list is a screen of its own; the form is given
	// what they came to only when the list closes.
	ticked   map[int]bool
	choosing int

	err error
	// wrongFor is how many times the window has looked at the store since err
	// was said: what went wrong goes on its own once it has been there long
	// enough to read, see Lingers.
	wrongFor int
}

// Lingers is how many beats what went wrong stays: long enough to read two
// lines, and gone before it is mistaken for something still true.
func Lingers() int { return 4 }

// New builds a window over whatever holds the work.
func New(work reading.Work, width, height int) Window {
	w := Window{work: work, width: width, height: height, keyboard: thisKeyboard()}
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

// Context is where what was written about the task a form is open on lives
// elsewhere, or nothing: no form, a new task, an objective, or a window not
// told where the notes are.
func (w Window) Context() string {
	if w.screen != OnForm || w.form.Adding || w.form.ForObjective {
		return ""
	}
	c, _ := w.snap.Card(w.form.Number)
	return c.Context
}

// ContextSays is what the foot of an open task writes about where its context
// lives: its name and the place, or nothing when there is no place to say.
// One list for the drawing and for what a pointer can hold, so what is drawn
// is what can be copied.
func (w Window) ContextSays() []Readable {
	at := w.Context()
	if at == "" {
		return nil
	}
	box := Over(w.width, w.height)
	return []Readable{one("name:context", ContextName(box), "context", small),
		one("field:context", ContextPath(box), at, small)}
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
	// How many are left and how many are done, beside what the box is: the
	// question somebody opens a task with parts to ask.
	if tally := reading.Tallied(w.Under()); tally != "" {
		return hangingHead + "   " + tally
	}
	return hangingHead
}

// Wonder asks the store whether anybody else has written to it, and reads the
// whole thing again only when the answer has moved. Every listing is two
// questions and a redraw; the version is one integer.
func (w Window) Wonder() Window {
	w = w.aged().autosaved()
	version, err := w.work.Version()
	if err != nil {
		w.err, w.wrongFor = err, 0
		return w
	}
	if version == w.version {
		return w
	}
	return w.Look()
}

// aged counts one more beat on what went wrong, and forgets it once it has
// lingered its time.
func (w Window) aged() Window {
	if w.err == nil {
		return w
	}
	w.wrongFor++
	if w.wrongFor >= Lingers() {
		w.err, w.wrongFor = nil, 0
	}
	return w
}

// Look reads the work again and draws from what came back.
func (w Window) Look() Window {
	found, err := reading.Look(w.work)
	if err != nil {
		w.err, w.wrongFor = err, 0
		return w
	}
	w.snap, w.version, w.err = found, found.Version, nil
	for i := range Columns {
		w.scroll[i] = w.settled(i, w.scroll[i])
	}
	for i := range w.lanes {
		w.lanes[i] = w.settledLane(i, w.lanes[i])
	}
	return w
}

// inColumn and inLane are the cards of one column and of one lane, in the order
// the engine keeps them.
func (w Window) inColumn(column int) []reading.Card {
	dealt, _ := reading.Deal(w.Showing().Cards, w.archived)
	return dealt[column].Cards
}

func (w Window) inLane(lane int) []reading.Card {
	_, stalled := reading.Deal(w.Showing().Cards, w.archived)
	return reading.InLanes(stalled)[lane]
}

// stateOf is the column a task stands in now.
func (w Window) stateOf(number int) string {
	c, _ := w.snap.Card(number)
	return c.State
}

// settled keeps a column's scroll between its two ends: never above the first
// card, never past what there is to see.
func (w Window) settled(column, want int) int {
	l := w.Layout()
	most := max(0, len(w.inColumn(column))*CardHeight-l.Column[column].H)
	return min(most, max(0, want))
}

// settledLane is the same for a lane of the band.
func (w Window) settledLane(lane, want int) int {
	most := max(0, len(w.inLane(lane))*CardHeight-w.Layout().Lane[lane].H)
	return min(most, max(0, want))
}

// line is one act of a task's history, written the way a person reads it —
// see reading.SaidOf.
func line(act trace.Act) string { return reading.SaidOf(act).Line() }

// abs is the distance between two points on one axis. Written without a branch:
// a branch here has a boundary at zero that both sides agree on, so a wrong
// comparison would go out unjudged.
func abs(n int) int { return max(n, -n) }
