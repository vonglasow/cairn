package gui

import (
	"time"

	"cairn/internal/engine"
	"cairn/internal/migration"
	"cairn/internal/trace"
)

// Screen is what has the window.
type Screen int

const (
	// OnTheBoard is the board itself, which is where a window opens.
	OnTheBoard Screen = iota
	// OnForm is the modal over the board, adding or editing.
	OnForm
	// OnSearch is the board narrowed to what a search reaches, while it is still being typed.
	OnSearch
	// OnAsking is the one question this window asks: deleting loses work, and losing work without being asked is how work is lost.
	OnAsking
	// OnHelp is what everything does, which is the one thing here that has to be read rather than pointed at.
	OnHelp
	// OnChoosing is the list of objectives a task may serve, open over the form.
	OnChoosing
	// OnCalendar is the calendar a deadline is chosen on, over the form.
	OnCalendar
	// OnMenu is what a right click offers, beside the pointer.
	OnMenu
	// OnBackups is the list of backups, one of which may be put back.
	OnBackups
	// OnLabels is the list of every label, one of which may be renamed or taken off everything.
	OnLabels
	// OnFiltering is the board's filters, each walked through its values, the board narrowing as they move.
	OnFiltering
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
	// Time is how long the work took, a week at a time: see week.go.
	Time
)

// Held is a card being dragged: which task, where it was picked up, and where the pointer is now.
type Held struct {
	Number int
	From   Point
	At     Point
	Moved  bool
}

// Grip is how far a pointer travels before a click becomes a drag.
const Grip = 4

// Window is everything the desktop board is, at one moment.
type Window struct {
	work engine.Work

	width, height int
	snap          engine.Snapshot
	version       int64

	screen     Screen
	tab        Tab
	chosen     int
	chosenGoal bool

	scroll                                              [engine.Columns]int
	lanes                                               [3]int
	goals                                               int
	history                                             int
	saying                                              int
	spareHistory, spareSaying, spareChoosing, spareHelp int
	helpFrom                                            int
	measure, measureHeads                               Measure
	keyboard                                            Keyboard
	now                                                 time.Time
	told                                                string
	toldFor                                             int
	native                                              bool
	calendarState
	menuState
	backupsState
	labelsState
	filesState
	timeState
	born                            time.Time
	introOver                       bool
	paused                          []string
	refused                         []string
	growing                         Unit
	anchor                          Picked
	choosingFor                     string
	filtering                       Written
	choosingFrom                    int
	labelsOffered                   []string
	copying                         string
	pasting                         bool
	holdingText                     bool
	serving                         int
	draggingSaying, draggingServing bool
	draggingChoosing                bool
	draggingBar                     bool

	held                  Held
	pressed               bool
	pressedOn             Screen
	acts                  []string
	form                  Form
	formMoved, formGrip   Point
	draggingForm, dragged bool
	sizes                 Sizes
	kept                  SizesKept
	widening              Widening
	filters               engine.Filters
	filterOn              int
	looking               Written
	archived              bool
	aimed                 string
	pages                 engine.Pages
	agreeing              engine.Agreeing
	quitting              bool
	waiting               []string
	hiding                Entry
	starts                string
	over                  string
	hovered               bool
	explaining            int
	resting               string
	restingSlot           Box
	picked                Picked
	pickingText           bool
	wheelOn               Screen
	wheelAt               time.Time
	question              Question
	draft                 Form
	drafts                Drafts
	drafting              []string
	taken                 []int
	gathering, stretching bool
	wasScreen             Screen
	ticked                map[int]bool
	choosing              int

	err      error
	wrongFor int
}

// Lingers is how many beats what went wrong stays: long enough to read two lines, and gone before it is mistaken for something still true.
func Lingers() int { return 4 }

// New builds a window over whatever holds the work.
func New(work engine.Work, width, height int) Window {
	w := Window{work: work, width: width, height: height, keyboard: thisKeyboard(), drafts: noDrafts{}, kept: noSizes{}}
	return w.sized()
}

// Resized answers the window this one becomes at another size.
func (w Window) Resized(width, height int) Window {
	w.width, w.height = width, height
	return w.sized()
}

