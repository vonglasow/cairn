package engine

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"cairn/internal/domain"
	"cairn/internal/objective"
	"cairn/internal/task"
)

// Choosing is what a task says to help choose it, beside what it carries: the open tasks it is in the way of, how far it stands from its deadline, how many days it has sat in its column, and whether an objective it serves is past its deadline unreached.
type Choosing struct {
	Blocks        []int
	Overdue       *int
	DueIn         *int
	Age           int
	ObjectiveLate bool
}

var blockerArgument = Argument{Name: "blocker", Kind: Number, Of: "a task", Means: "keep only the tasks this task is in the way of"}

func (e *Engine) choosing(shown []task.Task) (map[int]Choosing, error) {
	board, err := e.narrowed(&domain.Narrowing{})
	if err != nil {
		return nil, err
	}
	in := map[int][]int{}
	for _, t := range board {
		if f := factsOf(t); f.State != domain.Done {
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
	out := map[int]Choosing{}
	for _, t := range shown {
		f := factsOf(t)
		said := Choosing{Blocks: slices.Sorted(slices.Values(in[f.Number])), Age: daysFrom(moved[f.Number], today)}
		if !f.Deadline.IsZero() && f.State != domain.Done {
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
		return 0, false, fmt.Errorf("%q: %w", said, ErrBlockerRequired)
	}
	return n, true, nil
}

func inTheWayOf(tasks []task.Task, blocker int) []task.Task {
	return slices.DeleteFunc(tasks, func(t task.Task) bool { return !slices.Contains(factsOf(t).Blockers, blocker) })
}

func blocksOf(chosen map[int]Choosing) map[int]int {
	out := map[int]int{}
	for n, c := range chosen {
		out[n] = len(c.Blocks)
	}
	return out
}
