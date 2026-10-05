package store

// What a task serves. A task serves a set of objectives, so the link lives in a
// table of its own rather than in a column — the same shape as the labels and
// the blockers, and for the same reason.
//
// Nothing here releases the links of a deleted objective, and nothing has to: an
// objective with tasks under it cannot be deleted at all, so there is never a
// link left pointing at one that is gone. A function for it was written and
// taken back out — it could not be reached, so nothing could vouch for it.

import "fmt"

// The queries about what a task serves, written out for the reason labelStore
// gives: a table name interpolated into SQL reads as harmless right up to the
// day the name comes from somewhere else.
const (
	keepServing = `INSERT OR IGNORE INTO task_objectives (task_number, objective_number) VALUES (?, ?)`
	// Read in the order the objectives were numbered, so a card lays them out
	// the same way twice.
	servedBy = `SELECT objective_number FROM task_objectives WHERE task_number = ? ORDER BY objective_number`
	// Only what this task serves, for a task being changed.
	clearServing = `DELETE FROM task_objectives WHERE task_number = ?`
	// Which tasks serve one objective. Joined to the tasks by the reader, since
	// a query built out of taskColumns cannot be a package-level constant: the
	// concatenation would run outside any function, where the gate cannot judge
	// a mutation of it. See docs/checks.md.
	servingOne = `FROM tasks
		WHERE EXISTS (SELECT 1 FROM task_objectives
			WHERE task_objectives.task_number = tasks.number
			  AND task_objectives.objective_number = ?) `
)

// objectivesServedBy reads back what a task said it serves.
func (s *Store) objectivesServedBy(number int) ([]int, error) {
	rows, err := s.on.Query(servedBy, number)
	if err != nil {
		return nil, fmt.Errorf("reading what task %d serves: %w", number, err)
	}
	defer func() { _ = rows.Close() }()

	served := []int{}
	for rows.Next() {
		var objective int
		if err := rows.Scan(&objective); err != nil {
			return nil, fmt.Errorf("reading an objective served: %w", err)
		}
		served = append(served, objective)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading what task %d serves: %w", number, err)
	}
	return served, nil
}

// keepServedBy records what a task serves, replacing whatever it served before:
// a task carries a set, so one left out is one taken away.
func (s *Store) keepServedBy(number int, objectives []int) error {
	if _, err := s.on.Exec(clearServing, number); err != nil {
		return fmt.Errorf("clearing what task %d serves: %w", number, err)
	}
	for _, objective := range objectives {
		if _, err := s.on.Exec(keepServing, number, objective); err != nil {
			return fmt.Errorf("keeping what task %d serves: %w", number, err)
		}
	}
	return nil
}
