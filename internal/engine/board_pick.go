package engine

import (
	"cairn/internal/domain"
)

// FoundGoals is what a search reaches among the objectives, in the order a list keeps them; every one of them when nothing is looked for.
func FoundGoals(goals []Goal, query string) []Goal {
	if query == "" {
		return Sorted(goals)
	}
	var found []Goal
	for _, g := range goals {
		if ReachesGoal(g, query) {
			found = append(found, g)
		}
	}
	return Sorted(found)
}

func overrun(estimate, spent string) bool {
	reckoned, oneErr := domain.ParseEstimate(estimate)
	took, twoErr := domain.ParseEstimate(spent)
	return oneErr == nil && twoErr == nil && reckoned > 0 && took > reckoned
}
