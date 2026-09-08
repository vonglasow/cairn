package reading

// What an objective says about itself, and the order a list of them goes in.
// None of it is about drawing: the standing an objective is in, where that
// standing sorts, and whether a search reaches one are the same answers on a
// terminal and in a window.

import (
	"cmp"
	"slices"
	"strings"
)

// Standings are the four an objective can be in, in the order a list keeps
// them: what is being worked towards first, what was finished with last.
var Standings = []string{"active", "on-hold", "reached", "dropped"}

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

// Reaches says whether a search reaches an objective, looking where somebody
// typed something and nowhere else.
func Reaches(g Goal, query string) bool {
	query = strings.ToLower(query)
	return strings.Contains(strings.ToLower(g.Title), query) ||
		strings.Contains(strings.ToLower(g.Description), query)
}
