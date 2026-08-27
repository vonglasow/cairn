package store

// Which task is being worked on. One row holds it, so there is at most one of
// them and nothing has to be kept in step; and the row's reference to the task
// is declared ON DELETE SET NULL, so a task deleted takes the focus with it.

import (
	"database/sql"
	"errors"
	"fmt"
)

// NoFocus is what Focus answers when no task is being worked on. Nothing in
// focus is an ordinary state, not a store that failed to answer.
const NoFocus = 0

// Focus says which task is being worked on, or NoFocus when none is.
func (s *Store) Focus() (int, error) {
	var number sql.NullInt64
	err := s.db.QueryRow(`SELECT task_number FROM focus WHERE id = 1`).Scan(&number)
	// A store that has lost its one row says the same thing as one whose row
	// holds nothing: no task is being worked on. Asked one at a time rather
	// than as one condition, so each can be judged on its own.
	if errors.Is(err, sql.ErrNoRows) {
		return NoFocus, nil
	}
	if err != nil {
		return NoFocus, fmt.Errorf("reading what is in focus: %w", err)
	}
	if !number.Valid {
		return NoFocus, nil
	}
	return int(number.Int64), nil
}

// SetFocus makes a task the one being worked on, releasing whichever was
// before. That the task exists is settled before this is called.
func (s *Store) SetFocus(number int) error {
	return s.focusOn(sql.NullInt64{Int64: int64(number), Valid: true})
}

// ClearFocus leaves nothing in focus. Letting go of nothing is not a failure,
// so this says the same whether or not a task was held.
func (s *Store) ClearFocus() error {
	return s.focusOn(sql.NullInt64{})
}

// releaseFocusOn leaves nothing in focus when the task named was the one being
// worked on, and leaves the focus alone otherwise. Called when a task is
// deleted, since the reference the schema declares is not enforced.
func (s *Store) releaseFocusOn(number int) error {
	if _, err := s.db.Exec(
		`UPDATE focus SET task_number = NULL WHERE task_number = ?`, number); err != nil {
		return fmt.Errorf("releasing the focus on task %d: %w", number, err)
	}
	return nil
}

// focusOn writes the one row, making it when a store has somehow lost it.
func (s *Store) focusOn(number sql.NullInt64) error {
	if _, err := s.db.Exec(
		`INSERT INTO focus (id, task_number) VALUES (1, ?)
		 ON CONFLICT (id) DO UPDATE SET task_number = excluded.task_number`,
		number); err != nil {
		return fmt.Errorf("setting what is in focus: %w", err)
	}
	return nil
}