func (w Window) sized() Window {
	if TooSmall(w.width, w.height) {
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
func (w Window) Layout() Layout {
	l := LaidOut(w.width, w.height, w.sizes)
	l.Days = w.pages.WeekDays()
	return l
}

// Screen says what has the window.
func (w Window) Screen() Screen { return w.screen }

// Chosen is the task picked out, and zero when none is.
func (w Window) Chosen() int { return w.chosen }

// Wrong is what went wrong, if anything did.
func (w Window) Wrong() error { return w.err }

// Dragging says whether a card is in hand and has travelled far enough to be a drag rather than a click that wobbled.
func (w Window) Dragging() bool { return w.held.Number != 0 && w.held.Moved }

// Tab is which of the two is being looked at.
func (w Window) Tab() Tab { return w.tab }

// Context is where what was written about the task a form is open on lives elsewhere, or nothing: no form, a new task, an objective, or a window not told where the notes are.
func (w Window) Context() string {
	if w.screen != OnForm || w.form.Adding || w.form.ForObjective {
		return ""
	}
	c, _ := w.snap.Card(w.form.Number)
	return c.Context
}

// ContextSays is what the foot of an open task writes about where its context lives: its name and the place, or nothing when there is no place to say.
func (w Window) ContextSays() []Readable {
	at := w.Context()
	if at == "" {
		return nil
	}
	box := w.formPanel()
	return []Readable{one("name:context", ContextName(box), "context", small),
		one("field:context", ContextPath(box), at, small)}
}

// ChosenGoal says whether what is chosen is an objective rather than a task.
func (w Window) ChosenGoal() bool { return w.chosenGoal }

// Under is what the box beside the fields of an open objective holds: the work that serves it.
func (w Window) Under() []engine.Card {
	return engine.Serving(w.snap.Cards, w.chosen)
}

// UnderHead is what that box is called.
func (w Window) UnderHead() string { return servingHead }

// Wonder asks the store whether anybody else has written to it, and reads the whole thing again only when the answer has moved.
func (w Window) Wonder() Window {
	w = w.aged().autosaved()
	version, err := w.work.Version()
	if err != nil {
		return w.failing(err)
	}
	if version == w.version {
		return w
	}
	return w.Look()
}

func (w Window) failing(err error) Window {
	w.err, w.wrongFor = err, 0
	return w
}

func lingered(beats int) (int, bool) {
	beats++
	return beats, beats >= Lingers()
}

func (w Window) aged() Window {
	if w.told != "" {
		var gone bool
		if w.toldFor, gone = lingered(w.toldFor); gone {
			w.told, w.toldFor = "", 0
		}
	}
	if w.err != nil {
		var gone bool
		if w.wrongFor, gone = lingered(w.wrongFor); gone {
			w.err, w.wrongFor = nil, 0
		}
	}
	return w
}

// Leaving says whether the window has been asked to close, by a question whose only other way was to leave.
func (w Window) Leaving() bool { return w.quitting }

// Look reads the work again and draws from what came back.
func (w Window) Look() Window {
	found, err := engine.Look(w.work)
	if err != nil {
		if agreeing, ok := engine.AgreeingTo(engine.Operation{}, engine.Input{}, err); ok && w.screen != OnAsking {
			return w.agreeingTo(agreeing)
		}
		return w.failing(err)
	}
	w.snap, w.version, w.err = found, found.Version, nil
	for i := range engine.Columns {
		w.scroll[i] = w.settled(i, w.scroll[i])
	}
	for i := range w.lanes {
		w.lanes[i] = w.settledLane(i, w.lanes[i])
	}
	if w = w.freshened(); w.screen == OnForm && w.err == nil {
		w = w.readingActs()
	}
	if w.pages.Shows(engine.PageTime) {
		return w.readingWeek()
	}
	return w
}

func (w Window) agreeingTo(agreeing engine.Agreeing) Window {
	w = w.Closing()
	if agreeing.Reason == engine.RefusalOf(engine.ErrStoreBehind).Reason {
		return w.applying(agreeing)
	}
	if agreeing.Reason == engine.RefusalOf(engine.ErrNewerStore).Reason {
		w.screen, w.question, w.agreeing, w.err = OnAsking, aboutNewer, agreeing, nil
		return w
	}
	w.screen, w.question, w.agreeing = OnAsking, aboutFinishingParts, agreeing
	return w
}

func (w Window) applying(agreeing engine.Agreeing) Window {
	w.screen, w.question, w.agreeing, w.err = OnAsking, aboutApplying, agreeing, nil
	w.waiting = nil
	if listed, err := w.work.Ask(engine.MigrationList, engine.Input{}); err == nil {
		for _, m := range listed.Changes {
			if shape := migration.OfMigration(m); !shape.Applied {
				w.waiting = append(w.waiting, shape.Name)
			}
		}
	}
	return w
}

func (w Window) inColumn(column int) []engine.Card {
	dealt, _ := engine.Deal(w.Showing().Cards, w.archived)
	return dealt[column].Cards
}

func (w Window) inLane(lane int) []engine.Card {
	_, stalled := engine.Deal(w.Showing().Cards, w.archived)
	return engine.InLanes(stalled)[lane]
}

func (w Window) stateOf(number int) string {
	c, _ := w.snap.Card(number)
	return c.State
}

func (w Window) settled(column, want int) int {
	l := w.Layout()
	most := max(0, len(w.inColumn(column))*CardHeight-l.Column[column].H)
	return min(most, max(0, want))
}

func (w Window) settledLane(lane, want int) int {
	most := max(0, len(w.inLane(lane))*CardHeight-w.Layout().Lane[lane].H)
	return min(most, max(0, want))
}

func line(act trace.Act) string { return engine.SaidOf(act).Line() }

func abs(n int) int { return max(n, -n) }
