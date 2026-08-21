// Package domain holds the entities and what makes them valid. It knows
// nobody: an entity that knew how it is displayed would drag a terminal into
// every test that touches it, and one that knew how it is stored would drag a
// database in.
package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"
)

// ErrTitleRequired is returned when a title carries no text at all. It is a
// sentinel so a surface can recognise it with errors.Is rather than by
// matching a message that will be reworded one day.
//
// It carries no subject of its own: what needs a title is named by whoever
// wraps it, so "a task needs a title" and "an objective needs a title" are one
// rule with two wordings rather than two sentinels a surface has to know both
// of.
var ErrTitleRequired = errors.New("needs a title")

// ErrTitleNotOneLine is returned when a title carries something that is not
// text. A line break turns one row of a listing into two, the second carrying
// no number and belonging to nothing; a tab walks through the tabwriter that
// lines a card up and shifts columns it has no business in. Trimming it away
// quietly would mean keeping a title other than the one that was typed.
var ErrTitleNotOneLine = errors.New("a title is one line, with nothing but text on it")

// breaksALine says whether a rune would end the line it is written on, or
// otherwise fail to be text. IsControl covers C0 and C1 and stops there, so
// the two Unicode separators are named: U+2028 breaks a line and U+2029 a
// paragraph, and both split a listing row in two exactly like a newline.
func breaksALine(r rune) bool {
	return unicode.IsControl(r) || r == '\u2028' || r == '\u2029'
}

// oneLine says whether a title is text and nothing else. It is shared by a
// task and an objective, which are shown the same way and break the same way.
func oneLine(title string) error {
	// ContainsFunc rather than IndexFunc >= 0: the index is not wanted, and a
	// comparison on it is a boundary no test can be written against once the
	// first character is covered.
	if strings.ContainsFunc(title, breaksALine) {
		return fmt.Errorf("%q: %w", title, ErrTitleNotOneLine)
	}
	return nil
}

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
	Objective   int
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
		return Task{}, fmt.Errorf("a task %w", ErrTitleRequired)
	}
	if err := oneLine(title); err != nil {
		return Task{}, err
	}
	return Task{Title: title, State: Backlog, Priority: Normal}, nil
}
