package engine

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"cairn/internal/task"
)

var numbersArgument = Argument{
	Name:  "number",
	Kind:  List,
	Of:    "a task",
	Means: "the number of a task to answer, and no other; named twice, both — archived or not, and a number no task carries said under unknown",
}

func askedNumbers(in Input) ([]int, error) {
	asked, named := in.Listed[numbersArgument.Name]
	if !named {
		return nil, nil
	}
	if len(asked) == 0 {
		return nil, Misread(fmt.Errorf("%q: at least a task number is expected", numbersArgument.Name))
	}
	numbers := make([]int, 0, len(asked))
	for _, raw := range asked {
		number, err := strconv.Atoi(strings.TrimSpace(raw))
		if err != nil || number <= 0 {
			return nil, Misread(fmt.Errorf("%q: a task number is expected", raw))
		}
		if !slices.Contains(numbers, number) {
			numbers = append(numbers, number)
		}
	}
	return numbers, nil
}

func (e *Engine) numbered(tasks []task.Task, numbers []int) ([]task.Task, []int, error) {
	kept := []task.Task{}
	for _, t := range tasks {
		if slices.Contains(numbers, factsOf(t).Number) {
			kept = append(kept, t)
		}
	}
	var unknown []int
	for _, number := range numbers {
		if _, err := e.show(number); errors.Is(err, ErrNoSuchTask) {
			unknown = append(unknown, number)
		} else if err != nil {
			return nil, nil, err
		}
	}
	return kept, unknown, nil
}
