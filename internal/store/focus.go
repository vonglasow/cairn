package store

import "fmt"

// Focus says which tasks are being worked on, oldest first, and nothing at all when none are.
func (s *Store) Focus() ([]int, error) {
	rows, err := s.on.Query(`SELECT task_number FROM focused ORDER BY id`)
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

// AddFocus takes up a task.
func (s *Store) AddFocus(number int) error {
	if _, err := s.on.Exec(
		`INSERT OR IGNORE INTO focused (task_number) VALUES (?)`, number); err != nil {
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
