package domain

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
)

// dayLayout is how a deadline is written and read back: a day, never an
// instant. A deadline with a time of day would be a different promise.
const dayLayout = "2006-01-02"

// The reasons an attribute can be refused. Sentinels, so a surface recognises
// them with errors.Is rather than by matching a message.
var (
	ErrLabelRequired    = errors.New("a label needs a name")
	ErrNotAState        = errors.New("not one of the columns")
	ErrNotAPriority     = errors.New("not one of the priorities")
	ErrNotADay          = errors.New("not a date")
	ErrNotADuration     = errors.New("not a duration")
	ErrNegativeAmount   = errors.New("a duration cannot run backwards")
	ErrFinerThanASecond = errors.New("a duration is kept to the second, not finer")
)

// State is the column a task sits in. There are four, and nothing else is one.
type State string

// States holds the four columns, in the order they are declared above. It is
// what ParseState consults, so the two cannot say different things.
var States = []State{Backlog, Todo, Current, Done}

// The four columns. A task with no column stated starts in the backlog.
const (
	Backlog State = "backlog"
	Todo    State = "todo"
	Current State = "current"
	Done    State = "done"
)

// Priority is what a task carries. There are four, and nothing else is one.
type Priority string

// Priorities holds the four priorities, in the order they are declared above.
// It is what ParsePriority consults, so the two cannot say different things.
var Priorities = []Priority{Low, Normal, High, Critical}

// The four priorities. A task with no priority stated carries normal.
const (
	Low      Priority = "low"
	Normal   Priority = "normal"
	High     Priority = "high"
	Critical Priority = "critical"
)

// ParseState reads a column name. Nothing given returns the backlog, which is
// where a task starts and where clearing the column puts it back.
func ParseState(name string) (State, error) {
	if name == "" {
		return Backlog, nil
	}
	if slices.Contains(States, State(name)) {
		return State(name), nil
	}
	return "", fmt.Errorf("%q: %w", name, ErrNotAState)
}

// ParsePriority reads a priority name. Nothing given returns normal, which is
// what a task carries by default and what clearing it returns to.
func ParsePriority(name string) (Priority, error) {
	if name == "" {
		return Normal, nil
	}
	if slices.Contains(Priorities, Priority(name)) {
		return Priority(name), nil
	}
	return "", fmt.Errorf("%q: %w", name, ErrNotAPriority)
}

// ParseLabels turns what was given into the set a task carries: lower case, no
// duplicate, alphabetical order. An empty string is how a caller says "none";
// a string of spaces is a label somebody meant to name and did not.
func ParseLabels(given []string) ([]string, error) {
	given = listed(given)
	kept := make([]string, 0, len(given))
	for _, label := range given {
		if label == "" {
			continue
		}
		trimmed := strings.TrimSpace(label)
		if trimmed == "" {
			return nil, ErrLabelRequired
		}
		kept = append(kept, strings.ToLower(trimmed))
	}
	slices.Sort(kept)
	return slices.Compact(kept), nil
}

// ParseDeadline reads a day. Nothing given is no deadline, which is also how a
// deadline is taken away.
func ParseDeadline(day string) (time.Time, error) {
	if day == "" {
		return time.Time{}, nil
	}
	parsed, err := time.Parse(dayLayout, day)
	if err != nil {
		return time.Time{}, fmt.Errorf("%q: %w", day, ErrNotADay)
	}
	return parsed, nil
}

// ParseEstimate reads a duration. Nothing given is no estimate; a negative one
// is refused rather than stored, since no work takes less than no time.
func ParseEstimate(amount string) (time.Duration, error) {
	if amount == "" {
		return 0, nil
	}
	parsed, err := time.ParseDuration(amount)
	if err != nil {
		return 0, fmt.Errorf("%q: %w", amount, ErrNotADuration)
	}
	if parsed < 0 {
		return 0, fmt.Errorf("%q: %w", amount, ErrNegativeAmount)
	}
	// A duration is written back to the second and no finer, so anything with
	// a remainder cannot come back as it was given. Taking it and dropping the
	// remainder would be worse than refusing: 500ms was accepted, stored, and
	// read back as no estimate at all.
	if parsed%time.Second != 0 {
		return 0, fmt.Errorf("%q: %w", amount, ErrFinerThanASecond)
	}
	return parsed, nil
}

// FormatDuration writes a duration the way it was meant to be read: 90 minutes
// comes back as 1h30m, not as 1h30m0s. A duration of nothing writes nothing.
func FormatDuration(d time.Duration) string {
	if d == 0 {
		return ""
	}
	var out strings.Builder
	if h := int(d.Hours()); h > 0 {
		_, _ = fmt.Fprintf(&out, "%dh", h)
		d -= time.Duration(h) * time.Hour
	}
	if m := int(d.Minutes()); m > 0 {
		_, _ = fmt.Fprintf(&out, "%dm", m)
		d -= time.Duration(m) * time.Minute
	}
	if s := int(d.Seconds()); s > 0 {
		_, _ = fmt.Fprintf(&out, "%ds", s)
	}
	return out.String()
}

// FormatDay writes a deadline back as the day it was given. No deadline writes
// nothing.
func FormatDay(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(dayLayout)
}
