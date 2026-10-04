package reading

import (
	"errors"

	"cairn/internal/engine"
	"cairn/internal/objective"
	"cairn/internal/task"
)

// A stand-in for whatever holds the work. What is read here is how the two
// listings become cards and goals, and how a task tells its own facts — none
// of which needs a store, so this answers Work's two questions and writes down
// what it was asked.

type fake struct {
	asked   []asked
	tasks   []task.Task
	goals   []objective.Objective
	focus   []int
	freed   []int
	context string
	version int64
	// fail refuses everything; refuse refuses one operation by name, which is
	// what a test needs to reach the path that gives up halfway through a
	// reading rather than at its first question.
	fail   error
	refuse map[string]error
	// versionFails is its own field because a store can answer every listing
	// and still fail to say which version it is at, and a reading has to give
	// up on that too.
	versionFails error
}

// asked is one thing a reading asked the engine to do.
type asked struct {
	op engine.Operation
	in engine.Input
}

var errNope = errors.New("the store is not answering")

func newFake() *fake { return &fake{} }

func (f *fake) Ask(op engine.Operation, in engine.Input) (engine.Answer, error) {
	f.asked = append(f.asked, asked{op: op, in: in})
	if f.fail != nil {
		return engine.Answer{}, f.fail
	}
	named := op.Thing + " " + op.Verb
	if err, refused := f.refuse[named]; refused {
		return engine.Answer{}, err
	}
	switch named {
	case "task list":
		// A listing marks the focus, which is where a reading learns it.
		return engine.Answer{Tasks: f.tasks, InFocus: &f.focus, NothingBlocks: f.freed}, nil
	case "objective list":
		return engine.Answer{Objectives: f.goals}, nil
	case "task show":
		return engine.Answer{Context: f.context}, nil
	}
	return engine.Answer{}, nil
}

func (f *fake) Version() (int64, error) {
	if f.versionFails != nil {
		return 0, f.versionFails
	}
	return f.version, f.fail
}

// aTask and aGoal build one the only way there is: through its own membrane.
func aTask(title string) task.Task {
	built, err := task.NewTask(title)
	if err != nil {
		panic(err)
	}
	return built
}

func aGoal(title string) objective.Objective {
	built, err := objective.NewObjective(title)
	if err != nil {
		panic(err)
	}
	return built
}
