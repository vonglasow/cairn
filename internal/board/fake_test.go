package board

import (
	"errors"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"

	"cairn/internal/domain"
	"cairn/internal/engine"
	"cairn/internal/objective"
	"cairn/internal/task"
	"cairn/internal/trace"
)

// A stand-in for whatever holds the work. The board asks it questions and tells
// it to change things; a test answers, and writes down what it was told. No
// database, so a hundred of these run in the time one real store takes to open.

type fake struct {
	cards []Card
	goals []Goal
	focus int
	acts  []trace.Act

	version int64
	// fail is a store that cannot be read at all; failOp is one that reads
	// perfectly well and refuses what it was asked to change. They are two
	// different faults, and a board that muddled them would say the wrong one.
	fail   error
	failOp error

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
func (f *fake) List() ([]task.Task, error) { return nil, f.fail }

func (f *fake) ListObjectives() ([]objective.Objective, error) { return nil, f.fail }

func (f *fake) Focus() (int, error) { return f.focus, f.fail }

func (f *fake) TaskProgress(int) (domain.Progress, error) { return domain.Progress{}, nil }

func (f *fake) ObjectiveProgress(int) (domain.Progress, error) { return domain.Progress{}, nil }

func (f *fake) History(int) ([]trace.Act, error) { return f.acts, f.fail }

func (f *fake) Add(title string, c engine.Changes) (task.Task, error) {
	f.added = append(f.added, title)
	f.edited[0] = c
	return task.Task{}, f.failing()
}

func (f *fake) Edit(number int, c engine.Changes) (task.Task, error) {
	f.edited[number] = c
	return task.Task{}, f.failing()
}

func (f *fake) AddObjective(title string, c engine.ObjectiveChanges) (objective.Objective, error) {
	f.addedGoal = append(f.addedGoal, title)
	f.editedGoal[0] = c
	return objective.Objective{}, f.failing()
}

func (f *fake) EditObjective(number int, c engine.ObjectiveChanges) (objective.Objective, error) {
	f.editedGoal[number] = c
	return objective.Objective{}, f.failing()
}

func (f *fake) Delete(number int) error {
	f.deleted = append(f.deleted, number)
	return f.failing()
}

func (f *fake) DeleteObjective(number int) error {
	f.deletedGoal = append(f.deletedGoal, number)
	return f.failing()
}

// placed is one change of place the board asked for.
type placed struct {
	number int
	at     engine.Place
	target int
}

func (f *fake) Move(number int, at engine.Place, target int) error {
	f.moved = append(f.moved, placed{number, at, target})
	return f.failing()
}

func (f *fake) PutInFocus(number int) (task.Task, error) {
	f.tookHold = append(f.tookHold, number)
	return task.Task{}, f.failing()
}

func (f *fake) LetGo() error {
	f.letGo++
	return f.failing()
}

func (f *fake) Version() (int64, error) { return f.version, f.fail }

// standing builds a board already holding a reading, so a test starts where a
// person starts rather than at an empty screen.
func standing(f *fake) Board {
	b := New(f, GlyphsFor(Unicode), func() time.Time { return today }, 100, 40)
	b.snap = Snapshot{Cards: f.cards, Goals: f.goals, Focus: f.focus, Version: f.version}
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
		msg, answered := quickly(cmd)
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

// pumpOne runs a single command and hands the board what it produced.
func pumpOne(b Board, cmd tea.Cmd) (Board, tea.Cmd) {
	msg, answered := quickly(cmd)
	if !answered || msg == nil {
		return b, nil
	}
	model, next := b.Update(msg)
	return model.(Board), next
}

// quickly runs a command and answers what it produced, or says it never did.
func quickly(cmd tea.Cmd) (tea.Msg, bool) {
	if cmd == nil {
		return nil, false
	}
	answered := make(chan tea.Msg, 1)
	go func() { answered <- cmd() }()
	select {
	case msg := <-answered:
		return msg, true
	case <-time.After(5 * time.Millisecond):
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
