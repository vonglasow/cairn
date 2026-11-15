package board

import (
	"cairn/internal/engine"
	"cairn/internal/engine/item"
	"errors"
	"time"

	"cairn/internal/engine/objective"
	"cairn/internal/engine/task"
)

type fake struct {
	asked        []askedOf
	tasks        []task.Task
	goals        []objective.Objective
	focus        []int
	freed        []int
	chosen       map[int]task.Choosing
	context      string
	along        map[int]item.Progress
	goalsAlong   map[int]item.Progress
	version      int64
	fail         error
	refuse       map[string]error
	versionFails error
}

type askedOf struct {
	op engine.Operation
	in engine.Input
}

var errNope = errors.New("the store is not answering")

func newFake() *fake { return &fake{} }

func (f *fake) Ask(op engine.Operation, in engine.Input) (engine.Answer, error) {
	f.asked = append(f.asked, askedOf{op: op, in: in})
	if f.fail != nil {
		return engine.Answer{}, f.fail
	}
	named := op.Thing + " " + op.Verb
	if err, refused := f.refuse[named]; refused {
		return engine.Answer{}, err
	}
	switch named {
	case "task list":
		contexts := map[int]string{}
		for _, t := range f.tasks {
			contexts[task.OfTask(t).Number] = f.context
		}
		return engine.Answer{Tasks: f.tasks, InFocus: &f.focus, NothingBlocks: f.freed,
			Contexts: contexts, Along: f.along, Choosing: f.chosen}, nil
	case "objective list":
		spent := map[int]time.Duration{}
		for _, t := range f.tasks {
			for _, n := range t.Carried().Objectives {
				spent[n] += t.Carried().Spent
			}
		}
		return engine.Answer{Objectives: f.goals, Along: f.goalsAlong, Spent: spent}, nil
	}
	return engine.Answer{}, nil
}

func (f *fake) Version() (int64, error) {
	if f.versionFails != nil {
		return 0, f.versionFails
	}
	return f.version, f.fail
}

func aTask(title string) task.Task {
	built, err := task.NewTask(title)
	if err != nil {
		panic(err)
	}
	return built
}
