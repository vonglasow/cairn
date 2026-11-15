package engine

import (
	"cairn/internal/engine/item"
	"maps"
	"slices"

	"cairn/internal/engine/objective"
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
	return objective.MayWaitOn(o, t.By)
}

func (e *Engine) unmetEach(was map[int]item.State) ([]string, error) {
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

func (e *Engine) statesOf(numbers []int) (map[int]item.State, error) {
	was := map[int]item.State{}
	for _, n := range numbers {
		t, err := e.tasks.Find(n)
		if err != nil {
			return nil, err
		}
		was[n] = t.Carried().State
	}
	return was, nil
}

func (e *Engine) unmet(number int, was item.State) ([]string, error) {
	if was == item.Current {
		return nil, nil
	}
	now, err := e.tasks.Find(number)
	if err != nil || now.Carried().State != item.Current {
		return nil, err
	}
	t, err := e.tree()
	if err != nil {
		return nil, err
	}
	var said []string
	for _, n := range slices.Sorted(slices.Values(t.Counted(now.Carried().Objectives))) {
		said = append(said, objective.Unmet(t.By[n], t.By)...)
	}
	return said, nil
}
