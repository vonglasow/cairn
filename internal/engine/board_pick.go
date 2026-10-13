package engine

import (
	"slices"

	"cairn/internal/domain"
)

// FoundGoals is what a search reaches among the objectives, in the order a list
// keeps them; every one of them when nothing is looked for.
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

// Parentable is what the task numbered self may hang under, among what a
// search reaches: work on the board, never the task itself, and never a part —
// nothing hangs under a part. In the order somebody reaches for them, which is
// not the order the board reads: work is hung under what is being worked on,
// and the board's order put two dozen backlog tasks in front of it.
func Parentable(cards []Card, self int, query string) []Card {
	var out []Card
	for _, c := range OnTheBoard(cards, false) {
		if c.Number == self || c.Parent != 0 {
			continue
		}
		if query == "" || ReachesCard(c, query) {
			out = append(out, c)
		}
	}
	slices.SortStableFunc(out, func(a, b Card) int { return reachedFor(a.State) - reachedFor(b.State) })
	return out
}

// reachedFor is how soon a column is reached for when something is being hung
// under it: what is being worked on, then what is next, then what is stopped,
// then what has not been started, and last what is over.
func reachedFor(state string) int {
	switch domain.State(state) {
	case domain.Current:
		return 0
	case domain.Todo:
		return 1
	case domain.Blocked, domain.Watching, domain.Paused:
		return 2
	case domain.Backlog:
		return 3
	}
	return 4
}

// overrun says whether the time spent has gone past the estimate, read off the
// lengths themselves. Nothing estimated is never gone past.
func overrun(estimate, spent string) bool {
	reckoned, oneErr := domain.ParseEstimate(estimate)
	took, twoErr := domain.ParseEstimate(spent)
	return oneErr == nil && twoErr == nil && reckoned > 0 && took > reckoned
}
