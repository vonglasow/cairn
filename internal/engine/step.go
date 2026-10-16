package engine

import (
	"errors"
	"fmt"
	"slices"

	"cairn/internal/task"
)

// ErrDoneAndNotDone is returned when a step is asked to be ticked and unticked in one go.
var ErrDoneAndNotDone = errors.New("a step is done or not done, not both")

var (
	stepOfTask  = Argument{Name: "number", Kind: Number, Positional: true, Required: true, Of: "a task", Means: "the number of the task"}
	stepAtPlace = Argument{Name: "step", Kind: Number, Positional: true, Required: true, Means: "the place of the step in its task, the first being 1"}
)

func (e *Engine) changingSteps(number int, change func([]task.Step) ([]task.Step, error)) (Answer, error) {
	t, err := e.tasks.Find(number)
	if err != nil {
		return Answer{}, err
	}
	steps, err := change(t.Carried().Steps)
	if err != nil {
		return Answer{}, fmt.Errorf("task %d: %w", number, err)
	}
	if _, err := e.edit(number, Changes{Steps: &steps}); err != nil {
		return Answer{}, err
	}
	return e.stepsAnswer(number)
}

func (e *Engine) stepsAnswer(number int) (Answer, error) {
	t, err := e.tasks.Find(number)
	if err != nil {
		return Answer{}, err
	}
	p, err := e.taskProgress(number)
	if err != nil {
		return Answer{}, err
	}
	return Answer{Task: &t, Progress: p}, nil
}

func placed(steps []task.Step, in Input) (int, error) {
	at, _ := in.Whole("step")
	if at < 1 || at > len(steps) {
		return 0, fmt.Errorf("step %d: %w", at, task.ErrNoSuchStep)
	}
	return at - 1, nil
}

// StepNew adds a step to a task, behind the ones it has.
var StepNew = Operation{
	Thing: "step",
	Verb:  "new",
	Short: "Add a step to a task",
	Does:  "Add a step to a task, behind the steps it already has: a smaller piece of it, done or not. A step is not a task — it has no number, column, focus or time of its own. Answers the task's steps.",
	Example: `cairn step new 3 "Write the spec"
cairn step new 3 "Plan it" --done`,
	Level:   Write,
	Needs:   AStore,
	Again:   DoesItAgain,
	Answers: Steps{},
	Takes: []Argument{
		stepOfTask,
		{Name: "title", Kind: Text, Positional: true, Required: true, Means: "what the step is called"},
		{Name: "done", Kind: Bool, Means: "add it already done"},
	},
	Run: func(on On, in Input) (Answer, error) {
		n, _ := in.Whole("number")
		step, err := task.NewStep(in.Given["title"], in.Yes("done"))
		if err != nil {
			return Answer{}, err
		}
		return on.Engine.changingSteps(n, func(steps []task.Step) ([]task.Step, error) {
			return append(steps, step), nil
		})
	},
}

// StepEdit ticks, unticks or renames one step of a task.
var StepEdit = Operation{
	Thing: "step",
	Verb:  "edit",
	Short: "Tick, untick or rename a step",
	Does:  "Tick a step of a task, untick it, or rename it, the step named by its place in the task. Answers the task's steps.",
	Example: `cairn step edit 3 1 --done
cairn step edit 3 2 --title "Write the tests first"`,
	Level:   Write,
	Needs:   AStore,
	Again:   ChangesNothingMore,
	Answers: Steps{},
	Takes: []Argument{
		stepOfTask,
		stepAtPlace,
		{Name: "title", Kind: Text, Means: "what the step is called now"},
		{Name: "done", Kind: Bool, Means: "tick it"},
		{Name: "not-done", Kind: Bool, Means: "untick it"},
	},
	Run: func(on On, in Input) (Answer, error) {
		if in.Yes("done") && in.Yes("not-done") {
			return Answer{}, ErrDoneAndNotDone
		}
		if !in.Named("title") && !in.Named("done") && !in.Named("not-done") {
			return Answer{}, ErrNothingToChange
		}
		n, _ := in.Whole("number")
		return on.Engine.changingSteps(n, func(steps []task.Step) ([]task.Step, error) {
			at, err := placed(steps, in)
			if err != nil {
				return nil, err
			}
			step := steps[at]
			if given := in.Change("title"); given != nil {
				if step, err = task.NewStep(*given, step.Done); err != nil {
					return nil, err
				}
			}
			if in.Yes("done") || in.Yes("not-done") {
				step.Done = in.Yes("done")
			}
			steps[at] = step
			return steps, nil
		})
	},
}

// StepDelete takes a step away from a task.
var StepDelete = Operation{
	Thing:   "step",
	Verb:    "delete",
	Short:   "Take a step away from a task",
	Does:    "Take a step away from a task, the step named by its place in the task; the ones behind it move up a place. Answers the task's steps.",
	Example: "cairn step delete 3 2",
	Level:   Write,
	Needs:   AStore,
	Again:   DoesItAgain,
	Answers: Steps{},
	Takes:   []Argument{stepOfTask, stepAtPlace},
	Run: func(on On, in Input) (Answer, error) {
		n, _ := in.Whole("number")
		return on.Engine.changingSteps(n, func(steps []task.Step) ([]task.Step, error) {
			at, err := placed(steps, in)
			if err != nil {
				return nil, err
			}
			return slices.Delete(steps, at, at+1), nil
		})
	},
}

// StepList names the steps of a task, each with its place.
var StepList = Operation{
	Thing:   "step",
	Verb:    "list",
	Short:   "Name the steps of a task",
	Does:    "Name the steps of a task, in their order, each with its place and whether it is done.",
	Example: "cairn step list 3",
	Level:   Read,
	Needs:   AStore,
	Answers: Steps{},
	Takes:   []Argument{stepOfTask},
	Run: func(on On, in Input) (Answer, error) {
		n, _ := in.Whole("number")
		return on.Engine.stepsAnswer(n)
	},
}

// Steps is how a task's steps cross to a machine: the task they belong to, and each step with its place.
type Steps struct {
	Task  int         `json:"task"`
	Steps []StepShape `json:"steps"`
}

// StepShape is one step as it crosses, with the place it is named by.
type StepShape struct {
	Place int    `json:"place"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

func ofSteps(t task.Task) Steps {
	said := Steps{Task: factsOf(t).Number, Steps: []StepShape{}}
	for i, s := range t.Carried().Steps {
		said.Steps = append(said.Steps, StepShape{Place: i + 1, Title: s.Title, Done: s.Done})
	}
	return said
}

func stepsText(t task.Task) string {
	steps := t.Carried().Steps
	if len(steps) == 0 {
		return fmt.Sprintf("task %d has no steps\n", factsOf(t).Number)
	}
	return task.StepsText(steps)[1:]
}
