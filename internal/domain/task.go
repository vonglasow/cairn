// Package domain holds the entities and what makes them valid. It knows
// nobody: an entity that knew how it is displayed would drag a terminal into
// every test that touches it, and one that knew how it is stored would drag a
// database in.
package domain

import (
	"errors"
	"strings"
	"time"
)

// ErrTitleRequired is returned when a title carries no text at all. It is a
// sentinel so a surface can recognise it with errors.Is rather than by
// matching a message that will be reworded one day.
var ErrTitleRequired = errors.New("a task needs a title")

// NoParent is the number a task carries when it hangs under nothing. Zero is
// safe to mean "none" because the store never hands out zero: numbering starts
// at one.
const NoParent = 0

// Task is the unit of work. It carries a title, and a number to refer to it by
// afterwards. Everything else is optional and has a stated default: a task
// starts in the backlog, at normal priority, carrying nothing else.
type Task struct {
	Number      int
	Parent      int
	Title       string
	Description string
	State       State
	Priority    Priority
	Labels      []string
	Deadline    time.Time
	Estimate    time.Duration
	TimeSpent   time.Duration
}

// NewTask builds a task from a title, refusing one that carries no text.
// What surrounds the title is not part of it; what sits inside it is.
func NewTask(title string) (Task, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return Task{}, ErrTitleRequired
	}
	return Task{Title: title, State: Backlog, Priority: Normal}, nil
}
