package tui

import (
	"cairn/internal/reading"
	"time"

	tea "charm.land/bubbletea/v2"
)

// The board itself: what is on screen, what the cursor is holding, and what a
// key does about it. Everything here is a value — a key press turns one board
// into another, and drawing reads a board without changing it. That is what
// lets a scenario press a key and read the screen without a terminal.

// Tab is which of the two the board is showing.
type Tab int

// The two tabs.
const (
	Tasks Tab = iota
	Objectives
)

// Mode is what the cursor is holding.
type Mode int

// The two modes of the board. A third, the form, is a screen rather than a
// mode: it holds the whole window, so there is nothing underneath for a mode
// to be about.
const (
	Normal Mode = iota
	Insert
)

// Screen is what is filling the window.
type Screen int

// The screens. Board is the ordinary one; the others take the whole window and
// hand it back when they are closed.
const (
	OnTheBoard Screen = iota
	OnHelp
	OnDetail
	OnForm
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
	Snapshot reading.Snapshot
	Err      error
}

// Beat is the tick that asks whether anything changed.
type Beat time.Time

// Board is the whole screen.
type Board struct {
	work   reading.Work
	glyphs Glyphs
	skin   Skin
	clock  func() time.Time

	width, height int
	snap          reading.Snapshot

	tab    Tab
	mode   Mode
	screen Screen
	prompt Prompt

	// column is where the cursor stands, left to right. inBand says it has
	// gone down into the stalled band instead, and cameFrom remembers which
	// column it came down from — on the cursor, so it dies with the session.
	column   int
	inBand   bool
	cameFrom int

	// task and goal are numbers rather than places in a list. A card can move
	// under the cursor while somebody else writes, and a cursor holding a place
	// would then be holding a different task without anything saying so.
	task int
	goal int

	// inside is the objective whose own board is open, or nothing.
	inside int

	typed string
	query string
	// archived says what was taken off the board is drawn with the rest,
	// dimmed. Nothing is asked of the store either way.
	archived bool

	choosing bool
	// asking says a deletion has been proposed and not yet agreed to. The
	// command line wants --yes and the window puts a question; this had one
	// key and no question, and nothing anywhere can undo a deletion.
	asking bool
	// leaving says the question is about a form somebody typed into rather
	// than about a deletion. Both lose work; only one of them is a deletion.
	leaving bool
	// sweeping says the question is about taking a whole column off the board.
	sweeping bool
	choice   int

	// form is what is being added or edited, when the form has the window.
	form Form

	// pending is a key half typed: `g` waiting to see whether a second one
	// follows. Anything else clears it, the way vim clears a half-typed
	// command rather than remembering it forever.
	pending string

	// wasScreen is what had the window before the help took it, so closing the
	// help puts back what was there rather than the board — help asked from
	// inside a form has to give the form back, or it would be a way of losing
	// work by asking a question.
	wasScreen Screen

	// How far each column, the band and the objectives are scrolled, in rows.
	// Kept here rather than inside a viewport so that a board stays a value: a
	// viewport is built at drawing time, from these, and thrown away.
	offsets [reading.Columns]int
	bandAt  int
	goalAt  int
	helpAt  int

	err error
}

