package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// ErrNotAnObjectiveState is returned when something that is not one of the
// four is offered as an objective's state. It is kept apart from ErrNotAState
// on purpose: both are called a state, but a task's is a column it sits in and
// an objective's is not, so the two sets never overlap. A parser shared
// between them would accept "todo" for an objective without a word.
var ErrNotAnObjectiveState = errors.New("not one of the objective states")

// NoObjective is the number a task carries when it serves none. Zero is safe
// to mean "none" because the store never hands out zero: numbering starts at
// one, the same way NoParent works.
const NoObjective = 0

// ObjectiveState is where an objective stands. There are four, and nothing
// else is one. It is a type of its own rather than a State: they share the
// word because that is what both are called, but a task's state is a column it
// gets moved between and an objective's is not, so neither set of values is
// ever valid for the other.
type ObjectiveState string

// The four states an objective can be in. One with none stated is active.
const (
	Active  ObjectiveState = "active"
	OnHold  ObjectiveState = "on-hold"
	Reached ObjectiveState = "reached"
	Dropped ObjectiveState = "dropped"
)

// Objective is something worth doing, that tasks are put under. It carries a
// title and a number, and shares every other attribute's type with a task —
// but not its struct: the six fields they have in common say the same thing
// about two different things, and folding them together would make every
// future field of one a field of the other.
type Objective struct {
	Number      int
	Title       string
	Description string
	State       ObjectiveState
	Priority    Priority
	Labels      []string
	Deadline    time.Time
	Estimate    time.Duration
}

// NewObjective builds an objective from a title, refusing one that carries no
// text. What surrounds the title is not part of it; what sits inside it is.
func NewObjective(title string) (Objective, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return Objective{}, fmt.Errorf("an objective %w", ErrTitleRequired)
	}
	if err := oneLine(title); err != nil {
		return Objective{}, err
	}
	return Objective{Title: title, State: Active, Priority: Normal}, nil
}

// ParseObjectiveState reads the name of a state an objective can be in.
// Nothing given returns active, which is where an objective starts and what
// clearing the state returns to.
func ParseObjectiveState(name string) (ObjectiveState, error) {
	switch ObjectiveState(name) {
	case Active, OnHold, Reached, Dropped:
		return ObjectiveState(name), nil
	case "":
		return Active, nil
	}
	return "", fmt.Errorf("%q: %w", name, ErrNotAnObjectiveState)
}
