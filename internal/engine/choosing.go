package engine

import (
	"cairn/internal/engine/filter"
	"cairn/internal/engine/item"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"cairn/internal/engine/objective"
	"cairn/internal/engine/task"
)

var blockerArgument = Argument{Name: "blocker", Kind: Number, Of: "a task", Means: "keep only the tasks this task is in the way of"}

func (e *Engine) choosing(shown []task.Task) (map[int]task.Choosing, error) {
	board, err := e.narrowed(&filter.Narrowing{})
	if err != nil {
		return nil, err
	}
	in := map[int][]int{}
	for _, t := range board {
		if f := t.Carried(); f.State != item.Done {
			for _, b := range f.Blockers {
				in[b] = append(in[b], f.Number)
			}
		}
	}
	acts, err := e.since(time.Time{})
	if err != nil {
		return nil, err
	}
	moved := map[int]time.Time{}
	for _, a := range acts {
		if c := a.Carried(); c.Field == "column" {
			moved[c.Task] = c.At
		}
	}
	goals, err := e.objectives.AllObjectives()
	if err != nil {
		return nil, err
	}
	today := e.today()
	late := map[int]bool{}
	for _, o := range goals {
		c := o.Carried()
		late[c.Number] = !c.Deadline.IsZero() && c.Deadline.Before(today) && (c.State == objective.Active || c.State == objective.OnHold)
	}
	out := map[int]task.Choosing{}
	for _, t := range shown {
		f := t.Carried()
		said := task.Choosing{Blocks: slices.Sorted(slices.Values(in[f.Number])), Age: daysFrom(moved[f.Number], today)}
		if !f.Deadline.IsZero() && f.State != item.Done {
			days := daysFrom(today, f.Deadline)
			if days < 0 {
				days = -days
				said.Overdue = &days
			} else {
				said.DueIn = &days
			}
		}
		said.ObjectiveLate = slices.ContainsFunc(f.Objectives, func(n int) bool { return late[n] })
		out[f.Number] = said
	}
	return out, nil
}

func askedBlocker(in Input) (int, bool, error) {
	said, named := in.Given[blockerArgument.Name]
	if !named {
		return 0, false, nil
	}
	n, err := strconv.Atoi(strings.TrimSpace(said))
	if err != nil || n <= 0 {
		return 0, false, fmt.Errorf("%q: %w", said, task.ErrBlockerRequired)
	}
	return n, true, nil
}
