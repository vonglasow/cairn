package tui

import (
	rdebug "runtime/debug"
	"time"

	"cairn/internal/debug"
	"cairn/internal/engine"
	"cairn/internal/trace"

	tea "charm.land/bubbletea/v2"
)

// Tab is which of the tabs the board is showing.
type Tab int

// The tabs.
const (
	Tasks Tab = iota
	Objectives
	Time
	Series
)

// Mode is what the cursor is holding.
type Mode int

// The two modes of the board.
const (
	Normal Mode = iota
	Insert
)

// Screen is what is filling the window.
type Screen int

// The screens.
const (
	OnTheBoard Screen = iota
	OnHelp
	OnDetail
	OnForm
	// OnSaid is what a command answered, until esc.
	OnSaid
	// OnWorkspaces is the switcher, over the board, until a workspace is chosen or esc.
	OnWorkspaces
)

// Prompt is what, if anything, is taking what is typed at the foot.
type Prompt int

// The prompts.
const (
	NoPrompt Prompt = iota
	Searching
	Commanding
)

// Looked carries a fresh reading of the work.
type Looked struct {
	Snapshot engine.Snapshot
	Err      error
	Warned   []string
}

// Heard is the history of the task that is open, read when it opens and again whenever the board is read again — never while it is being drawn.
type Heard struct {
	Number int
	Acts   []trace.Act
	Err    error
}

// Beat is the tick that asks whether anything changed.
type Beat time.Time

// Board is the whole screen.
type Board struct {
	work   engine.Work
	glyphs Glyphs
	skin   Skin
	clock  func() time.Time

	lookedAt time.Time

	width, height int
	snap          engine.Snapshot

	tab    Tab
	pages  engine.Pages
	mode   Mode
	screen Screen
	prompt Prompt

	column   int
	inBand   bool
	cameFrom int

	task int
	goal int

	inside  int
	segment int
	folded  []int

	typed    string
	query    string
	archived bool

	filters   engine.Filters
	filterOn  int
	filtering bool

	viewing      string
	choosingView bool

	choosing         bool
	asking           bool
	setting          string
	leaving          bool
	sweeping         bool
	agreeing         *engine.Agreeing
	going            *Going
	stepOn           int
	saying, sayingOf int
	choice           int

	form Form

	heard Heard

	pending string

	wasScreen Screen

	offsets [engine.Columns]int
	bandAt  int
	goalAt  int
	helpAt  int

	week   time.Time
	sheet  engine.Timesheet
	said   string
	saidAt int
	timeAt int

	place    Placed
	switcher Switcher
	spaces   []engine.Workspace
	recent   []engine.RecentFolder
	sought   string
	spaceAt  int
	opened   string
	spaceAsk spaceTyping
	spaceIn  string
	newSpace string
	keeping  engine.RecentFolder

	seriesOn      int
	inOccurrences bool
	occurrenceAt  int
	seriesGoing   *seriesGoing
	wholeSave     *Form

	warned    []string
	deleting  *engine.Deleting
	deleteRow int
	refusing  string

	err   error
	noted *debug.Log
}

// New opens a board over the work, at a given size.
func New(w engine.Work, glyphs Glyphs, clock func() time.Time, width, height int) Board {
	return Board{
		work:   w,
		glyphs: glyphs,
		skin:   SkinOf(Default),
		clock:  clock,
		width:  width,
		height: height,
	}
}

// Init asks for the first reading and starts the beat.
func (b Board) Init() tea.Cmd {
	return tea.Batch(b.look(), tick(), b.placing())
}

func (b Board) look() tea.Cmd {
	return func() tea.Msg {
		found, err := engine.Look(b.work)
		return Looked{Snapshot: found, Err: err}
	}
}

func tick() tea.Cmd {
	return tea.Tick(beat(), func(t time.Time) tea.Msg { return Beat(t) })
}

func beat() time.Duration { return 2 * time.Second }

// Update takes a message and answers the board it leaves behind; in debug mode it writes down each key, each resize, and what the board shows each time that changes.
func (b Board) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if b.noted == nil {
		return b.update(msg)
	}
	defer func() {
		if stopped := recover(); stopped != nil {
			b.noted.Panicked(stopped, rdebug.Stack())
			panic(stopped)
		}
	}()
	b.noting(msg)
	next, cmd := b.update(msg)
	if after, ok := next.(Board); ok && after.Seen() != b.Seen() {
		b.noted.Say("board", after.Seen())
	}
	return next, cmd
}

