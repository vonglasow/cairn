package store

// What blocks a task. A blocker is a link between two tasks, so it lives in a
// table of its own; the words beside it are a fact of the task, so they live in
// a column. Both are read and written here.

import (
	"fmt"

	"cairn/internal/domain"
)

// The queries about what blocks a task, written out for the reason labelStore
// gives: a table name interpolated into SQL reads as harmless right up to the
// day the name comes from somewhere else.
const (
	keepBlocker = `INSERT OR IGNORE INTO blockers (task_number, blocker_number) VALUES (?, ?)`
	// Read in the order the blockers were numbered, so a card lays them out the
	// same way twice.
	blockersOf = `SELECT blocker_number FROM blockers WHERE task_number = ? ORDER BY blocker_number`
	// Both directions of the link, for a task on its way out: what blocked it,
	// and what it was blocking.
	clearBlockers = `DELETE FROM blockers WHERE task_number = ? OR blocker_number = ?`
	// Only the ones it declared, for a task being changed: a task's own edit
	// says what blocks it, and never what it blocks.
	clearWhatBlocks = `DELETE FROM blockers WHERE task_number = ?`
)

// blockersOn reads back what a task said blocks it.
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

// keepBlockersOn records what a task says blocks it, replacing whatever it said
// before: a task carries a set, so one left out is one taken away.
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

// releaseBlockersOn takes away both sides of every link a task was part of. By
// hand, for the reason Delete gives about the labels: the foreign keys the
// schema declares do not fire, because this connection leaves them off.
func (s *Store) releaseBlockersOn(number int) error {
	if _, err := s.on.Exec(clearBlockers, number, number); err != nil {
		return fmt.Errorf("releasing what blocked task %d: %w", number, err)
	}
	return nil
}

// NothingBlocks hands back the tasks sitting in the blocked column that named
// blockers, named no words beside them, and whose every named blocker is done.
//
// Worked out here rather than above, because it is a question about the store
// and not about any one task: a listing narrowed to one column cannot see the
// states of the tasks that block its rows, and a task cannot see them at all.
//
// The three conditions are what makes the answer honest. Blockers were named,
// or there is nothing cairn was told about. No words, since "the client's
// answer" is either in or out and cairn has no way to know which. And every one
// of them done, not merely one.
//
// The two columns are bound rather than spelled, so their names are the ones
// domain gives them and nowhere else.
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
