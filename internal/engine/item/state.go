package item

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

// The reasons an attribute can be refused.
var (
	ErrNotAState    = errors.New("not one of the columns")
	ErrNotAPriority = errors.New("not one of the priorities")
)

// State is the column a task sits in.
type State string

// States holds the columns in the order they are read, which is the order they are declared in below.
var States = []State{Backlog, Todo, Current, Blocked, Watching, Paused, Done}

// The columns.
const (
	Backlog  State = "backlog"
	Todo     State = "todo"
	Current  State = "current"
	Blocked  State = "blocked"
	Watching State = "watching"
	Paused   State = "paused"
	Done     State = "done"
)

// Priority is what a task carries.
type Priority string

// Priorities holds the four priorities, in the order they are declared above.
var Priorities = []Priority{Low, Normal, High, Critical}

// The four priorities.
const (
	Low      Priority = "low"
	Normal   Priority = "normal"
	High     Priority = "high"
	Critical Priority = "critical"
)

func naming[T ~string](of []T) string {
	words := make([]string, len(of))
	for i, one := range of {
		words[i] = string(one)
	}
	return strings.Join(words, ", ")
}

// ParseState reads a column name.
func ParseState(name string) (State, error) { return parseOne(name, Backlog, States, ErrNotAState) }

// ParsePriority reads a priority name.
func ParsePriority(name string) (Priority, error) {
	return parseOne(name, Normal, Priorities, ErrNotAPriority)
}

func parseOne[T ~string](name string, fallback T, set []T, outside error) (T, error) {
	if name == "" {
		return fallback, nil
	}
	if slices.Contains(set, T(name)) {
		return T(name), nil
	}
	return "", fmt.Errorf("%q: %w — %s", name, outside, naming(set))
}
