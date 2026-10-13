package store

import (
	"fmt"
	"time"

	"cairn/internal/trace"
)

const traceColumns = `task_number, did, field, was, became, through, at`

// Record writes down one act.
func (s *Store) Record(a trace.Act) error {
	b := a.Carried()
	if _, err := s.on.Exec(
		`INSERT INTO trace (`+traceColumns+`) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		b.Task, b.Did, b.Field, b.Was, b.Became, b.Through,
		b.At.UTC().Format(time.RFC3339)); err != nil {
		return fmt.Errorf("writing down what happened to task %d: %w", b.Task, err)
	}
	return nil
}

// ActsOn hands back everything done to a task, oldest first.
func (s *Store) ActsOn(number int) ([]trace.Act, error) {
	acts, err := rowsOf(s, scanAct,
		`SELECT `+traceColumns+` FROM trace WHERE task_number = ? ORDER BY id`, number)
	if err != nil {
		return nil, fmt.Errorf("reading what happened to task %d: %w", number, err)
	}
	return acts, nil
}

// ActsSince hands back everything done from a moment onward, across every task, oldest first.
func (s *Store) ActsSince(from time.Time) ([]trace.Act, error) {
	acts, err := rowsOf(s, scanAct,
		`SELECT `+traceColumns+` FROM trace WHERE at >= ? ORDER BY id`,
		from.UTC().Format(time.RFC3339))
	if err != nil {
		return nil, fmt.Errorf("reading what happened since %s: %w", from, err)
	}
	return acts, nil
}

func scanAct(row scanner) (trace.Act, error) {
	var (
		task                             int
		did, field, was, became, through string
		at                               string
	)
	if err := row.Scan(&task, &did, &field, &was, &became, &through, &at); err != nil {
		return trace.Act{}, err
	}
	when, err := time.Parse(time.RFC3339, at)
	if err != nil {
		return trace.Act{}, fmt.Errorf("the moment %q cannot be read: %w", at, err)
	}

	return trace.Made(trace.Carried{Task: task, Did: did, Field: field, Was: was, Became: became, Through: through, At: when}), nil
}
