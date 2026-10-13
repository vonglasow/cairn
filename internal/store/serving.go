package store

import "fmt"

const (
	keepServing  = `INSERT OR IGNORE INTO task_objectives (task_number, objective_number) VALUES (?, ?)`
	servedBy     = `SELECT objective_number FROM task_objectives WHERE task_number = ? ORDER BY objective_number`
	clearServing = `DELETE FROM task_objectives WHERE task_number = ?`
	servingOne   = `FROM tasks
		WHERE EXISTS (SELECT 1 FROM task_objectives
			WHERE task_objectives.task_number = tasks.number
			  AND task_objectives.objective_number = ?) `
)

func (s *Store) objectivesServedBy(number int) ([]int, error) {
	served, err := rowsOf(s, one[int], servedBy, number)
	if err != nil {
		return nil, fmt.Errorf("reading what task %d serves: %w", number, err)
	}
	return served, nil
}

func (s *Store) keepServedBy(number int, objectives []int) error {
	return s.relink(clearServing, keepServing, number, objectives, "what serves")
}

func (s *Store) relink(clear, keep string, number int, others []int, what string) error {
	if _, err := s.on.Exec(clear, number); err != nil {
		return fmt.Errorf("clearing %s task %d: %w", what, number, err)
	}
	for _, other := range others {
		if _, err := s.on.Exec(keep, number, other); err != nil {
			return fmt.Errorf("keeping %s task %d: %w", what, number, err)
		}
	}
	return nil
}
