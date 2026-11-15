package engine

import (
	"cairn/internal/engine/item"
	"fmt"
	"slices"
	"time"

	"cairn/internal/engine/act"
	"cairn/internal/engine/task"
)

func (e *Engine) historyOf(number int) ([]act.Act, error) {
	acts, err := e.history.ActsOn(number)
	if err != nil {
		return nil, err
	}
	if len(acts) == 0 {
		return nil, fmt.Errorf("task %d: %w", number, task.ErrNoSuchTask)
	}
	return acts, nil
}

func (e *Engine) stoodSince(number int) (time.Time, error) {
	acts, err := e.historyOf(number)
	if err != nil {
		return time.Time{}, err
	}
	return act.LastNamed(acts, "column"), nil
}

func (e *Engine) since(day time.Time) ([]act.Act, error) {
	return e.history.ActsSince(day)
}

func (e *Engine) recordCreated(t task.Task) error {
	f := t.Carried()
	if err := e.record(f.Number, act.Created, "", "", ""); err != nil {
		return err
	}
	for _, a := range act.TheAttributes {
		value := a.Value(f)
		if value == "" {
			continue
		}
		if err := e.record(f.Number, act.Created, a.Named, "", value); err != nil {
			return err
		}
	}
	return e.recordOtherEnds(f.Number, nil, f.Linked)
}

func (e *Engine) recordChanges(was, became task.Task) error {
	before, after := was.Carried(), became.Carried()
	for _, a := range act.TheAttributes {
		from, to := a.Value(before), a.Value(after)
		if from == to {
			continue
		}
		if err := e.record(after.Number, act.Changed, a.Named, from, to); err != nil {
			return err
		}
	}
	if err := e.recordOtherEnds(after.Number, before.Linked, after.Linked); err != nil {
		return err
	}
	if (after.State == item.Done || after.State == item.Blocked) && before.State != after.State {
		return e.letGoOf(after.Number)
	}
	return nil
}

func (e *Engine) recordOtherEnds(n int, was, became []int) error {
	for _, other := range act.GainedOrLost(was, became) {
		found, err := e.tasks.Find(other)
		if err != nil {
			return err
		}
		after := found.Carried().Linked
		before := act.Without(after, n)
		if !slices.Contains(became, other) {
			before = append(slices.Clone(after), n)
			slices.Sort(before)
		}
		if err := e.record(other, act.Changed, "links", act.NumbersOrNothing(before), act.NumbersOrNothing(after)); err != nil {
			return err
		}
	}
	return nil
}

func (e *Engine) record(number int, did, field, was, became string) error {
	return e.history.Record(act.Made(act.Carried{
		Task: number, Did: did, Field: field, Was: was, Became: became, Through: e.acting.Through, By: e.acting.By, At: e.home.now(),
	}))
}
