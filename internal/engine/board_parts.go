package engine

import (
	"cairn/internal/task"
)

// Added is the number of the task an add answered with.
func Added(answer Answer) int {
	if answer.Task == nil {
		return 0
	}
	return task.OfTask(*answer.Task).Number
}

var stepsArgument = Argument{Name: "steps", Kind: Text,
	Means: `its steps, one a line: "[x] its title" for one done, "[ ] its title" for one not; they replace the steps it had`}

// StepsIn reads steps written one a line, refusing a line with no title.
func StepsIn(said string) ([]task.Step, error) {
	parts := PartsIn(said)
	steps := make([]task.Step, 0, len(parts))
	for _, part := range parts {
		step, err := task.NewStep(part.Title, part.Done)
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)
	}
	return steps, nil
}
