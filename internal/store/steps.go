package store

import (
	"fmt"

	"cairn/internal/task"
)

func (s *Store) keepSteps(number int, steps []task.Step) error {
	if _, err := s.on.Exec(`DELETE FROM steps WHERE task_number = ?`, number); err != nil {
		return fmt.Errorf("clearing the steps of task %d: %w", number, err)
	}
	for i, step := range steps {
		if _, err := s.on.Exec(`INSERT INTO steps (task_number, place, title, done) VALUES (?, ?, ?, ?)`,
			number, i+1, step.Title, step.Done); err != nil {
			return fmt.Errorf("keeping the steps of task %d: %w", number, err)
		}
	}
	return nil
}

func scanStep(r scanner) (task.Step, error) {
	var step task.Step
	err := r.Scan(&step.Title, &step.Done)
	return step, err
}

func (s *Store) stepsOf(number int) ([]task.Step, error) {
	steps, err := rowsOf(s, scanStep, `SELECT title, done FROM steps WHERE task_number = ? ORDER BY place`, number)
	if err != nil {
		return nil, fmt.Errorf("reading the steps of task %d: %w", number, err)
	}
	return steps, nil
}

func (s *Store) everyStep() (map[int][]task.Step, error) {
	rows, err := s.on.Query(`SELECT task_number, title, done FROM steps ORDER BY task_number, place`)
	if err != nil {
		return nil, fmt.Errorf("reading the steps: %w", err)
	}
	defer func() { _ = rows.Close() }()
	out := map[int][]task.Step{}
	for rows.Next() {
		var number int
		var step task.Step
		if err := rows.Scan(&number, &step.Title, &step.Done); err != nil {
			return nil, fmt.Errorf("reading the steps: %w", err)
		}
		out[number] = append(out[number], step)
	}
	return out, rows.Err()
}
