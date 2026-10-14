package engine

import (
	"time"

	"cairn/internal/domain"
	"cairn/internal/task"
	"cairn/internal/trace"
)

type spentTasks struct {
	tasks
	history history
}

func (s spentTasks) spent() (map[int]time.Duration, error) {
	acts, err := s.history.ActsSince(time.Time{})
	if err != nil {
		return nil, err
	}
	return trace.Spent(acts, time.Now()), nil
}

func (s spentTasks) told(found []task.Task, err error) ([]task.Task, error) {
	if err != nil {
		return nil, err
	}
	spent, err := s.spent()
	if err != nil {
		return nil, err
	}
	for i, t := range found {
		found[i] = withSpent(t, spent)
	}
	return found, nil
}

func withSpent(t task.Task, spent map[int]time.Duration) task.Task {
	c := t.Carried()
	c.Spent = spent[c.Number]
	return task.Made(c)
}

func (s spentTasks) Find(number int) (task.Task, error) {
	found, err := s.told(single(s.tasks.Find(number)))
	if err != nil {
		return task.Task{}, err
	}
	return found[0], nil
}

func (s spentTasks) Keep(t task.Task) (task.Task, error) {
	found, err := s.told(single(s.tasks.Keep(t)))
	if err != nil {
		return task.Task{}, err
	}
	return found[0], nil
}

func (s spentTasks) All() ([]task.Task, error) { return s.told(s.tasks.All()) }

func (s spentTasks) Narrowed(kept *domain.Narrowing) ([]task.Task, error) {
	return s.told(s.tasks.Narrowed(kept))
}

func (s spentTasks) Under(number int) ([]task.Task, error) { return s.told(s.tasks.Under(number)) }

func single(t task.Task, err error) ([]task.Task, error) { return []task.Task{t}, err }

func (e *Engine) spendOn(number int, was, becomes time.Duration) error {
	if becomes == was {
		return nil
	}
	return e.addTime(number, startOfDay(time.Now()), (becomes - was).String())
}
