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
	focus    []string
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

// InFocus keeps what is being worked on, or what is not; it means something only to a task.
func (n *narrowing) InFocus(held bool) {
	clause := `number IN (SELECT task_number FROM focused)`
	if !held {
		clause = `number NOT IN (SELECT task_number FROM focused)`
	}
	n.focus = append(n.focus, clause)
}

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

// Serving keeps what serves any one of the objectives named together, among however many others it also serves.
func (n *narrowing) Serving(objectives []int) {
	n.clauses = append(n.clauses, `EXISTS (SELECT 1 FROM task_objectives
		WHERE task_objectives.task_number = tasks.number
		  AND task_objectives.objective_number IN (`+placeholders(len(objectives))+`))`)
	for _, o := range objectives {
		n.args = append(n.args, o)
	}
}

func placeholders(n int) string {
	return strings.TrimSuffix(strings.Repeat("?, ", n), ", ")
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
	return whereAll(clauses), args
}

func whereAll(clauses []string) string {
	return fmt.Sprintf(` WHERE %s`, strings.Join(clauses, ` AND `))
}

const offTheBoard = `archived`

func (n *narrowing) whereTasks(labels labelStore) (string, []any) {
	var clauses []string
	var args []any
	if !n.both {
		clauses, args = []string{offTheBoard + ` = ?`}, []any{n.archived}
	}
	clauses = append(append(clauses, n.clauses...), n.focus...)
	args = append(args, n.args...)
	for _, label := range n.labels {
		clauses = append(clauses, labels.carrier)
		args = append(args, label)
	}
	if len(clauses) == 0 {
		return "", nil
	}
	return whereAll(clauses), args
}

// Narrowed hands back the tasks a narrowing keeps, in the one order there is.
func (s *Store) Narrowed(kept *domain.Narrowing) ([]task.Task, error) {
	every, err := s.everyLink()
	if err != nil {
		return nil, err
	}
	where, args := readBy(kept).whereTasks(taskLabels)
	tasks, err := rowsOf(s, func(r scanner) (task.Task, error) { return s.scanTask(r, every) },
		fmt.Sprintf(`SELECT %s FROM tasks%s %s`, taskColumns, where, taskOrder), args...)
	if err != nil {
		return nil, fmt.Errorf("reading the store: %w", err)
	}
	return tasks, nil
}
