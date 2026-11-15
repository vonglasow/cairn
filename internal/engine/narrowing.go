package engine

import (
	"cairn/internal/engine/filter"
	"cairn/internal/engine/label"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"cairn/internal/engine/task"
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

func keeping(in Input, taken []Argument, e *Engine) (*filter.Narrowing, error) {
	for _, argument := range taken {
		if err := accepted(in, argument); err != nil {
			return nil, err
		}
	}
	kept := &filter.Narrowing{}
	if state, named := in.Given["state"]; named {
		kept.InColumn(state)
	}
	if priority, named := in.Given["priority"]; named {
		kept.Carrying(priority)
	}
	if due, named := in.Given["deadline-before"]; named {
		kept.DueBefore(due)
	}
	if serving := in.changed("objective"); serving != nil {
		for _, one := range *serving {
			number, err := strconv.Atoi(strings.TrimSpace(one))
			if err != nil || number <= 0 {
				return nil, fmt.Errorf("%q: %w", one, errObjectiveRequired)
			}
			reached, err := reaching(e, number, in.yes("itself"))
			if err != nil {
				return nil, err
			}
			kept.Serving(reached)
		}
	}
	if labels := in.changed("label"); labels != nil {
		written, err := label.ParseLabels(*labels)
		if err != nil {
			return nil, err
		}
		kept.Labelled(written)
	}
	if in.yes("focus") && in.yes("no-focus") {
		return nil, Refusing("focus-and-not", "--focus and --no-focus keep nothing together: name one")
	}
	if in.yes("focus") || in.yes("no-focus") {
		kept.InFocus(in.yes("focus"))
	}
	if in.yes("archived") {
		kept.Archived()
	}
	if in.yes("with-archived") {
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
		return fmt.Errorf("%q: %w %s", given, errNotOneOf, strings.Join(argument.OneOf, ", "))
	}
	return nil
}

// errNotOneOf is returned when a value falls outside the closed set its argument accepts.
var errNotOneOf = errors.New("not one of")

func heldAmong(tasks []task.Task, inFocus []int) []int {
	kept := make([]int, 0, len(inFocus))
	for _, number := range inFocus {
		if slices.ContainsFunc(tasks, func(one task.Task) bool { return one.Carried().Number == number }) {
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
	return t.Reach(number), nil
}
