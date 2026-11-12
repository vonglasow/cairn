package engine

import (
	"errors"
	"maps"
	"slices"

	"cairn/internal/domain"
	"cairn/internal/objective"
)

// What a change to what an objective waits on can be turned down for.
var (
	ErrWaitsOnRequired    = errors.New("what an objective waits on is an objective number")
	ErrWaitsOnBothWays    = errors.New("what an objective waits on is either replaced or changed, not both")
	ErrWaitsOnSignAtBirth = errors.New("what an objective waits on is added or taken away only when changing it")
)

func (e *Engine) mayWait(o objective.Objective) error {
	on := o.Carried().WaitsOn
	if len(on) == 0 {
		return nil
	}
	t, err := e.tree()
	if err != nil {
		return err
	}
	for _, n := range on {
		if _, err := e.objectives.FindObjective(n); err != nil {
			return err
		}
	}
	return objective.MayWaitOn(o, t.by)
}

func (e *Engine) unmetEach(was map[int]domain.State) ([]string, error) {
	var said []string
	for _, n := range slices.Sorted(maps.Keys(was)) {
		more, err := e.unmet(n, was[n])
		if err != nil {
			return nil, err
		}
		for _, one := range more {
			if !slices.Contains(said, one) {
				said = append(said, one)
			}
		}
	}
	return said, nil
}

func (e *Engine) statesOf(numbers []int) (map[int]domain.State, error) {
	was := map[int]domain.State{}
	for _, n := range numbers {
		t, err := e.tasks.Find(n)
		if err != nil {
			return nil, err
		}
		was[n] = factsOf(t).State
	}
	return was, nil
}

func (e *Engine) unmet(number int, was domain.State) ([]string, error) {
	if was == domain.Current {
		return nil, nil
	}
	now, err := e.tasks.Find(number)
	if err != nil || factsOf(now).State != domain.Current {
		return nil, err
	}
	t, err := e.tree()
	if err != nil {
		return nil, err
	}
	var said []string
	for _, n := range slices.Sorted(slices.Values(t.counted(factsOf(now).Objectives))) {
		said = append(said, objective.Unmet(t.by[n], t.by)...)
	}
	return said, nil
}