// New opens a board over the work, at a given size.
func New(w reading.Work, glyphs Glyphs, clock func() time.Time, width, height int) Board {
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

// look reads the work. It is a command rather than a call so that the reading
// happens off the update loop, the way bubbletea expects anything that can be
// slow.
func (b Board) look() tea.Cmd {
	return func() tea.Msg {
		found, err := reading.Look(b.work)
		return Looked{Snapshot: found, Err: err}
	}
}

// tick asks again in a moment. How long a moment is is worked out here rather
// than kept as a constant, so the gate can judge it: a package-level constant
// sits outside every covered block.
//
// It reads one integer rather than the store, so two seconds costs nothing.
func tick() tea.Cmd {
	return tea.Tick(beat(), func(t time.Time) tea.Msg { return Beat(t) })
}

// beat is how long the board waits before asking again.
func beat() time.Duration { return 2 * time.Second }

// Update takes a message and answers the board it leaves behind.
func (b Board) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		b.width, b.height = m.Width, m.Height
		return b, nil
	case Looked:
		return b.took(m), nil
	case Beat:
		return b, tea.Batch(b.wonder(), tick())
	case tea.KeyPressMsg:
		key := keyName(m)
		// The form takes the press itself. Help still answers over it, because
		// help answers everywhere — including from inside a form somebody has
		// got stuck in.
		// A question waiting to be agreed to takes the next key, whatever is
		// open underneath it — the form is still on screen behind it.
		if b.asking {
			return b.pressed(key)
		}
		if b.screen == OnForm {
			// Two keys the form never sees: help, which answers everywhere,
			// and escape, which closes it. Huh binds neither — measured: the
			// bar said escape cancelled and escape did nothing at all.
			if key == "?" {
				return b.pressed(key)
			}
			if key == "esc" {
				// A form somebody typed into is asked about before it goes.
				// The board guards a deletion of two words and threw ten
				// minutes of typing away on one key.
				if b.form.Changed() {
					// The question is put at the foot, where every other
					// question is put, so the board is given back while it
					// waits. The form is still held, and backing out gives it
					// back with everything in it.
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
	// Anything else, while the form has the window, is the form's. It asks for
	// things of its own — a field to focus, a cursor to blink, its own idea of
	// what has changed — and a board that only passed on key presses would
	// leave it drawn but inert. Measured: without this the first field takes
	// what is typed and nothing ever moves off it.
	if b.screen == OnForm {
		return b.inForm(msg)
	}
	return b, nil
}

// wonder asks whether the store moved, and looks again only if it did.
func (b Board) wonder() tea.Cmd {
	return func() tea.Msg {
		version, err := b.work.Version()
		if err != nil {
			return Looked{Err: err}
		}
		if version == b.snap.Version {
			return nil
		}
		found, err := reading.Look(b.work)
		return Looked{Snapshot: found, Err: err}
	}
}

// took settles a fresh reading in, and puts the cursor back on the task it was
// holding — or on the next one along, when that task is gone.
func (b Board) took(l Looked) Board {
	if l.Err != nil {
		b.err = l.Err
		return b
	}
	// A write that went through is what closes a form, so a refused one stays
	// open with everything still in it.
	if b.screen == OnForm {
		b.screen, b.form = OnTheBoard, Form{}
	}
	was := b.task
	// Where the task stood has to be worked out before the new reading goes in,
	// because afterwards it is not there to be found. The cursor falls to
	// whatever now stands in its place, which is the next one along — and to the
	// one before when it was the last.
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

// holds says whether a task is still there to be held.
func (b Board) holds(number int) bool {
	for _, c := range b.cards() {
		if c.Number == number {
			return true
		}
	}
	return false
}

// standingAt answers the task now in a given place in the same column: the one
// that moved up into it, or the one before it when the place was the last.
func (b Board) standingAt(place int) int {
	here := b.here()
	if len(here) == 0 || place < 0 {
		return 0
	}
	return here[min(place, len(here)-1)].Number
}

// cards is every task the board is drawing, narrowed by whatever search is in
// force and by whichever objective is open.
func (b Board) cards() []reading.Card {
	var kept []reading.Card
	// What a board shows of a search, so that what is counted and what the
	// arrows walk is the same set as what is drawn — a part that matched
	// answering as the work it is part of.
	for _, c := range reading.Found(b.snap.Cards, b.query, b.archived) {
		if b.inside != 0 && !serves(c, b.inside) {
			continue
		}
		kept = append(kept, c)
	}
	return kept
}

// serves says whether a task counts towards an objective in its own right. A
// sub-task never does: it counts through the task it hangs under.
func serves(c reading.Card, number int) bool {
	for _, n := range c.Objectives {
		if n == number {
			return true
		}
	}
	return false
}

// here is the cards the cursor is standing over: a column's, or the band's.
func (b Board) here() []reading.Card {
	all, band := reading.Deal(b.cards(), b.archived)
	if b.inBand {
		return band
	}
	// No guard on the column: everything that moves it clamps it, so it is
	// always one of the four.
	return all[b.column].Cards
}

// Cursor is where the board's cursor stands, and what it is holding. One value
// answered by one method rather than a handful of accessors: a board tells what
// it is doing, the way everything else here tells what it carries.
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
