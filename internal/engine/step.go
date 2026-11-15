package engine

import (
	"fmt"
	"slices"

	"cairn/internal/engine/task"
)

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
	at, _ := in.whole("step")
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
	Level:   write,
	Needs:   aStore,
	Again:   doesItAgain,
	Answers: task.Steps{},
	Takes: []Argument{
		stepOfTask,
		{Name: "title", Kind: Text, Positional: true, Required: true, Means: "what the step is called"},
		{Name: "done", Kind: Bool, Means: "add it already done"},
	},
	Run: func(on workedOn, in Input) (Answer, error) {
		n, _ := in.whole("number")
		step, err := task.NewStep(in.Given["title"], in.yes("done"))
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
	Level:   write,
	Needs:   aStore,
	Again:   ChangesNothingMore,
	Answers: task.Steps{},
	Takes: []Argument{
		stepOfTask,
		stepAtPlace,
		{Name: "title", Kind: Text, Means: "what the step is called now"},
		{Name: "done", Kind: Bool, Means: "tick it"},
		{Name: "not-done", Kind: Bool, Means: "untick it"},
	},
	Run: func(on workedOn, in Input) (Answer, error) {
		if in.yes("done") && in.yes("not-done") {
			return Answer{}, task.ErrDoneAndNotDone
		}
		if !in.Named("title") && !in.Named("done") && !in.Named("not-done") {
			return Answer{}, errNothingToChange
		}
		n, _ := in.whole("number")
		return on.Engine.changingSteps(n, func(steps []task.Step) ([]task.Step, error) {
			at, err := placed(steps, in)
			if err != nil {
				return nil, err
			}
			step := steps[at]
			if given := in.change("title"); given != nil {
				if step, err = task.NewStep(*given, step.Done); err != nil {
					return nil, err
				}
			}
			if in.yes("done") || in.yes("not-done") {
				step.Done = in.yes("done")
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
	Level:   write,
	Needs:   aStore,
	Again:   doesItAgain,
	Answers: task.Steps{},
	Takes:   []Argument{stepOfTask, stepAtPlace},
	Run: func(on workedOn, in Input) (Answer, error) {
		n, _ := in.whole("number")
		return on.Engine.changingSteps(n, func(steps []task.Step) ([]task.Step, error) {
			at, err := placed(steps, in)
			if err != nil {
				return nil, err
			}
			return slices.Delete(steps, at, at+1), nil
		})
	},
}

// stepList names the steps of a task, each with its place.
var stepList = Operation{
	Thing:   "step",
	Verb:    "list",
	Short:   "Name the steps of a task",
	Does:    "Name the steps of a task, in their order, each with its place and whether it is done.",
	Example: "cairn step list 3",
	Level:   Read,
	Needs:   aStore,
	Answers: task.Steps{},
	Takes:   []Argument{stepOfTask},
	Run: func(on workedOn, in Input) (Answer, error) {
		n, _ := in.whole("number")
		return on.Engine.stepsAnswer(n)
	},
}

func stepsText(t task.Task) string {
	steps := t.Carried().Steps
	if len(steps) == 0 {
		return fmt.Sprintf("task %d has no steps\n", t.Carried().Number)
	}
	return task.StepsText(steps)[1:]
}
