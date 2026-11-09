package domain

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

// Scope is what a label may be put on: tasks, objectives, or both.
type Scope string

// The three scopes a label can have; a label nobody scoped is for both.
const (
	ForTasks      Scope = "tasks"
	ForObjectives Scope = "objectives"
	ForBoth       Scope = "both"
)

// Scopes holds the three scopes, in the order they are offered.
var Scopes = []Scope{ForTasks, ForObjectives, ForBoth}

// Bearer is what a label is put on: a task or an objective.
type Bearer string

// The two things a label can be put on.
const (
	OnTasks      Bearer = "tasks"
	OnObjectives Bearer = "objectives"
)

// ErrNoSuchScope is returned when a scope named is none of the three.
var ErrNoSuchScope = errors.New("a label's scope is tasks, objectives or both")

// ErrOutOfScope is returned when a label is put on what its scope leaves out.
var ErrOutOfScope = errors.New("a label is put only on what its scope names")

// ParseScope reads a scope's name the one way it is written; nothing is both.
func ParseScope(given string) (Scope, error) {
	name := Scope(strings.ToLower(strings.TrimSpace(given)))
	if name == "" {
		return ForBoth, nil
	}
	if !slices.Contains(Scopes, name) {
		return "", fmt.Errorf("%q: %w", given, ErrNoSuchScope)
	}
	return name, nil
}

// Admits says whether a label of this scope may be put on what is named.
func (s Scope) Admits(on Bearer) bool { return s == ForBoth || string(s) == string(on) }

// Refuses says why label, of this scope, may not be put on what is named; nothing when it may.
func (s Scope) Refuses(label string, on Bearer) error {
	if s.Admits(on) {
		return nil
	}
	one := "a task"
	if on == OnObjectives {
		one = "an objective"
	}
	return &outOfScope{said: fmt.Sprintf("%q is a label for %s only, not for %s", label, s, one)}
}

type outOfScope struct{ said string }

func (o *outOfScope) Error() string { return o.said }

func (o *outOfScope) Unwrap() error { return ErrOutOfScope }
