package engine

// Narrowing a listing down to what was asked for: which arguments a listing
// takes, and how what a caller named becomes what the store keeps by.

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"cairn/internal/store"
	"cairn/internal/task"
)

// narrowedBy turns the attributes a thing carries into the arguments a listing
// of them takes.
//
// Read off the attributes rather than written again, so a closed set lives in
// one place: add a column to domain.States and the flag that keeps by it
// accepts it, with no second list to remember. Only what the flag means is
// written here, because setting a fact and keeping what carries it are two
// different sentences — "the column it sits in" is what an edit does, and a
// listing puts nothing anywhere.
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

// keeping turns what a caller named into what the store keeps by. A fact not
// named is not passed on at all, so a listing nobody narrowed reads the store
// whole rather than reading it through a filter that lets everything past.
//
// A value outside a closed set is refused rather than answered with nothing.
// Answering nothing would be a lie a caller cannot tell from the truth: asked
// for a column that does not exist, "there is no task" reads as an empty
// column. Held against what the argument itself declares it accepts, so the
// closed set stays in the one place it is written.
func keeping(in Input, taken []Argument) (*store.Narrowing, error) {
	for _, argument := range taken {
		if err := accepted(in, argument); err != nil {
			return nil, err
		}
	}
	kept := &store.Narrowing{}
	if state, named := in.Given["state"]; named {
		kept.InColumn(state)
	}
	if priority, named := in.Given["priority"]; named {
		kept.Carrying(priority)
	}
	// Already read as a day by Check, which refuses what would not parse
	// before Run is reached. Passed on as the text the store holds, so the
	// comparison is the one SQL makes on the column as it is written.
	if due, named := in.Given["due-before"]; named {
		kept.DueBefore(due)
	}
	// Named one at a time and kept one at a time: a task serves a set, so
	// naming two objectives keeps what serves both — the same as naming two
	// labels, and for the same reason.
	if serving := in.Changed("objective"); serving != nil {
		for _, one := range *serving {
			number, err := strconv.Atoi(strings.TrimSpace(one))
			if err != nil || number <= 0 {
				return nil, fmt.Errorf("%q: %w", one, ErrObjectiveRequired)
			}
			kept.Serving(number)
		}
	}
	if labels := in.Changed("label"); labels != nil {
		kept.Labelled(*labels)
	}
	// Naming it keeps what is off the board and nothing else; a board asks for
	// both and draws them apart.
	if in.Yes("archived") {
		kept.Archived()
	}
	if in.Yes("with-archived") {
		kept.WithArchived()
	}
	return kept, nil
}

// accepted refuses a value the argument does not admit, naming the value and
// what it accepts instead. An argument with no closed set admits anything, and
// one nobody named is not held to anything.
func accepted(in Input, argument Argument) error {
	if len(argument.OneOf) == 0 {
		return nil
	}
	given, named := in.Given[argument.Name]
	if !named || slices.Contains(argument.OneOf, given) {
		return nil
	}
	return fmt.Errorf("%q: not one of %s", given, strings.Join(argument.OneOf, ", "))
}

// heldAmong says which task among these is the one being worked on, and none
// when it is not among them.
//
// The focus is a flag on a task, so a listing that does not carry the task has
// nothing to say about it. Said anyway, a caller would read a number it cannot
// find among the tasks it was handed — and an unnarrowed listing carries every
// task, so nothing changes there.
func heldAmong(tasks []task.Task, inFocus []int) []int {
	kept := make([]int, 0, len(inFocus))
	for _, one := range tasks {
		if number := task.OfTask(one).Number; slices.Contains(inFocus, number) {
			kept = append(kept, number)
		}
	}
	return kept
}
