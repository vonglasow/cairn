package store

// Narrowing a read down to what was asked for. It happens here, in SQL, rather
// than by reading everything and dropping most of it above: what a narrowed
// listing is for is not handing over what nobody asked for, and a filter that
// reads the whole store first only hides the cost.

import (
	"fmt"
	"strings"

	"cairn/internal/task"
)

// Narrowing is what a read was told to keep. Facts are named one at a time,
// and a fact never named narrows nothing — so a Narrowing nobody named
// anything on reads the store whole.
//
// It carries the SQL rather than the values because that is where the two can
// be kept from drifting: every clause below sits beside the placeholder it
// needs, and a value never reaches the query as text.
type Narrowing struct {
	clauses []string
	args    []any
	labels  []string
}

// InColumn keeps what sits in one column.
func (n *Narrowing) InColumn(state string) {
	n.keep(`state = ?`, state)
}

// Carrying keeps what carries one priority.
func (n *Narrowing) Carrying(priority string) {
	n.keep(`priority = ?`, priority)
}

// Serving keeps what serves one objective. An objective nothing serves keeps
// nothing rather than failing: a filter says what is wanted, and whether it
// exists is a question about objectives that reading tasks does not ask.
func (n *Narrowing) Serving(objective int) {
	n.keep(`objective = ?`, objective)
}

// Labelled keeps what carries every label named. Every one and not any: naming
// two labels is naming two conditions.
//
// Held aside rather than turned into a clause here, because where labels are
// kept depends on what is being read and not on what was asked: a task and an
// objective keep theirs in tables of their own. The read says which.
func (n *Narrowing) Labelled(labels []string) {
	n.labels = labels
}

// Narrows says whether anything was named at all. What is done with the answer
// depends on it: a listing nobody narrowed is shown as the tree it is, and a
// narrowed one is shown flat, since where a task sits is not what was asked.
func (n *Narrowing) Narrows() bool {
	return len(n.clauses)+len(n.labels) > 0
}

// keep records one clause and the value it holds to.
func (n *Narrowing) keep(clause string, value any) {
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
func (n *Narrowing) where(labels labelStore) (string, []any) {
	if !n.Narrows() {
		return "", nil
	}
	clauses, args := n.clauses, n.args
	for _, label := range n.labels {
		clauses = append(clauses, labels.carrier)
		args = append(args, label)
	}
	return ` WHERE ` + strings.Join(clauses, ` AND `), args
}

// Narrowed hands back the tasks a narrowing keeps, in the one order there is.
func (s *Store) Narrowed(narrowing *Narrowing) ([]task.Task, error) {
	where, args := narrowing.where(taskLabels)
	rows, err := s.db.Query(
		`SELECT `+taskColumns+` FROM tasks`+where+` `+taskOrder, args...)
	if err != nil {
		return nil, fmt.Errorf("reading the store: %w", err)
	}
	defer func() { _ = rows.Close() }()

	tasks := []task.Task{}
	for rows.Next() {
		t, err := s.scanTask(rows)
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
