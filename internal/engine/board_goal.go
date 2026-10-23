package engine

import (
	"cmp"
	"slices"
	"strconv"

	"cairn/internal/domain"
	"cairn/internal/objective"
)

// Standings are the four an objective can be in, in the order a list keeps them: what is being worked towards first, what was finished with last.
var Standings = []string{string(objective.Active), string(objective.OnHold), string(objective.Reached), string(objective.Dropped)}

func standingAt(state string) int {
	for i, s := range Standings {
		if s == state {
			return i
		}
	}
	return len(Standings)
}

// Sorted puts the objectives in the order a list draws them.
func Sorted(goals []Goal) []Goal {
	out := make([]Goal, len(goals))
	copy(out, goals)
	slices.SortStableFunc(out, func(a, b Goal) int {
		if by := cmp.Compare(standingAt(a.State), standingAt(b.State)); by != 0 {
			return by
		}
		return cmp.Compare(a.Number, b.Number)
	})
	return out
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
