package engine

import (
	"slices"
	"strconv"

	"cairn/internal/domain"
	"cairn/internal/objective"
)

// Standings are the four an objective can be in, in the order a list keeps them: what is being worked towards first, what was finished with last.
var Standings = []string{string(objective.Active), string(objective.OnHold), string(objective.Reached), string(objective.Dropped)}

// Stepped is what moves an objective one place up or down among its siblings, a parent with its sub-objectives; false at the first or the last, since crossing to another parent is a deliberate act.
func Stepped(goals []Goal, n int, up bool) (Input, bool) {
	parent := -1
	for _, g := range goals {
		if g.Number == n {
			parent = g.Parent
		}
	}
	var siblings []int
	for _, g := range goals {
		if g.Parent == parent {
			siblings = append(siblings, g.Number)
		}
	}
	at := slices.Index(siblings, n)
	if at < 0 || (up && at == 0) || (!up && at == len(siblings)-1) {
		return Input{}, false
	}
	given := map[string]string{"number": strconv.Itoa(n)}
	switch {
	case up && at == 1:
		given["first"] = "true"
	case up:
		given["after"] = strconv.Itoa(siblings[at-2])
	default:
		given["after"] = strconv.Itoa(siblings[at+1])
	}
	return Input{Given: given}, true
}

// Sorted puts the objectives in the order a list draws them: the order a person gave them, as the store keeps it.
func Sorted(goals []Goal) []Goal {
	return slices.Clone(goals)
}

// ReachesGoal says whether what was typed reaches an objective, by the same rule a card is reached by.
func ReachesGoal(g Goal, query string) bool {
	return Reaches(g.Said, query)
}

// Counted is how much of an objective is done, the way a board writes it: its tasks, then the steps of its tasks apart, a step being no task.
func (g Goal) Counted() string {
	tasks := strconv.Itoa(g.Done-g.Steps.Done) + "/" + strconv.Itoa(g.Total-g.Steps.Total)
	if g.Steps.Total == 0 {
		return tasks
	}
	return tasks + " · " + strconv.Itoa(g.Steps.Done) + "/" + strconv.Itoa(g.Steps.Total) + " steps"
}

// Percent is the share of it done, by the domain's rule: 0 and 100 are kept for nothing done and everything done.
func (g Goal) Percent() int {
	return domain.Progress{Tasks: domain.Tally{Total: g.Total, Done: g.Done}}.Percent()
}
