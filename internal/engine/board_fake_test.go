package engine

import (
	"errors"

	"cairn/internal/domain"
	"cairn/internal/objective"
	"cairn/internal/task"
)

type fake struct {
	asked        []askedOf
	tasks        []task.Task
	goals        []objective.Objective
	focus        []int
	freed        []int
	context      string
	along        map[int]domain.Progress
	goalsAlong   map[int]domain.Progress
	version      int64
	fail         error
	refuse       map[string]error
	versionFails error
}

type askedOf struct {
	op Operation
	in Input
}

var errNope = errors.New("the store is not answering")

func newFake() *fake { return &fake{} }

func (f *fake) Ask(op Operation, in Input) (Answer, error) {
	f.asked = append(f.asked, askedOf{op: op, in: in})
	if f.fail != nil {
		return Answer{}, f.fail
	}
	named := op.Thing + " " + op.Verb
	if err, refused := f.refuse[named]; refused {
		return Answer{}, err
	}
	switch named {
	case "task list":
		contexts := map[int]string{}
		for _, t := range f.tasks {
			contexts[task.OfTask(t).Number] = f.context
		}
		return Answer{Tasks: f.tasks, InFocus: &f.focus, NothingBlocks: f.freed,
			Contexts: contexts, Along: f.along}, nil
	case "objective list":
		return Answer{Objectives: f.goals, Along: f.goalsAlong}, nil
	}
	return Answer{}, nil
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
