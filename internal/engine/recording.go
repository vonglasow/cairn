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
	value func(task.Carried) string
}{
	{"title", func(f task.Carried) string { return f.Title }},
	{"description", func(f task.Carried) string { return f.Description }},
	{"column", func(f task.Carried) string { return string(f.State) }},
	{"priority", func(f task.Carried) string { return string(f.Priority) }},
	{"labels", func(f task.Carried) string { return strings.Join(f.Labels, ", ") }},
	{"blockers", func(f task.Carried) string { return numbersOrNothing(f.Blockers) }},
	{"because", func(f task.Carried) string { return f.Because }},
	{"deadline", func(f task.Carried) string { return domain.FormatDay(f.Deadline) }},
	{"estimate", func(f task.Carried) string { return domain.FormatDuration(f.Estimate) }},
	{"parent", func(f task.Carried) string { return numberOrNothing(f.Parent) }},
	{"objectives", func(f task.Carried) string { return numbersOrNothing(f.Objectives) }},
	{"archived", func(f task.Carried) string { return yesOrNothing(f.Archived) }},
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
	if err := e.record(f.Number, trace.Created, "", "", ""); err != nil {
		return err
	}
	for _, a := range theAttributes {
		value := a.value(f)
		if value == "" {
			continue
		}
		if err := e.record(f.Number, trace.Created, a.named, "", value); err != nil {
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
		if err := e.record(after.Number, trace.Changed, a.named, from, to); err != nil {
			return err
		}
	}
	if after.State == domain.Done && before.State != domain.Done {
		return e.letGoOf(after.Number)
	}
	return nil
}

func (e *Engine) record(number int, did, field, was, became string) error {
	return e.history.Record(trace.Made(trace.Carried{
		Task: number, Did: did, Field: field, Was: was, Became: became, Through: e.through, At: time.Now(),
	}))
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
