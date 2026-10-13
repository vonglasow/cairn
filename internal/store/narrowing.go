package store

import (
	"fmt"
	"strings"

	"cairn/internal/domain"
	"cairn/internal/task"
)

type narrowing struct {
	clauses  []string
	args     []any
	labels   []string
	archived bool
	both     bool
}

func readBy(kept *domain.Narrowing) *narrowing {
	n := &narrowing{}
	kept.Tell(n)
	return n
}

// Archived keeps the tasks taken off the board, instead of the ones still on it.
func (n *narrowing) Archived() { n.archived = true }

// WithArchived keeps both, which is what a board asks for: it draws the two apart rather than reading the store twice.
func (n *narrowing) WithArchived() { n.both = true }

// InColumn keeps what sits in one column.
func (n *narrowing) InColumn(state string) {
	n.keep(`state = ?`, state)
}

// DueBefore keeps what is due on or before a day, the day itself included: a deadline is the last day something may be done, so a listing of what is due by the first that left out the first would answer a different question.
func (n *narrowing) DueBefore(day string) {
	n.keep(`deadline != '' AND deadline <= ?`, day)
}

// Carrying keeps what carries one priority.
func (n *narrowing) Carrying(priority string) {
	n.keep(`priority = ?`, priority)
}

// Serving keeps what serves one objective, among however many others it also serves.
func (n *narrowing) Serving(objective int) {
	n.keep(`EXISTS (SELECT 1 FROM task_objectives
		WHERE task_objectives.task_number = tasks.number
		  AND task_objectives.objective_number = ?)`, objective)
}

// Labelled keeps what carries every label named.
func (n *narrowing) Labelled(labels []string) {
	n.labels = labels
}

func (n *narrowing) narrows() bool {
	return len(n.clauses)+len(n.labels) > 0 || n.archived
}

func (n *narrowing) keep(clause string, value any) {
	n.clauses = append(n.clauses, clause)
	n.args = append(n.args, value)
}

func (n *narrowing) where(labels labelStore) (string, []any) {
	if !n.narrows() {
		return "", nil
	}
	clauses, args := n.clauses, n.args
	for _, label := range n.labels {
		clauses = append(clauses, labels.carrier)
		args = append(args, label)
	}
	return ` WHERE ` + strings.Join(clauses, ` AND `), args
}

const offTheBoard = `(archived = 1 OR parent IN (SELECT whole.number FROM tasks AS whole WHERE whole.archived = 1))`

func (n *narrowing) whereTasks(labels labelStore) (string, []any) {
	var clauses []string
	var args []any
	if !n.both {
		clauses, args = []string{offTheBoard + ` = ?`}, []any{n.archived}
	}
	clauses = append(clauses, n.clauses...)
	args = append(args, n.args...)
	for _, label := range n.labels {
		clauses = append(clauses, labels.carrier)
		args = append(args, label)
	}
	if len(clauses) == 0 {
		return "", nil
	}
	return ` WHERE ` + strings.Join(clauses, ` AND `), args
}

// Narrowed hands back the tasks a narrowing keeps, in the one order there is.
func (s *Store) Narrowed(kept *domain.Narrowing) ([]task.Task, error) {
	every, err := s.everyLink()
	if err != nil {
		return nil, err
	}
	where, args := readBy(kept).whereTasks(taskLabels)
	rows, err := s.on.Query(
		`SELECT `+taskColumns+` FROM tasks`+where+` `+taskOrder, args...)
	if err != nil {
		return nil, fmt.Errorf("reading the store: %w", err)
	}
	defer func() { _ = rows.Close() }()

	tasks := []task.Task{}
	for rows.Next() {
		t, err := s.scanTask(rows, every)
		if err != nil {
			return nil, fmt.Errorf("reading a task: %w", err)
		}
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading the store: %w", err)
	}
	return tasks, nil
}
