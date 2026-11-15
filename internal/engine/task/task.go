// Package task holds the task: what it carries, the rules that keep it valid, and the membrane anything uses to read or build one.
package task

import (
	"cairn/internal/engine/item"
	"errors"
	"fmt"
	"slices"
	"time"
)

// NoObjective is the number nothing serves.
const NoObjective = 0

// Task is the unit of work.
type Task struct {
	number      int
	objectives  []int
	title       string
	description string
	state       item.State
	priority    item.Priority
	labels      []string
	blockers    []int
	linked      []int
	because     string
	deadline    time.Time
	estimate    time.Duration
	timeSpent   time.Duration
	archived    bool
	steps       []Step
}

// Carried is everything a task carries, as plain facts.
type Carried struct {
	Number               int
	Title, Description   string
	State                item.State
	Priority             item.Priority
	Labels               []string
	Blockers, Objectives []int
	Because              string
	Deadline             time.Time
	Estimate, Spent      time.Duration
	Archived             bool
	Steps                []Step
	Linked               []int
}

// Carried hands over what the task carries.
func (t Task) Carried() Carried {
	return Carried{
		Number: t.number, Title: t.title, Description: t.description,
		State: t.state, Priority: t.priority, Labels: slices.Clone(t.labels),
		Blockers: slices.Clone(t.blockers), Objectives: slices.Clone(t.objectives),
		Because: t.because, Deadline: t.deadline, Estimate: t.estimate, Spent: t.timeSpent,
		Archived: t.archived, Steps: slices.Clone(t.steps),
		Linked: slices.Clone(t.linked),
	}
}

// Made is the task that carries what is given.
func Made(c Carried) Task {
	return Task{
		number: c.Number, title: c.Title, description: c.Description,
		state: c.State, priority: c.Priority, labels: c.Labels, blockers: c.Blockers,
		objectives: c.Objectives, because: c.Because, deadline: c.Deadline,
		estimate: c.Estimate, timeSpent: c.Spent, archived: c.Archived, steps: c.Steps,
		linked: c.Linked,
	}
}

// ErrAlreadyArchived is returned when a task off the board is archived again.
var ErrAlreadyArchived = errors.New("the task is already archived")

// ErrNotArchived is returned when a task on the board is brought back onto it.
var ErrNotArchived = errors.New("the task is on the board")

// MayArchive says whether this task may be taken off the board.
func (t Task) MayArchive() error {
	if t.archived {
		return fmt.Errorf("task %d: %w", t.number, ErrAlreadyArchived)
	}
	return nil
}

// MayUnarchive says whether a task can be brought back onto the board.
func (t Task) MayUnarchive() error {
	if !t.archived {
		return fmt.Errorf("task %d: %w", t.number, ErrNotArchived)
	}
	return nil
}

// NewTask builds a task from a title, refusing one that carries no text.
func NewTask(raw string) (Task, error) {
	title, err := item.Title("a task", raw)
	if err != nil {
		return Task{}, err
	}
	return Task{title: title, state: item.Backlog, priority: item.Normal}, nil
}

// Finder finds a task by its number.
type Finder interface {
	Find(int) (Task, error)
}

// ErrNothingToLabel is returned when labelling is given nothing to narrow by.
var ErrNothingToLabel = errors.New("labelling needs something to narrow by")

// ErrNothingToArchive is returned when archiving is given neither a number nor a narrowing.
var ErrNothingToArchive = errors.New("archiving needs a number or something to narrow by")

// ErrAlsoWithoutNumber is returned when archiving names more tasks beside a first, and no first: a narrowing already takes many, and "also" said beside one is a question about which.
var ErrAlsoWithoutNumber = errors.New("--also names tasks beside a number, and no number was given")

// OnPage keeps, of what each term reached, the tasks on the page.
func OnPage(terms map[string][]int, page []Task) map[string][]int {
	if terms == nil {
		return nil
	}
	shown := make(map[int]bool, len(page))
	for _, t := range page {
		shown[t.Carried().Number] = true
	}
	kept := make(map[string][]int, len(terms))
	for term, numbers := range terms {
		kept[term] = slices.DeleteFunc(slices.Clone(numbers), func(n int) bool { return !shown[n] })
	}
	return kept
}

// MatchingAny is the tasks any term reaches, and the tasks each term reached.
func MatchingAny(tasks []Task, named []string) ([]Task, map[string][]int) {
	found := []Task{}
	terms := make(map[string][]int, len(named))
	for _, term := range named {
		terms[term] = []int{}
	}
	for _, t := range tasks {
		carried := item.Among(Said(t), named)
		if len(carried) > 0 {
			found = append(found, t)
		}
		for _, term := range carried {
			terms[term] = append(terms[term], t.Carried().Number)
		}
	}
	return found, terms
}

// WithTypedTitle says what was typed beside a title refused as empty.
func WithTypedTitle(err error, typed string) error {
	if errors.Is(err, item.ErrTitleRequired) {
		return fmt.Errorf("%w, got %q", err, typed)
	}
	return err
}
