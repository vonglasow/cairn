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
	blockers, err := rowsOf(s, one[int], blockersOf, number)
	if err != nil {
		return nil, fmt.Errorf("reading what blocks task %d: %w", number, err)
	}
	return blockers, nil
}

func (s *Store) keepBlockersOn(number int, blockers []int) error {
	return s.relink(clearWhatBlocks, keepBlocker, number, blockers, "what blocks")
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
	free, err := rowsOf(s, one[int], nothingBlocks, string(domain.Blocked), string(domain.Done))
	if err != nil {
		return nil, fmt.Errorf("reading what is no longer blocked: %w", err)
	}
	return free, nil
}
