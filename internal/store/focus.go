package store

import (
	"fmt"
	"time"
)

// Focus says which tasks are being worked on, oldest first, and nothing at all when none are.
func (s *Store) Focus() ([]int, error) {
	held, err := rowsOf(s, one[int], `SELECT task_number FROM focused ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("reading what is in focus: %w", err)
	}
	if len(held) == 0 {
		return nil, nil
	}
	return held, nil
}

// Held is a task in focus and the moment its focus was last used, nothing at all for one that never said.
type Held struct {
	Task int
	Seen time.Time
}

// FocusSeen says which tasks are in focus, oldest first, each with the moment its focus was last used.
func (s *Store) FocusSeen() ([]Held, error) {
	held, err := rowsOf(s, func(row scanner) (Held, error) {
		var (
			one  Held
			seen string
		)
		if err := row.Scan(&one.Task, &seen); err != nil {
			return one, err
		}
		if seen == "" {
			return one, nil
		}
		var err error
		one.Seen, err = time.Parse(time.RFC3339, seen)
		return one, err
	}, `SELECT task_number, seen FROM focused ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("reading when what is in focus was last used: %w", err)
	}
	return held, nil
}

// TouchFocus says that every task in focus was used at that moment.
func (s *Store) TouchFocus(at time.Time) error {
	if _, err := s.on.Exec(`UPDATE focused SET seen = ?`, at.UTC().Format(time.RFC3339)); err != nil {
		return fmt.Errorf("saying what is in focus was used: %w", err)
	}
	return nil
}

// AddFocus takes up a task, used at that moment.
func (s *Store) AddFocus(number int, at time.Time) error {
	if _, err := s.on.Exec(
		`INSERT OR IGNORE INTO focused (task_number, seen) VALUES (?, ?)`, number, at.UTC().Format(time.RFC3339)); err != nil {
		return fmt.Errorf("taking up task %d: %w", number, err)
	}
	return nil
}

// DropFocus lets go of one task and leaves the rest held.
func (s *Store) DropFocus(number int) error {
	if _, err := s.on.Exec(`DELETE FROM focused WHERE task_number = ?`, number); err != nil {
		return fmt.Errorf("letting go of task %d: %w", number, err)
	}
	return nil
}

// ClearFocus leaves nothing in focus at all.
func (s *Store) ClearFocus() error {
	if _, err := s.on.Exec(`DELETE FROM focused`); err != nil {
		return fmt.Errorf("letting go of everything in focus: %w", err)
	}
	return nil
}

func (s *Store) releaseFocusOn(number int) error {
	return s.DropFocus(number)
}
