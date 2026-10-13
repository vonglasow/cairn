package store

import (
	"fmt"

	"cairn/internal/domain"
)

const (
	keepBlocker     = `INSERT OR IGNORE INTO blockers (task_number, blocker_number) VALUES (?, ?)`
	blockersOf      = `SELECT blocker_number FROM blockers WHERE task_number = ? ORDER BY blocker_number`
	clearBlockers   = `DELETE FROM blockers WHERE task_number = ? OR blocker_number = ?`
	clearWhatBlocks = `DELETE FROM blockers WHERE task_number = ?`
)

func (s *Store) blockersOn(number int) ([]int, error) {
	rows, err := s.on.Query(blockersOf, number)
	if err != nil {
		return nil, fmt.Errorf("reading what blocks task %d: %w", number, err)
	}
	defer func() { _ = rows.Close() }()

	blockers := []int{}
	for rows.Next() {
		var blocker int
		if err := rows.Scan(&blocker); err != nil {
			return nil, fmt.Errorf("reading a blocker: %w", err)
		}
		blockers = append(blockers, blocker)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading what blocks task %d: %w", number, err)
	}
	return blockers, nil
}

func (s *Store) keepBlockersOn(number int, blockers []int) error {
	if _, err := s.on.Exec(clearWhatBlocks, number); err != nil {
		return fmt.Errorf("clearing what blocks task %d: %w", number, err)
	}
	for _, blocker := range blockers {
		if _, err := s.on.Exec(keepBlocker, number, blocker); err != nil {
			return fmt.Errorf("keeping what blocks task %d: %w", number, err)
		}
	}
	return nil
}

func (s *Store) releaseBlockersOn(number int) error {
	if _, err := s.on.Exec(clearBlockers, number, number); err != nil {
		return fmt.Errorf("releasing what blocked task %d: %w", number, err)
	}
	return nil
}

const nothingBlocks = `
SELECT number FROM tasks
WHERE state = ?
  AND because = ''
  AND EXISTS (SELECT 1 FROM blockers WHERE task_number = tasks.number)
  AND NOT EXISTS (
	SELECT 1 FROM blockers
	JOIN tasks AS blocker ON blocker.number = blockers.blocker_number
	WHERE blockers.task_number = tasks.number AND blocker.state <> ?)
ORDER BY number`

// NothingBlocks reads which tasks are waiting on nothing any more.
func (s *Store) NothingBlocks() ([]int, error) {
	rows, err := s.on.Query(nothingBlocks, string(domain.Blocked), string(domain.Done))
	if err != nil {
		return nil, fmt.Errorf("reading what is no longer blocked: %w", err)
	}
	defer func() { _ = rows.Close() }()

	free := []int{}
	for rows.Next() {
		var number int
		if err := rows.Scan(&number); err != nil {
			return nil, fmt.Errorf("reading a task no longer blocked: %w", err)
		}
		free = append(free, number)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading what is no longer blocked: %w", err)
	}
	return free, nil
}
