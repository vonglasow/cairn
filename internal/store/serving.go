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
