// Package gui draws cairn in a window a mouse can reach.
package gui

import (
	"cairn/internal/board"
	"cairn/internal/gui/design"
	"cairn/internal/gui/geom"
	"time"

	"cairn/internal/engine"
	"cairn/internal/engine/act"
	"cairn/internal/engine/migration"
	"cairn/internal/linuxapp"
)

// Held is a card being dragged: which task, where it was picked up, and where the pointer is now.
type Held struct {
	Number int
	From   geom.Point
	At     geom.Point
	Moved  bool
}

// Wobble is how far a pointer travels before a click becomes a drag.
const Wobble = 4

// Window is everything the desktop board is, at one moment.
type Window struct {
	board.Board
	work board.Work

	width, height int

	scroll                                   [board.Columns]int
	lanes                                    [3]int
	goals                                    int
	history                                  int
	saying                                   int
	spareHistory, spareSaying, spareChoosing int
	measure, measureHeads                    Measure
	keyboard                                 Keyboard
	native                                   bool

	ruling ruleState
	menuState
	filesState
	hoverKey string
	notes    notesState
	tiles    []schemeOffered
	skin     *design.Skin

	palette  paletteState
	switcher switcherState
	going    bool
	goneAt   time.Time

	savedAt                         time.Time
	born                            time.Time
	introOver                       bool
	lessMotion                      bool
	introPlayed                     string
	paused                          []string
	refused                         []string
	growing                         Unit
	anchor                          Picked
	copying                         string
	pasting                         bool
	holdingText                     bool
	serving                         int
	draggingSaying, draggingServing bool
	draggingChoosing                bool
	draggingBar                     bool

	held                  Held
	carry                 carried
	grabbedGoal           int
	hoverGoal             int
	hoverSince            time.Time
	dragOpened            []int
	pressed               bool
	pressedOn             board.Screen
	formMoved, formGrip   geom.Point
	draggingForm, dragged bool
	resizingForm          bool
	movingPanel           bool
	colourKey             string
	sliding               string
	slidTo                int
	calendarKey           string
	resizingPanel         bool
	panels                map[board.Screen]panelPlace
	resizedFrom           geom.Box
	sizes                 geom.Sizes
	kept                  SizesKept
	widening              Widening
	aimed                 string
	agreeing              engine.Agreeing
	quitting              bool
	waiting               []string
	hiding                linuxapp.Entry
	starts                string
	over                  string
	hovered               bool
	explaining            int
	lookedAt              time.Time
	resting               []string
	restingAt             geom.Box
	restingWide           int
	pointed               geom.Box
	bubbled               []string
	bubbledFrom           geom.Box
	picked                Picked
	pickingText           bool
	wheelOn               board.Screen
	wheelAt               time.Time
	drafts                Drafts
	drafting              []string
	gathering, stretching bool
	wasScreen             board.Screen

	wrongFor int
}

// Lingers is how many beats what went wrong stays: long enough to read two lines, and gone before it is mistaken for something still true.
func Lingers() int { return 4 }

// New builds a window over whatever holds the work, recurrence off until its pages say otherwise, as the configuration has it by default.
func New(work board.Work, width, height int) Window {
	w := Window{Board: board.Board{Pages: engine.Pages{}.WithoutRecurrence()}, work: work, width: width, height: height, keyboard: thisKeyboard(), drafts: noDrafts{}, kept: noSizes{}}
	return w.sized()
}

// Resized answers the window this one becomes at another size.
func (w Window) Resized(width, height int) Window {
	w.width, w.height = width, height
	return w.sized()
}

func (w Window) sized() Window {
	if geom.TooSmall(w.width, w.height) {
		if w.Screen != board.TooSmallToDraw {
			w.wasScreen = w.Screen
		}
		w.Screen = board.TooSmallToDraw
		return w
	}
	if w.Screen == board.TooSmallToDraw {
		w.Screen = w.wasScreen
	}
	return w
}

// Layout is every size this window draws from.
func (w Window) Layout() geom.Layout {
	l := geom.LaidOut(w.width, w.height, w.sizes)
	l.Days = w.Pages.WeekDays()
	if w.PlaceSays() != "" {
		l.Place = w.placeAt(l)
		l.Tabs.X, l.Tabs.W = l.Tabs.X+l.Place.W+geom.Gap, l.Tabs.W-l.Place.W-geom.Gap
	}
	return l
}

// Dragging says whether a card is in hand and has travelled far enough to be a drag rather than a click that wobbled.
func (w Window) Dragging() bool { return w.held.Number != 0 && w.held.Moved }

