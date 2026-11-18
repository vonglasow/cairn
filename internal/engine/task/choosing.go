package task

import (
	"slices"
)

// Choosing is what a task says to help choose it, beside what it carries: the open tasks it is in the way of, how far it stands from its deadline, how many days it has sat in its column, and whether an objective it serves is past its deadline unreached.
type Choosing struct {
	Blocks        []int
	Overdue       *int
	DueIn         *int
	Age           int
	ObjectiveLate bool
	Serves        []string
}

// InTheWayOf is the tasks a blocker is in the way of.
func InTheWayOf(tasks []Task, blocker int) []Task {
	return slices.DeleteFunc(tasks, func(t Task) bool { return !slices.Contains(t.Carried().Blockers, blocker) })
}

// BlocksOf is how many open tasks each task chosen is in the way of.
func BlocksOf(chosen map[int]Choosing) map[int]int {
	out := map[int]int{}
	for n, c := range chosen {
		out[n] = len(c.Blocks)
	}
	return out
}

// ServesOf is the objectives each task chosen serves, each under its parent.
func ServesOf(chosen map[int]Choosing) map[int][]string {
	out := map[int][]string{}
	for n, c := range chosen {
		out[n] = c.Serves
	}
	return out
}
