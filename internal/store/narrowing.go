package store

// Narrowing a read down to what was asked for. It happens here, in SQL, rather
// than by reading everything and dropping most of it above: what a narrowed
// listing is for is not handing over what nobody asked for, and a filter that
// reads the whole store first only hides the cost.

import (
	"fmt"
	"strings"

	"cairn/internal/domain"
	"cairn/internal/task"
)

// narrowing is what a read was told to keep, turned into SQL: it is told the
// facts a domain.Narrowing names, one at a time, and keeps each as a clause.
//
// It carries the SQL rather than the values because that is where the two can
// be kept from drifting: every clause below sits beside the placeholder it
// needs, and a value never reaches the query as text.
type narrowing struct {
	clauses []string
	args    []any
	labels  []string
	// archived keeps what was taken off the board instead of what is on it;
	// both keeps either.
	archived bool
	both     bool
}

// readBy turns what a read was told to keep into the clauses that keep it.
func readBy(kept *domain.Narrowing) *narrowing {
	n := &narrowing{}
	kept.Tell(n)
	return n
}

// Archived keeps the tasks taken off the board, instead of the ones still on it.
func (n *narrowing) Archived() { n.archived = true }

// WithArchived keeps both, which is what a board asks for: it draws the two
// apart rather than reading the store twice.
func (n *narrowing) WithArchived() { n.both = true }

// InColumn keeps what sits in one column.
func (n *narrowing) InColumn(state string) {
	n.keep(`state = ?`, state)
}

// DueBefore keeps what is due on or before a day, the day itself included: a
// deadline is the last day something may be done, so a listing of what is due
// by the first that left out the first would answer a different question.
//
// What carries no deadline is left out. It is not due by that day — it is not
// due at all — and a listing of what is coming up that answered every dateless
// task would answer most of the board.
func (n *narrowing) DueBefore(day string) {
	n.keep(`deadline != '' AND deadline <= ?`, day)
}

// Carrying keeps what carries one priority.
func (n *narrowing) Carrying(priority string) {
	n.keep(`priority = ?`, priority)
}

// Serving keeps what serves one objective, among however many others it also
// serves. An objective nothing serves keeps nothing rather than failing: a
// filter says what is wanted, and whether it exists is a question about
// objectives that reading tasks does not ask.
func (n *narrowing) Serving(objective int) {
	n.keep(`EXISTS (SELECT 1 FROM task_objectives
		WHERE task_objectives.task_number = tasks.number
		  AND task_objectives.objective_number = ?)`, objective)
}

// Labelled keeps what carries every label named. Every one and not any: naming
// two labels is naming two conditions.
//
// Held aside rather than turned into a clause here, because where labels are
// kept depends on what is being read and not on what was asked: a task and an
// objective keep theirs in tables of their own. The read says which.
func (n *narrowing) Labelled(labels []string) {
	n.labels = labels
}

// narrows says whether anything was named at all.
func (n *narrowing) narrows() bool {
	return len(n.clauses)+len(n.labels) > 0 || n.archived
}

// keep records one clause and the value it holds to.
func (n *narrowing) keep(clause string, value any) {
	n.clauses = append(n.clauses, clause)
	n.args = append(n.args, value)
}

// where turns what was named into the tail of a query, and the values to hand
// it. Nothing named gives no tail at all rather than a clause that is always
// true, so a read nobody narrowed is the read it always was.
//
// The string is built from the clauses above and nothing else. No value is ever
// part of it: each sits behind its own placeholder, which is why naming a label
// called "' OR 1=1" narrows to the tasks carrying that label and to nothing
// else.
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

// whereTasks is the same for tasks, which alone can be taken off the board.
// Its own method because a narrowing is shared with objectives, which cannot.
func (n *narrowing) whereTasks(labels labelStore) (string, []any) {
	// Always a clause, named or not — unless both are wanted.
	var clauses []string
	var args []any
	if !n.both {
		clauses, args = []string{`archived = ?`}, []any{n.archived}
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
