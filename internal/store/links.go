package store

import "fmt"

type links interface {
	of(number int) (labels []string, blockers, objectives []int, err error)
}

type askedEach struct{ s *Store }

func (a askedEach) of(number int) ([]string, []int, []int, error) {
	labels, err := a.s.labelsOf(taskLabels, number)
	if err != nil {
		return nil, nil, nil, err
	}
	blockers, err := a.s.blockersOn(number)
	if err != nil {
		return nil, nil, nil, err
	}
	objectives, err := a.s.objectivesServedBy(number)
	if err != nil {
		return nil, nil, nil, err
	}
	return labels, blockers, objectives, nil
}

type readAtOnce struct {
	labels            map[int][]string
	blockers, serving map[int][]int
}

func (r readAtOnce) of(number int) ([]string, []int, []int, error) {
	return r.labels[number], r.blockers[number], r.serving[number], nil
}

func (s *Store) everyLink() (readAtOnce, error) {
	labels, err := pairs[string](s, `SELECT task_number, label FROM labels ORDER BY task_number, label`)
	if err != nil {
		return readAtOnce{}, fmt.Errorf("reading the labels: %w", err)
	}
	blockers, err := pairs[int](s, `SELECT task_number, blocker_number FROM blockers ORDER BY task_number, blocker_number`)
	if err != nil {
		return readAtOnce{}, fmt.Errorf("reading what blocks the tasks: %w", err)
	}
	serving, err := pairs[int](s, `SELECT task_number, objective_number FROM task_objectives ORDER BY task_number, objective_number`)
	if err != nil {
		return readAtOnce{}, fmt.Errorf("reading what the tasks serve: %w", err)
	}
	return readAtOnce{labels: labels, blockers: blockers, serving: serving}, nil
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
