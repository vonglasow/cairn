package store

import (
	"database/sql"
	"fmt"
	"slices"
)

// PlaceObjective puts an objective under a parent, none for level one, and among its siblings there: first when after is 0, last when after is -1, else directly behind the objective after.
func (s *Store) PlaceObjective(number, parent, after int) error {
	if _, err := s.FindObjective(number); err != nil {
		return err
	}
	was, err := s.parentOf(number)
	if err != nil {
		return err
	}
	siblings, err := s.siblings(parent, number)
	if err != nil {
		return err
	}
	at := len(siblings)
	switch {
	case after == 0:
		at = 0
	case after > 0:
		i := slices.Index(siblings, after)
		if i < 0 {
			return fmt.Errorf("objective %d does not sit where objective %d would go", after, number)
		}
		at = i + 1
	}
	siblings = slices.Insert(siblings, at, number)
	if _, err := s.on.Exec(`UPDATE objectives SET parent = ? WHERE number = ?`,
		sql.NullInt64{Int64: int64(parent), Valid: parent != 0}, number); err != nil {
		return fmt.Errorf("moving objective %d: %w", number, err)
	}
	if err := s.renumber(siblings); err != nil {
		return err
	}
	if was == parent {
		return nil
	}
	left, err := s.siblings(was, number)
	if err != nil {
		return err
	}
	return s.renumber(left)
}

func (s *Store) parentOf(number int) (int, error) {
	var parent sql.NullInt64
	if err := s.on.QueryRow(`SELECT parent FROM objectives WHERE number = ?`, number).Scan(&parent); err != nil {
		return 0, fmt.Errorf("reading where objective %d sits: %w", number, err)
	}
	return int(parent.Int64), nil
}

func (s *Store) siblings(parent, without int) ([]int, error) {
	return rowsOf(s, one[int],
		`SELECT number FROM objectives WHERE COALESCE(parent, 0) = ? AND number != ? ORDER BY position, number`, parent, without)
}

func (s *Store) renumber(numbers []int) error {
	for i, n := range numbers {
		if _, err := s.on.Exec(`UPDATE objectives SET position = ? WHERE number = ?`, i+1, n); err != nil {
			return fmt.Errorf("placing objective %d: %w", n, err)
		}
	}
	return nil
}
