package sqlite

import (
	"fmt"
	"time"

	"cairn/internal/engine/act"
)

const traceColumns = `task_number, did, field, was, became, through, author, at`

// Record writes down one act.
func (s *Store) Record(a act.Act) error {
	b := a.Carried()
	if _, err := s.on.Exec(
		`INSERT INTO trace (`+traceColumns+`) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		b.Task, b.Did, b.Field, b.Was, b.Became, b.Through, b.By,
		b.At.UTC().Format(time.RFC3339)); err != nil {
		return fmt.Errorf("writing down what happened to task %d: %w", b.Task, err)
	}
	return nil
}

// ActsOn hands back everything done to a task, oldest first.
func (s *Store) ActsOn(number int) ([]act.Act, error) {
	acts, err := rowsOf(s, scanAct,
		`SELECT `+traceColumns+` FROM trace WHERE task_number = ? ORDER BY id`, number)
	if err != nil {
		return nil, fmt.Errorf("reading what happened to task %d: %w", number, err)
	}
	return acts, nil
}

// ActsSince hands back everything done from a moment onward, across every task, oldest first.
func (s *Store) ActsSince(from time.Time) ([]act.Act, error) {
	acts, err := rowsOf(s, scanAct,
		fmt.Sprintf(`SELECT %s FROM trace WHERE at >= ? ORDER BY id`, traceColumns),
		from.UTC().Format(time.RFC3339))
	if err != nil {
		return nil, fmt.Errorf("reading what happened since %s: %w", from, err)
	}
	return acts, nil
}

func scanAct(row scanner) (act.Act, error) {
	var (
		task                             int
		did, field, was, became, through string
		by, at                           string
	)
	if err := row.Scan(&task, &did, &field, &was, &became, &through, &by, &at); err != nil {
		return act.Act{}, err
	}
	when, err := time.Parse(time.RFC3339, at)
	if err != nil {
		return act.Act{}, fmt.Errorf("the moment %q cannot be read: %w", at, err)
	}

	return act.Made(act.Carried{Task: task, Did: did, Field: field, Was: was, Became: became, Through: through, By: by, At: when}), nil
}
