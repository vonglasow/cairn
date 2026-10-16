package tui

import (
	"time"

	"cairn/internal/engine"
	"cairn/internal/trace"

	tea "charm.land/bubbletea/v2"
)

// Tab is which of the two the board is showing.
type Tab int

// The two tabs.
const (
	Tasks Tab = iota
	Objectives
	Time
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
)

// Prompt is what, if anything, is taking what is typed at the foot.
type Prompt int

// The prompts.
const (
	NoPrompt Prompt = iota
	Searching
	Commanding
)

// Finishing is an act refused because something hanging under its task is still open, held while the person is asked whether to finish that too.
type Finishing struct {
	Op engine.Operation
	In engine.Input
}

// Looked carries a fresh reading of the work.
type Looked struct {
	Snapshot engine.Snapshot
	Err      error
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

	width, height int
	snap          engine.Snapshot

	tab    Tab
	mode   Mode
	screen Screen
	prompt Prompt

	column   int
	inBand   bool
	cameFrom int

	task int
	goal int

	inside int

	typed    string
	query    string
	archived bool

	choosing         bool
	asking           bool
	leaving          bool
	sweeping         bool
	finishing        *Finishing
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

	err error
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
	return tea.Batch(b.look(), tick())
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

// Update takes a message and answers the board it leaves behind.
func (b Board) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		b.width, b.height = m.Width, m.Height
		return b, nil
	case Looked:
		b = b.took(m)
		if b.screen == OnDetail {
			return b, b.hearing(b.task)
		}
		return b, nil
	case Finishing:
		b.asking, b.finishing = true, &m
		return b, nil
	case Heard:
		b.heard = m
		return b, nil
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
		return b, tea.Batch(b.wonder(), tick())
	case tea.KeyPressMsg:
		key := keyName(m)
		b.err = nil
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

func (b Board) wonder() tea.Cmd {
	return func() tea.Msg {
		version, err := b.work.Version()
		if err != nil {
			return Looked{Err: err}
		}
		if version == b.snap.Version {
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
	if b.form.asked != nil {
		b.screen, b.form = OnTheBoard, Form{}
	}
	was := b.task
	stood := b.at(b.here())
	b.snap = l.Snapshot
	b.err = nil
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
	found := engine.Found(b.snap.Cards, b.query, b.archived)
	if b.inside == 0 {
		return found
	}
	return engine.Serving(found, b.inside)
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
