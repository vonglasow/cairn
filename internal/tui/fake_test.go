package tui

import (
	"errors"
	"slices"
	"strconv"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"

	"cairn/internal/domain"
	"cairn/internal/engine"
	"cairn/internal/reading"
	"cairn/internal/task"
	"cairn/internal/trace"
	"cairn/test"
)

// A stand-in for whatever holds the work. The board asks it questions and tells
// it to change things; a test answers, and writes down what it was told. No
// database, so a hundred of these run in the time one real store takes to open.

type fake struct {
	cards []reading.Card
	goals []reading.Goal
	focus []int
	acts  []trace.Act

	version int64
	// fail is a store that cannot be read at all; failOp is one that reads
	// perfectly well and refuses what it was asked to change. They are two
	// different faults, and a board that muddled them would say the wrong one.
	fail   error
	failOp error
	// refuseVerb refuses one verb and nothing else, which is how the later of
	// several writes in one gesture is reached: the ones before it have to go
	// through for it to be asked for at all.
	refuseVerb string

	// asked is every operation the board asked for, in order, with what it
	// gave. The fields below are the same thing read the way a test wants it:
	// Ask decodes an Input into them, so an argument named wrong arrives as an
	// absence and the assertion that wanted it fails.
	asked []asked

	edited      map[int]engine.Changes
	added       []string
	deleted     []int
	deletedGoal []int
	moved       []placed
	tookHold    []int
	letGo       int
	addedGoal   []string
	editedGoal  map[int]engine.ObjectiveChanges
}

// asked is one thing the board asked the engine to do.
type asked struct {
	op engine.Operation
	in engine.Input
}

func newFake() *fake {
	return &fake{edited: map[int]engine.Changes{}, editedGoal: map[int]engine.ObjectiveChanges{}}
}

var (
	errNope   = errors.New("the store is not answering")
	errRefuse = errors.New("that cannot be done")
)

// failing is what an operation answers: whatever it was told to refuse with,
// or the store being out of reach when the whole thing is.
func (f *fake) failing() error {
	if f.failOp != nil {
		return f.failOp
	}
	return f.fail
}

// The board reads through Look, which speaks in tasks and objectives. A fake
// answers with none of those and is read through the same door instead: what a
// test wants to set is what a card carries, not how a task is built.
// placed is one move, read back the way a test wants to read it: which task,
// to which of the three places, and behind what when that is the place.
type placed struct {
	number int
	at     engine.Place
	target int
}

// Version is the one thing the board asks for that is not an operation: it is
// the store noticing another connection's write, which nobody asks cairn to do.
func (f *fake) Version() (int64, error) { return f.version, f.fail }

// standing builds a board already holding a reading, so a test starts where a
// person starts rather than at an empty screen.
func standing(f *fake) Board {
	b := New(f, GlyphsFor(Unicode), func() time.Time { return today }, 100, 40)
	// A card carries whether it is in focus, the way reading.Look marks it.
	cards := make([]reading.Card, 0, len(f.cards))
	for _, c := range f.cards {
		c.InFocus = slices.Contains(f.focus, c.Number)
		cards = append(cards, c)
	}
	b.snap = test.Searchable(reading.Snapshot{Cards: cards, Goals: f.goals, Focus: f.focus, Version: f.version})
	return b
}

// hit presses a key the way a terminal would — through Update, so whatever has
// the window gets the press itself — and answers the board it leaves behind. The
// command is thrown away: what it would have done is checked by what the fake
// was told, which is asked for separately.
func hit(b Board, keys ...string) Board {
	for _, key := range keys {
		was := b.screen
		model, cmd := b.Update(keyPress(key))
		b = pump(model.(Board), cmd)
		if was != OnForm && b.screen == OnForm {
			b = settleForm(b)
		}
	}
	return b
}

// settleForm drains what a form asks for the moment it opens, the way the
// runtime would. bubbletea unpacks a sequence itself; a test cannot, because
// the message that carries one is the runtime's own. So the form is asked
// directly, which a test in this package may do and nothing else may.
//
// Only on opening. Calling it again would ask the form to start over, and a
// form that starts over puts the cursor back on its first field.
func settleForm(b Board) Board {
	if b.screen != OnForm {
		return b
	}
	cmd := b.form.asked.Init()
	for depth := 0; cmd != nil && depth < 24; depth++ {
		msg, answered := quickly(cmd)
		if !answered || msg == nil {
			return b
		}
		model, next := b.form.asked.Update(msg)
		if form, ok := model.(*huh.Form); ok {
			b.form.asked = form
		}
		cmd = next
	}
	return b
}

