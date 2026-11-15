package board

import (
	"cairn/internal/engine/calendar"
	"cmp"
	"slices"
)

// TasksNoneSays is what a list of tasks reads when no other task reaches what is typed.
const TasksNoneSays = "no other task"

// FoundCards is what a search reaches among the tasks but one, the newest first; every one of them when nothing is looked for.
func FoundCards(cards []Card, query string, except int) []Card {
	var found []Card
	for _, c := range cards {
		if c.Number != except && (query == "" || ReachesCard(c, query)) {
			found = append(found, c)
		}
	}
	slices.SortFunc(found, func(a, b Card) int { return cmp.Compare(b.Number, a.Number) })
	return found
}

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
	reckoned, _ := calendar.ParseEstimate(estimate)
	took, _ := calendar.ParseEstimate(spent)
	return reckoned > 0 && took > reckoned
}
