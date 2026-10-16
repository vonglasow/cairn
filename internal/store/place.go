package store

import (
	"database/sql"
	"errors"
	"fmt"
)

const sisters = `state = ?`

func (s *Store) whereItSits(number int) (state string, place int, err error) {
	row := s.on.QueryRow(
		`SELECT state, position FROM tasks WHERE number = ?`, number)
	err = row.Scan(&state, &place)
	if errors.Is(err, sql.ErrNoRows) {
		return "", 0, fmt.Errorf("task %d: %w", number, ErrNoSuchTask)
	}
	if err != nil {
		return "", 0, fmt.Errorf("reading where task %d sits: %w", number, err)
	}
	return state, place, nil
}

// MoveFirst puts a task at the head of the tasks it sits with, and says whether it went anywhere.
func (s *Store) MoveFirst(number int) (bool, error) {
	return s.place(number, func(_, _ int) int { return 1 })
}

// MoveLast puts a task behind all the tasks it sits with, and says whether it went anywhere.
func (s *Store) MoveLast(number int) (bool, error) {
	return s.place(number, func(_, among int) int { return among })
}

// MoveAfter puts a task directly behind another.
func (s *Store) MoveAfter(number, target int) (bool, error) {
	_, behind, err := s.whereItSits(target)
	if err != nil {
		return false, err
	}
	return s.place(number, func(from, _ int) int {
		if from <= behind {
			return behind
		}
		return behind + 1
	})
}

func (s *Store) place(number int, chosen func(from, among int) int) (bool, error) {
	state, from, err := s.whereItSits(number)
	if err != nil {
		return false, err
	}
	among, err := s.countSisters(state)
	if err != nil {
		return false, err
	}
	to := chosen(from, among)
	if to == from {
		return false, nil
	}
	if err := s.makeRoom(state, from, to); err != nil {
		return false, err
	}
	if _, err := s.on.Exec(
		`UPDATE tasks SET position = ? WHERE number = ?`, to, number); err != nil {
		return false, fmt.Errorf("moving task %d: %w", number, err)
	}
	return true, nil
}

func (s *Store) countSisters(state string) (int, error) {
	var among int
	if err := s.on.QueryRow(
		`SELECT COUNT(*) FROM tasks WHERE `+sisters, state).Scan(&among); err != nil {
		return 0, fmt.Errorf("counting what task sits with: %w", err)
	}
	return among, nil
}

func (s *Store) closeTheGapLeftIn(state string, left int) error {
	if _, err := s.on.Exec(
		`UPDATE tasks SET position = position - 1
		 WHERE `+sisters+` AND position > ?`, state, left); err != nil {
		return fmt.Errorf("closing the gap left behind: %w", err)
	}
	return nil
}

func (s *Store) makeRoom(state string, from, to int) error {
	if _, err := s.on.Exec(
		`UPDATE tasks SET position = position + CASE WHEN ? < ? THEN -1 ELSE 1 END
		 WHERE `+sisters+` AND position BETWEEN ? AND ? AND position <> ?`,
		from, to, state, min(from, to), max(from, to), from); err != nil {
		return fmt.Errorf("making room: %w", err)
	}
	return nil
}
