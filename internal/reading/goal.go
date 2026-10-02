package reading

// What an objective says about itself, and the order a list of them goes in.
// None of it is about drawing: the standing an objective is in, where that
// standing sorts, and whether a search reaches one are the same answers on a
// terminal and in a window.

import (
	"cmp"
	"slices"
	"strconv"
	"strings"

	"cairn/internal/domain"
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

// ReachesGoal says whether a search reaches an objective, over everything it
// carries and with every word that was typed — the same search a card gets, so
// that one word does not mean two things a tab apart.
func ReachesGoal(g Goal, query string) bool {
	if query == "" {
		return false
	}
	said := []string{
		strconv.Itoa(g.Number),
		g.Title,
		g.Description,
		g.State,
		g.Priority,
		g.Estimate,
	}
	said = append(said, g.Labels...)
	if !g.Deadline.IsZero() {
		said = append(said, domain.FormatDay(g.Deadline))
	}
	for _, word := range strings.Fields(strings.ToLower(query)) {
		if !carries(said, word) {
			return false
		}
	}
	return true
}
