package engine

import (
	"fmt"
	"strings"
	"time"

	"cairn/internal/domain"
	"cairn/internal/task"
	"cairn/internal/trace"
)

type history interface {
	Record(trace.Act) error
	ActsOn(int) ([]trace.Act, error)
	ActsSince(time.Time) ([]trace.Act, error)
}

var theAttributes = []struct {
	named string
	value func(taskFacts) string
}{
	{"title", func(f taskFacts) string { return f.title }},
	{"description", func(f taskFacts) string { return f.description }},
	{"column", func(f taskFacts) string { return string(f.state) }},
	{"priority", func(f taskFacts) string { return string(f.priority) }},
	{"labels", func(f taskFacts) string { return strings.Join(f.labels, ", ") }},
	{"blockers", func(f taskFacts) string { return numbersOrNothing(f.blockers) }},
	{"because", func(f taskFacts) string { return f.because }},
	{"deadline", func(f taskFacts) string { return domain.FormatDay(f.deadline) }},
	{"estimate", func(f taskFacts) string { return domain.FormatDuration(f.estimate) }},
	{"time spent", func(f taskFacts) string { return domain.FormatDuration(f.timeSpent) }},
	{"parent", func(f taskFacts) string { return numberOrNothing(f.parent) }},
	{"objectives", func(f taskFacts) string { return numbersOrNothing(f.objectives) }},
	{"archived", func(f taskFacts) string { return yesOrNothing(f.archived) }},
}

func (e *Engine) historyOf(number int) ([]trace.Act, error) {
	acts, err := e.history.ActsOn(number)
	if err != nil {
		return nil, err
	}
	if len(acts) == 0 {
		return nil, fmt.Errorf("task %d: %w", number, ErrNoSuchTask)
	}
	return acts, nil
}

func (e *Engine) stoodSince(number int) (time.Time, error) {
	acts, err := e.historyOf(number)
	if err != nil {
		return time.Time{}, err
	}
	return trace.LastNamed(acts, "column"), nil
}

func (e *Engine) since(day time.Time) ([]trace.Act, error) {
	return e.history.ActsSince(day)
}

func (e *Engine) recordCreated(t task.Task) error {
	f := factsOf(t)
	if err := e.record(f.number, trace.Created, "", "", ""); err != nil {
		return err
	}
	for _, a := range theAttributes {
		value := a.value(f)
		if value == "" {
			continue
		}
		if err := e.record(f.number, trace.Created, a.named, "", value); err != nil {
			return err
		}
	}
	return nil
}

func whatItHeld(t task.Task) string {
	f := factsOf(t)
	var held []string
	for _, a := range theAttributes {
		value := a.value(f)
		if value == "" {
			continue
		}
		held = append(held, a.named+": "+value)
	}
	return strings.Join(held, "; ")
}

func (e *Engine) recordChanges(was, became task.Task) error {
	before, after := factsOf(was), factsOf(became)
	for _, a := range theAttributes {
		from, to := a.value(before), a.value(after)
		if from == to {
			continue
		}
		if err := e.record(after.number, trace.Changed, a.named, from, to); err != nil {
			return err
		}
	}
	return nil
}

func (e *Engine) record(number int, did, field, was, became string) error {
	var a trace.AsTold
	a.On(number)
	a.Did(did)
	domain.TellOrNot(field, a.Changing, a.NothingNamed)
	domain.TellOrNot(was, a.Was, a.WasNothing)
	domain.TellOrNot(became, a.Became, a.BecameNothing)
	domain.TellOrNot(e.through, a.Through, a.ByHand)
	a.At(time.Now())
	return e.history.Record(a.Act())
}

func yesOrNothing(holds bool) string {
	if holds {
		return "yes"
	}
	return ""
}

func numberOrNothing(number int) string {
	if number == task.NoParent {
		return ""
	}
	return fmt.Sprint(number)
}

func numbersOrNothing(numbers []int) string {
	if len(numbers) == 0 {
		return ""
	}
	written := make([]string, 0, len(numbers))
	for _, one := range numbers {
		written = append(written, domain.Ref(one))
	}
	return strings.Join(written, ", ")
}