// pump runs what a board asked for, the way bubbletea would, and hands back
// whatever the answers left behind.
//
// A command that does not answer quickly is dropped rather than waited on. That
// is not laziness: the commands a form asks for at the start include a cursor
// blink, and a blink is a timer that sleeps for as long as a blink lasts. A
// test that ran them would spend its whole life waiting to see a cursor it
// cannot look at.
func pump(b Board, cmd tea.Cmd) Board {
	for depth := 0; cmd != nil && depth < 16; depth++ {
		msg, answered := ran(b, cmd)
		if !answered || msg == nil {
			return b
		}
		if batch, ok := msg.(tea.BatchMsg); ok {
			var next tea.Cmd
			for _, one := range batch {
				b, next = pumpOne(b, one)
			}
			cmd = next
			continue
		}
		if _, beating := msg.(Beat); beating {
			return b
		}
		model, next := b.Update(msg)
		b = model.(Board)
		cmd = next
	}
	return b
}

// ran carries out one command, and how it does so turns on who has the window.
//
// The form's own commands are raced against a deadline, because a cursor blink
// is among them and a blink is a timer nobody can wait out. Everything else is
// simply called: those are the board's own commands, they write to the store
// and read it back, and racing one is how a scenario comes to pass nine times
// in ten — measured, once in ten runs of `prek run --all-files` under the load
// of ten suites at once.
//
// The board's state is read each time round rather than once at the start,
// because the press that saves a form is a press the form has already handed
// the window back on.
func ran(b Board, cmd tea.Cmd) (tea.Msg, bool) {
	if b.screen == OnForm {
		return quickly(cmd)
	}
	msg := cmd()
	return msg, msg != nil
}

// pumpOne runs a single command and hands the board what it produced. Through
// ran, like the rest: a batch is where the board's own work and the form's
// arrive together, so the same rule has to hold inside one.
func pumpOne(b Board, cmd tea.Cmd) (Board, tea.Cmd) {
	msg, answered := ran(b, cmd)
	if !answered || msg == nil {
		return b, nil
	}
	model, next := b.Update(msg)
	return model.(Board), next
}

// quickly runs a command and answers what it produced, or says it never did.
//
// The deadline is here for one kind of command only: a timer. A cursor blink
// and the board's own two-second beat both sleep, and a suite that waited for
// them would take minutes instead of seconds. Everything else answers as fast
// as the store does.
//
// It was five milliseconds and that was wrong. Measured: one run in ten of
// `prek run --all-files` failed the scenario "pushing a task down puts it one
// place lower", because ten copies of the suite run at once there and a write
// plus a re-read against real SQLite crossed five milliseconds — so the refresh
// was thrown away and the board answered from its old reading. Nine runs in ten
// passed, which is exactly what a deadline set at the cost of the work looks
// like.
//
// Raising it was the wrong answer, and it is written down so nobody tries
// again: a quarter of a second made the suite twelve times slower, because it
// then waited out every blink it exists to drop. Deferring the ones that did
// not answer cost eleven seconds for the same reason.
//
// What was actually wrong is that commands with nothing to do with the form
// came through here at all. See ran, above.
const patience = 5 * time.Millisecond

func quickly(cmd tea.Cmd) (tea.Msg, bool) {
	if cmd == nil {
		return nil, false
	}
	answered := make(chan tea.Msg, 1)
	go func() { answered <- cmd() }()
	select {
	case msg := <-answered:
		return msg, true
	case <-time.After(patience):
		return nil, false
	}
}

// finish walks a form to its end and answers it, keeping or throwing away what
// was typed. Tab moves on from every kind of field, including the box the
// description is written in, where enter would only add a line.
func finish(b Board, keep bool) Board {
	for range len(b.form.fields) {
		b = hit(b, "tab")
	}
	if !keep {
		b = hit(b, "left")
	}
	// Twice: the first answers the last question, and the form asks for one
	// more thing before it says it is done.
	return hit(hit(b, "enter"), "enter")
}

