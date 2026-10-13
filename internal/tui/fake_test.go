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
	"cairn/internal/task"
	"cairn/internal/trace"
	"cairn/test"
)

type fake struct {
	cards []engine.Card
	goals []engine.Goal
	focus []int
	acts  []trace.Act

	version    int64
	fail       error
	failOp     error
	refuseVerb string

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

func (f *fake) failing() error {
	if f.failOp != nil {
		return f.failOp
	}
	return f.fail
}

type placed struct {
	number int
	at     engine.Place
	target int
}

func (f *fake) Version() (int64, error) { return f.version, f.fail }

func standing(f *fake) Board {
	b := New(f, GlyphsFor(Unicode), func() time.Time { return today }, 100, 40)
	cards := make([]engine.Card, 0, len(f.cards))
	for _, c := range f.cards {
		c.InFocus = slices.Contains(f.focus, c.Number)
		cards = append(cards, c)
	}
	b.snap = test.Searchable(engine.Snapshot{Cards: cards, Goals: f.goals, Focus: f.focus, Version: f.version})
	return b
}

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

func ran(b Board, cmd tea.Cmd) (tea.Msg, bool) {
	if b.screen == OnForm {
		return quickly(cmd)
	}
	msg := cmd()
	return msg, msg != nil
}

func pumpOne(b Board, cmd tea.Cmd) (Board, tea.Cmd) {
	msg, answered := ran(b, cmd)
	if !answered || msg == nil {
		return b, nil
	}
	model, next := b.Update(msg)
	return model.(Board), next
}

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

func finish(b Board, keep bool) Board {
	for range len(b.form.fields) {
		b = hit(b, "tab")
	}
	if !keep {
		b = hit(b, "left")
	}
	return hit(hit(b, "enter"), "enter")
}

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

func (b Board) with(skin Skin) Board {
	b.skin = skin
	return b
}

func (f *fake) Ask(op engine.Operation, in engine.Input) (engine.Answer, error) {
	f.asked = append(f.asked, asked{op: op, in: in})
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
		return engine.Answer{InFocus: &f.focus}, nil
	case "trace show":
		return engine.Answer{Acts: f.acts}, nil
	}
	return answerShaped(op), nil
}

func answerShaped(op engine.Operation) engine.Answer {
	switch op.Answers.(type) {
	case engine.Task, engine.Focus:
		return engine.Answer{Task: &task.Task{}}
	case engine.BackupShape:
		return engine.Answer{Backup: &domain.Backup{Number: 1}, Holds: &domain.Held{Tasks: 1}}
	}
	return engine.Answer{}
}

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

func goalChangesAsked(in engine.Input) engine.ObjectiveChanges {
	var c engine.ObjectiveChanges
	if v, ok := in.Given["state"]; ok {
		c.State = &v
	}
	return c
}

func changesAsked(in engine.Input) engine.Changes {
	var c engine.Changes
	if v, ok := in.Given["state"]; ok {
		c.State = &v
	}
	return c
}

func heardOf(b Board, number int) Board {
	model, _ := b.Update(b.hearing(number)())
	return model.(Board)
}
