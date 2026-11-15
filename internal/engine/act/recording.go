package act

import (
	"cairn/internal/engine/calendar"
	"cairn/internal/engine/item"
	"cairn/internal/engine/task"
	"slices"
	"strings"
)

// TheAttributes is each attribute of a task the trace follows, and how it reads.
var TheAttributes = []struct {
	Named string
	Value func(task.Carried) string
}{
	{"title", func(f task.Carried) string { return f.Title }},
	{"description", func(f task.Carried) string { return f.Description }},
	{"column", func(f task.Carried) string { return string(f.State) }},
	{"priority", func(f task.Carried) string { return string(f.Priority) }},
	{"labels", func(f task.Carried) string { return strings.Join(f.Labels, ", ") }},
	{"blockers", func(f task.Carried) string { return NumbersOrNothing(f.Blockers) }},
	{"links", func(f task.Carried) string { return NumbersOrNothing(f.Linked) }},
	{"because", func(f task.Carried) string { return f.Because }},
	{"deadline", func(f task.Carried) string { return calendar.FormatDay(f.Deadline) }},
	{"estimate", func(f task.Carried) string { return calendar.FormatDuration(f.Estimate) }},
	{"steps", func(f task.Carried) string { return stepsSaid(f.Steps) }},
	{"objectives", func(f task.Carried) string { return NumbersOrNothing(f.Objectives) }},
	{"archived", func(f task.Carried) string { return yesOrNothing(f.Archived) }},
}

// WhatItHeld is what a task carried, as the trace keeps it when the task is deleted.
func WhatItHeld(t task.Task) string {
	f := t.Carried()
	var held []string
	for _, a := range TheAttributes {
		value := a.Value(f)
		if value == "" {
			continue
		}
		held = append(held, a.Named+": "+value)
	}
	return strings.Join(held, "; ")
}

// GainedOrLost is the numbers in one set and not the other.
func GainedOrLost(was, became []int) []int {
	var differing []int
	for _, one := range was {
		if !slices.Contains(became, one) {
			differing = append(differing, one)
		}
	}
	for _, one := range became {
		if !slices.Contains(was, one) {
			differing = append(differing, one)
		}
	}
	return differing
}

// Without is the numbers, less one.
func Without(numbers []int, n int) []int {
	return slices.DeleteFunc(slices.Clone(numbers), func(one int) bool { return one == n })
}

func yesOrNothing(holds bool) string {
	if holds {
		return "yes"
	}
	return ""
}

// NumbersOrNothing is the numbers as the trace writes them, nothing when there are none.
func NumbersOrNothing(numbers []int) string {
	if len(numbers) == 0 {
		return ""
	}
	written := make([]string, 0, len(numbers))
	for _, one := range numbers {
		written = append(written, item.Ref(one))
	}
	return strings.Join(written, ", ")
}

func stepsSaid(steps []task.Step) string {
	said := make([]string, 0, len(steps))
	for _, s := range steps {
		mark := " "
		if s.Done {
			mark = "x"
		}
		said = append(said, "["+mark+"] "+s.Title)
	}
	return strings.Join(said, "; ")
}
