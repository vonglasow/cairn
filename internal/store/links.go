package store

import (
	"fmt"

	"cairn/internal/task"
)

type linked struct {
	labels               []string
	blockers, objectives []int
	linked               []int
	steps                []task.Step
}

type links interface {
	of(number int) (linked, error)
}

type askedEach struct{ s *Store }

func (a askedEach) of(number int) (linked, error) {
	labels, err := a.s.labelsOf(taskLabels, number)
	if err != nil {
		return linked{}, err
	}
	blockers, err := a.s.blockersOn(number)
	if err != nil {
		return linked{}, err
	}
	objectives, err := a.s.objectivesServedBy(number)
	if err != nil {
		return linked{}, err
	}
	steps, err := a.s.stepsOf(number)
	if err != nil {
		return linked{}, err
	}
	others, err := a.s.linkedTo(number)
	if err != nil {
		return linked{}, err
	}
	return linked{labels: labels, blockers: blockers, objectives: objectives, steps: steps, linked: others}, nil
}

type readAtOnce struct {
	labels            map[int][]string
	blockers, serving map[int][]int
	steps             map[int][]task.Step
	linked            map[int][]int
}

func (r readAtOnce) of(number int) (linked, error) {
	return linked{labels: r.labels[number], blockers: r.blockers[number], objectives: r.serving[number], steps: r.steps[number], linked: r.linked[number]}, nil
}

func (s *Store) everyLink() (readAtOnce, error) {
	labels, err := pairs[string](s, `SELECT task_number, label FROM labels ORDER BY task_number, label`)
	if err != nil {
		return readAtOnce{}, fmt.Errorf("reading the labels: %w", err)
	}
	blockers, linked, err := s.everyBlockerAndLink()
	if err != nil {
		return readAtOnce{}, err
	}
	serving, err := pairs[int](s, `SELECT task_number, objective_number FROM task_objectives ORDER BY task_number, objective_number`)
	if err != nil {
		return readAtOnce{}, fmt.Errorf("reading what the tasks serve: %w", err)
	}
	steps, err := s.everyStep()
	if err != nil {
		return readAtOnce{}, err
	}
	return readAtOnce{labels: labels, blockers: blockers, serving: serving, steps: steps, linked: linked}, nil
}

func pairs[T any](s *Store, query string) (map[int][]T, error) {
	rows, err := s.on.Query(query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := map[int][]T{}
	for rows.Next() {
		var number int
		var carried T
		if err := rows.Scan(&number, &carried); err != nil {
			return nil, err
		}
		out[number] = append(out[number], carried)
	}
	return out, rows.Err()
}

func (s *Store) everyBlockerAndLink() (blockers, linked map[int][]int, err error) {
	rows, err := s.on.Query(`SELECT 0, task_number, blocker_number FROM blockers
		UNION ALL SELECT 1, low, high FROM links
		UNION ALL SELECT 1, high, low FROM links
		ORDER BY 2, 3`)
	if err != nil {
		return nil, nil, fmt.Errorf("reading what blocks and what links the tasks: %w", err)
	}
	defer func() { _ = rows.Close() }()
	blockers, linked = map[int][]int{}, map[int][]int{}
	for rows.Next() {
		var isLink, number, other int
		if err := rows.Scan(&isLink, &number, &other); err != nil {
			return nil, nil, err
		}
		if isLink == 1 {
			linked[number] = append(linked[number], other)
		} else {
			blockers[number] = append(blockers[number], other)
		}
	}
	return blockers, linked, rows.Err()
}