// ContextSays is what the foot of an open task writes about where its context lives: its name and the place, or nothing when there is no place to say.
func (w Window) ContextSays() []Readable {
	at := w.Context()
	if at == "" || w.contextPath().W == 0 {
		return nil
	}
	return []Readable{one("name:context", ContextName(w.formPanel()), "notes", small),
		one("field:context", w.contextPath(), at, small)}
}

func (w Window) contextPath() geom.Box {
	box := w.formPanel()
	end := box.X + box.W
	if offers := w.offersAlong(box); len(offers) > 0 {
		end = offers[len(offers)-1].X
	}
	path := ContextPath(box, end)
	return geom.Box{X: path.X + copyRoom, Y: path.Y, W: max(0, path.W-copyRoom), H: path.H}
}

// Wonder asks the store whether anybody else has written to it, and reads the whole thing again only when the answer has moved.
func (w Window) Wonder() Window {
	w = w.aged().autosaved()
	version, err := w.work.Version()
	if err != nil {
		return w.failing(err)
	}
	w = w.rereading()
	if version == w.Version && !w.focusAged() {
		return w
	}
	return w.Look()
}

// FocusShown is how long the time a task in focus says is left standing: a minute, the least it counts in.
const FocusShown = time.Minute

func (w Window) focusAged() bool {
	return len(w.Snap.Focus) > 0 && w.Now.Sub(w.lookedAt) >= FocusShown
}

func (w Window) failing(err error) Window {
	w.Err, w.wrongFor = err, 0
	return w
}

func lingered(beats int) (int, bool) {
	beats++
	return beats, beats >= Lingers()
}

func (w Window) aged() Window {
	if w.Telling != "" {
		var gone bool
		if w.ToldFor, gone = lingered(w.ToldFor); gone {
			w.Telling, w.ToldFor = "", 0
		}
	}
	if w.Err != nil {
		var gone bool
		if w.wrongFor, gone = lingered(w.wrongFor); gone {
			w.Err, w.wrongFor = nil, 0
		}
	}
	return w
}

// Leaving says whether the window has been asked to close, by a question whose only other way was to leave.
func (w Window) Leaving() bool { return w.quitting }

// Look reads the work again and draws from what came back.
func (w Window) Look() Window {
	found, err := board.Look(w.work)
	if err != nil {
		if agreeing, ok := engine.AgreeingTo(engine.Operation{}, engine.Input{}, err); ok && w.Screen != board.OnAsking {
			return w.agreeingTo(agreeing)
		}
		return w.failing(err)
	}
	w = w.readingPlace()
	w.Snap, w.Version, w.Err, w.lookedAt = board.Recurring(w.Pages, found), found.Version, nil, w.Now
	for i := range board.Columns {
		w.scroll[i] = w.settled(i, w.scroll[i])
	}
	for i := range w.lanes {
		w.lanes[i] = w.settledLane(i, w.lanes[i])
	}
	if w = w.freshened(); w.Screen == board.OnForm && w.Err == nil {
		w = w.readingActs()
	}
	if w.Pages.Shows(engine.PageTime) {
		return w.readingWeek()
	}
	return w
}

func (w Window) agreeingTo(agreeing engine.Agreeing) Window {
	w = w.Closing()
	if agreeing.Reason == engine.RefusalOf(migration.ErrStoreBehind).Reason {
		return w.applying(agreeing)
	}
	if agreeing.Reason == engine.RefusalOf(migration.ErrNewerStore).Reason {
		w.Screen, w.Question, w.agreeing, w.Err = board.OnAsking, board.AboutNewer, agreeing, nil
		return w
	}
	w.Screen, w.Question, w.agreeing = board.OnAsking, board.AboutFinishingParts, agreeing
	return w
}

func (w Window) applying(agreeing engine.Agreeing) Window {
	w.Screen, w.Question, w.agreeing, w.Err = board.OnAsking, board.AboutApplying, agreeing, nil
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

func (w Window) settled(column, want int) int {
	l := w.Layout()
	most := max(0, len(w.InColumn(column))*geom.CardHeight-l.Column[column].H)
	return min(most, max(0, want))
}

func (w Window) settledLane(lane, want int) int {
	most := max(0, len(w.InLane(lane))*geom.CardHeight-w.Layout().Lane[lane].H)
	return min(most, max(0, want))
}

func line(act act.Act) string { return board.SaidOf(act).Line() }

func abs(n int) int { return max(n, -n) }
