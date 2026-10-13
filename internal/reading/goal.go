package reading

// What an objective says about itself, and the order a list of them goes in.
// None of it is about drawing: the standing an objective is in, where that
// standing sorts, and whether a search reaches one are the same answers on a
// terminal and in a window.

import (
	"cmp"
	"slices"
	"strconv"

	"cairn/internal/domain"
	"cairn/internal/engine"
	"cairn/internal/objective"
)

// Standings are the four an objective can be in, in the order a list keeps
// them: what is being worked towards first, what was finished with last.
var Standings = []string{string(objective.Active), string(objective.OnHold), string(objective.Reached), string(objective.Dropped)}

// standingAt answers where a state stands in the order. Anything unrecognised
// sorts last rather than first: a word nobody planned for should not push what
// is being worked on off the top of the screen.
func standingAt(state string) int {
	for i, s := range Standings {
		if s == state {
			return i
		}
	}
	return len(Standings)
}

// Sorted puts the objectives in the order a list draws them. There are no
// headings, so this order is the only thing saying what is live and what is
// over — and inside one standing the numbers keep their own order, which is
// the order they were made in.
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

// ReachesGoal says whether what was typed reaches an objective, by the same
// rule a card is reached by.
func ReachesGoal(g Goal, query string) bool {
	return engine.Reaches(g.Said, query)
}

// Counted is how much of an objective is done, the way a board writes it.
func (g Goal) Counted() string { return strconv.Itoa(g.Done) + "/" + strconv.Itoa(g.Total) }

// Percent is the share of it done, by the domain's rule: 0 and 100 are kept for
// nothing done and everything done.
func (g Goal) Percent() int {
	return domain.Progress{Tasks: domain.Tally{Total: g.Total, Done: g.Done}}.Percent()
}
