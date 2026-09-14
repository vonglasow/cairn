package store

// Which tasks are being worked on. A row for each, keyed by the task so it
// cannot be held twice, and the order the rows were made is the order they
// were taken up — which is the order they are answered in.
//
// It held one, in one row of one table. A person works on one thing at a time,
// which is what that was written for; but an agent session is started per
// task, and several run at once, so a focus that held one made every other one
// invisible to whoever asked what was being worked on.

import "fmt"

// NoFocus is the number no task carries, for a caller that has to name "none".
const NoFocus = 0

// Focus says which tasks are being worked on, oldest first, and nothing at all
// when none are. Nothing in focus is an ordinary answer, not a store that
// failed to answer.
func (s *Store) Focus() ([]int, error) {
	rows, err := s.db.Query(`SELECT task_number FROM focused ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("reading what is in focus: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var held []int
	for rows.Next() {
		var number int
		if err := rows.Scan(&number); err != nil {
			return nil, fmt.Errorf("reading what is in focus: %w", err)
		}
		held = append(held, number)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading what is in focus: %w", err)
	}
	return held, nil
}

// AddFocus takes up a task. One already being worked on keeps the place in the
// order it had rather than moving to the end: taking up what you already hold
// is not a change. That the task exists is settled before this is called.
func (s *Store) AddFocus(number int) error {
	if _, err := s.db.Exec(
		`INSERT OR IGNORE INTO focused (task_number) VALUES (?)`, number); err != nil {
		return fmt.Errorf("taking up task %d: %w", number, err)
	}
	return nil
}

// DropFocus lets go of one task and leaves the rest held. Letting go of a task
// that was not being worked on is not a failure, so this says the same either
// way.
func (s *Store) DropFocus(number int) error {
	if _, err := s.db.Exec(`DELETE FROM focused WHERE task_number = ?`, number); err != nil {
		return fmt.Errorf("letting go of task %d: %w", number, err)
	}
	return nil
}

// ClearFocus leaves nothing in focus at all.
func (s *Store) ClearFocus() error {
	if _, err := s.db.Exec(`DELETE FROM focused`); err != nil {
		return fmt.Errorf("letting go of everything in focus: %w", err)
	}
	return nil
}

// releaseFocusOn lets go of a task on its way to being deleted. Called from
// the delete, since the reference the schema declares is never enforced.
func (s *Store) releaseFocusOn(number int) error {
	return s.DropFocus(number)
}
