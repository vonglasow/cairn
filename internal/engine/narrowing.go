package engine

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"cairn/internal/domain"
	"cairn/internal/task"
)

func narrowedBy(attributes []Argument, means map[string]string) []Argument {
	taken := make([]Argument, 0, len(means))
	for _, attribute := range attributes {
		said, wanted := means[attribute.Name]
		if !wanted {
			continue
		}
		attribute.Means = said
		attribute.Required = false
		attribute.Positional = false
		taken = append(taken, attribute)
	}
	return taken
}

func keeping(in Input, taken []Argument, e *Engine) (*domain.Narrowing, error) {
	for _, argument := range taken {
		if err := accepted(in, argument); err != nil {
			return nil, err
		}
	}
	kept := &domain.Narrowing{}
	if state, named := in.Given["state"]; named {
		kept.InColumn(state)
	}
	if priority, named := in.Given["priority"]; named {
		kept.Carrying(priority)
	}
	if due, named := in.Given["deadline-before"]; named {
		kept.DueBefore(due)
	}
	if serving := in.Changed("objective"); serving != nil {
		for _, one := range *serving {
			number, err := strconv.Atoi(strings.TrimSpace(one))
			if err != nil || number <= 0 {
				return nil, fmt.Errorf("%q: %w", one, ErrObjectiveRequired)
			}
			reached, err := reaching(e, number, in.Yes("itself"))
			if err != nil {
				return nil, err
			}
			kept.Serving(reached)
		}
	}
	if labels := in.Changed("label"); labels != nil {
		written, err := domain.ParseLabels(*labels)
		if err != nil {
			return nil, err
		}
		kept.Labelled(written)
	}
	if in.Yes("focus") && in.Yes("no-focus") {
		return nil, Refusing("focus-and-not", "--focus and --no-focus keep nothing together: name one")
	}
	if in.Yes("focus") || in.Yes("no-focus") {
		kept.InFocus(in.Yes("focus"))
	}
	if in.Yes("archived") {
		kept.Archived()
	}
	if in.Yes("with-archived") {
		kept.WithArchived()
	}
	return kept, nil
}

func accepted(in Input, argument Argument) error {
	if len(argument.OneOf) == 0 {
		return nil
	}
	values := slices.Clone(in.Listed[argument.Name])
	if given, named := in.Given[argument.Name]; named {
		values = append(values, given)
	}
	for _, given := range values {
		if slices.Contains(argument.OneOf, given) {
			continue
		}
		if argument.Outside != nil {
			return fmt.Errorf("%q: %w — %s", given, argument.Outside, strings.Join(argument.OneOf, ", "))
		}
		return fmt.Errorf("%q: %w %s", given, ErrNotOneOf, strings.Join(argument.OneOf, ", "))
	}
	return nil
}

// ErrNotOneOf is returned when a value falls outside the closed set its argument accepts.
var ErrNotOneOf = errors.New("not one of")

func heldAmong(tasks []task.Task, inFocus []int) []int {
	kept := make([]int, 0, len(inFocus))
	for _, number := range inFocus {
		if slices.ContainsFunc(tasks, func(one task.Task) bool { return factsOf(one).Number == number }) {
			kept = append(kept, number)
		}
	}
	return kept
}

func reaching(e *Engine, number int, itself bool) ([]int, error) {
	if e == nil || itself {
		return []int{number}, nil
	}
	t, err := e.tree()
	if err != nil {
		return nil, err
	}
	return t.reach(number), nil
}