// anAct builds one act, through the trace's own membrane, for the tests that
// need a history to read.
func anAct(did, became string) trace.Act {
	var told trace.AsTold
	told.On(1)
	told.Did(did)
	told.NothingNamed()
	told.WasNothing()
	told.Became(became)
	told.At(today)
	return told.Act()
}

// with hands back the same board dressed differently, so a test can read what
// a card says instead of what it looks like.
func (b Board) with(skin Skin) Board {
	b.skin = skin
	return b
}

// Ask is the one door the board goes through. It writes down the operation and
// what came with it, then decodes the input into the fields the assertions read.
func (f *fake) Ask(op engine.Operation, in engine.Input) (engine.Answer, error) {
	f.asked = append(f.asked, asked{op: op, in: in})
	// A store that cannot be read at all refuses everything; a store that
	// refuses what it was asked to change still answers a listing. Which is
	// the whole point of keeping the two apart: a gesture that fails and then
	// reads the board back has to leave the failure visible, and it could not
	// if the read failed too.
	refusal := f.fail
	if op.Level != engine.Read {
		refusal = f.failing()
	}
	if f.refuseVerb != "" {
		refusal = nil
		if op.Verb == f.refuseVerb {
			refusal = errRefuse
		}
	}
	if refusal != nil {
		return engine.Answer{}, refusal
	}
	number, _ := strconv.Atoi(in.Given["number"])
	switch op.Thing + " " + op.Verb {
	case "task edit":
		f.edited[number] = changesAsked(in)
	case "task move":
		f.moved = append(f.moved, placedBy(number, in))
	case "task focus":
		f.tookHold = append(f.tookHold, number)
	case "task unfocus":
		f.letGo++
	case "task delete":
		f.deleted = append(f.deleted, number)
	case "objective delete":
		f.deletedGoal = append(f.deletedGoal, number)
	case "task new":
		f.added = append(f.added, in.Given["title"])
	case "objective new":
		f.addedGoal = append(f.addedGoal, in.Given["title"])
	case "objective edit":
		f.editedGoal[number] = goalChangesAsked(in)
	case "task list":
		// A listing marks the focus, so this is where the board learns it.
		return engine.Answer{InFocus: &f.focus}, nil
	case "trace show":
		return engine.Answer{Acts: f.acts}, nil
	}
	return answerShaped(op), nil
}

// answerShaped is an answer of the shape the operation declares, holding
// nothing: what a command typed at the prompt answers is written out, and
// the engine writes only what it was handed.
func answerShaped(op engine.Operation) engine.Answer {
	switch op.Answers.(type) {
	case engine.Task, engine.Focus:
		return engine.Answer{Task: &task.Task{}}
	case engine.BackupShape:
		return engine.Answer{Backup: &domain.Backup{Number: 1}, Holds: &domain.Held{Tasks: 1}}
	}
	return engine.Answer{}
}

// placedBy reads back which of `task move`'s three places was named.
func placedBy(number int, in engine.Input) placed {
	switch {
	case in.Named("first"):
		return placed{number: number, at: engine.First}
	case in.Named("last"):
		return placed{number: number, at: engine.Last}
	default:
		target, _ := strconv.Atoi(in.Given["after"])
		return placed{number: number, at: engine.Behind, target: target}
	}
}

// goalChangesAsked is changesAsked for an objective, which carries fewer
// things and none of them about what is in the way.
func goalChangesAsked(in engine.Input) engine.ObjectiveChanges {
	var c engine.ObjectiveChanges
	if v, ok := in.Given["state"]; ok {
		c.State = &v
	}
	return c
}

// changesAsked reads an Input the way the engine's own operation reads it, so
// what a test asserts is what the store would have been told. An argument the
// board named wrong lands in none of these, and the assertion that wanted it
// sees an absence.
func changesAsked(in engine.Input) engine.Changes {
	var c engine.Changes
	if v, ok := in.Given["state"]; ok {
		c.State = &v
	}
	return c
}

// heardOf is the board once the history of a task has been read, the way
// opening the task reads it.
func heardOf(b Board, number int) Board {
	model, _ := b.Update(b.hearing(number)())
	return model.(Board)
}