func (b Board) update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		b.width, b.height = m.Width, m.Height
		return b, nil
	case Looked:
		if agreeing, asks := engine.AgreeingTo(engine.Operation{}, engine.Input{}, m.Err); asks {
			b.asking, b.agreeing = true, &agreeing
			return b, nil
		}
		b = b.took(m)
		if b.screen == OnDetail {
			return b, b.hearing(b.task)
		}
		return b, nil
	case engine.Agreeing:
		b.asking, b.agreeing = true, &m
		return b, nil
	case Heard:
		b.heard = m
		return b, nil
	case Placed:
		b.place = m
		return b, nil
	case Listed:
		return b.listedAll(m), nil
	case Switched:
		return b.switched(m)
	case Sheeted:
		b.sheet, b.err = m.Sheet, m.Err
		return b, nil
	case Answered:
		b = b.took(m.Looked)
		if m.Err == nil {
			b = b.answering(m.Text)
		}
		return b, nil
	case Beat:
		aged := b.focusAged(time.Time(m))
		if aged {
			b.lookedAt = time.Time(m)
		}
		return b, tea.Batch(b.wonder(aged), tick())
	case tea.KeyPressMsg:
		key := keyName(m)
		b.err, b.opened, b.warned = nil, "", nil
		if b.asking {
			return b.pressed(key)
		}
		if b.screen == OnForm {
			if key == "?" {
				return b.pressed(key)
			}
			if key == "ctrl+s" {
				return b.save()
			}
			if key == "esc" {
				if b.form.Changed() {
					b.asking, b.leaving = true, true
					b.screen = OnTheBoard
					return b, nil
				}
				b.screen = OnTheBoard
				return b, nil
			}
			return b.inForm(m)
		}
		return b.pressed(key)
	}
	if b.screen == OnForm {
		return b.inForm(msg)
	}
	return b, nil
}

// FocusShown is how long the time a task in focus says is left standing: a minute, the least it counts in.
const FocusShown = time.Minute

func (b Board) focusAged(now time.Time) bool {
	return len(b.snap.Focus) > 0 && now.Sub(b.lookedAt) >= FocusShown
}

func (b Board) wonder(aged bool) tea.Cmd {
	return func() tea.Msg {
		version, err := b.work.Version()
		if err != nil {
			return Looked{Err: err}
		}
		if version == b.snap.Version && !aged {
			return nil
		}
		found, err := engine.Look(b.work)
		return Looked{Snapshot: found, Err: err}
	}
}

func (b Board) took(l Looked) Board {
	if l.Err != nil {
		b.err = l.Err
		if b.form.asked != nil {
			b.screen = OnForm
		}
		return b
	}
	if b.aboutStore() {
		b.asking, b.agreeing = false, nil
	}
	if b.form.asked != nil && !b.formStays() {
		b.screen, b.form = OnTheBoard, Form{}
	}
	was := b.task
	stood := b.at(b.here())
	b.snap = b.recurring(l.Snapshot)
	b.err = nil
	b.warned = l.Warned
	if b.form.asked != nil {
		b = b.freshened()
	}
	if was != 0 && !b.holds(was) {
		b.task = b.standingAt(stood)
		if b.task == 0 {
			b.mode = Normal
		}
	}
	return b
}

func (b Board) holds(number int) bool {
	for _, c := range b.cards() {
		if c.Number == number {
			return true
		}
	}
	return false
}

func (b Board) standingAt(place int) int {
	here := b.here()
	if len(here) == 0 || place < 0 {
		return 0
	}
	return here[min(place, len(here)-1)].Number
}

func (b Board) cards() []engine.Card {
	found := b.unopened()
	if b.inside == 0 {
		return found
	}
	reach := b.snap.Reach(b.inside)
	if segments := b.segments(); b.segment < len(segments) {
		reach = segments[b.segment].Reach
	}
	return engine.Serving(found, reach...)
}

func (b Board) unopened() []engine.Card {
	cards := b.snap.Viewed(b.viewing, b.snap.Cards)
	if kept, err := b.filters.Kept(cards); err == nil {
		cards = kept
	}
	return engine.Found(cards, b.query, b.archived)
}

func (b Board) references() []engine.Reference {
	if b.inside == 0 && b.viewing == "" && !b.filters.NarrowsBy("objective") {
		return nil
	}
	return engine.Touching(b.cards(), engine.OnTheBoard(b.snap.Cards, b.archived))
}

func (b Board) here() []engine.Card {
	all, band := engine.Deal(b.cards(), b.archived)
	if b.inBand {
		return band
	}
	return all[b.column].Cards
}

// Cursor is where the board's cursor stands, and what it is holding.
type Cursor struct {
	Mode    Mode
	Column  int
	InBand  bool
	Task    int
	Goal    int
	Showing int
}

// Asking says whether a deletion is waiting to be agreed to.
func (b Board) Asking() bool { return b.asking }

// Where answers it.
func (b Board) Where() Cursor {
	return Cursor{
		Mode:    b.mode,
		Column:  b.column,
		InBand:  b.inBand,
		Task:    b.task,
		Goal:    b.goal,
		Showing: len(b.cards()),
	}
}
