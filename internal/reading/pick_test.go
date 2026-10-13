package reading

import (
	"slices"
	"testing"
)

// numbersOfCards lists the numbers of cards, in order.
func numbersOfCards(cards []Card) []int {
	out := make([]int, 0, len(cards))
	for _, c := range cards {
		out = append(out, c.Number)
	}
	return out
}

// What a search reaches among the objectives, in the order a list keeps them —
// one rule for a board's search box and for the window's list of objectives to
// serve, which matched on the title alone while the box beside it matched on
// everything.
func TestFoundGoals_isWhatASearchReachesInTheOrderAListKeeps(t *testing.T) {
	goals := []Goal{
		{Number: 1, Title: "Ship it", State: "reached"},
		{Number: 2, Title: "Write it", State: "active", Labels: []string{"docs"}, Said: []string{"Write it", "docs"}},
		{Number: 3, Title: "Plan it", State: "active"},
	}
	var got []int
	for _, g := range FoundGoals(goals, "") {
		got = append(got, g.Number)
	}
	if !slices.Equal(got, []int{2, 3, 1}) {
		t.Errorf("nothing looked for lists %v, want every one, active first", got)
	}
	got = nil
	for _, g := range FoundGoals(goals, "docs") {
		got = append(got, g.Number)
	}
	if !slices.Equal(got, []int{2}) {
		t.Errorf("looking for a label found %v, want 2", got)
	}
}

// What a task may hang under: work on the board, never itself and never a
// part, in the order somebody reaches for it — what is being worked on first.
// The engine refuses the rest; offering them was offering refusals.
func TestParentable_isWhatATaskMayHangUnder(t *testing.T) {
	cards := []Card{
		{Number: 1, Title: "Old", State: "backlog"},
		{Number: 2, Title: "Now", State: "current"},
		{Number: 3, Title: "A part", State: "todo", Parent: 2},
		{Number: 4, Title: "Itself", State: "todo"},
		{Number: 5, Title: "Next", State: "todo", Said: []string{"Next"}},
		{Number: 6, Title: "Over", State: "done"},
		{Number: 7, Title: "Stuck", State: "blocked"},
	}
	if got := numbersOfCards(Parentable(cards, 4, "")); !slices.Equal(got, []int{2, 5, 7, 1, 6}) {
		t.Errorf("Parentable = %v", got)
	}
	if got := numbersOfCards(Parentable(cards, 4, "next")); !slices.Equal(got, []int{5}) {
		t.Errorf("Parentable looking for next = %v", got)
	}
}

// Whether a task has gone past what it was reckoned to take is read off the
// lengths themselves, not off how a card happens to write them.
func TestOverrun_isSpentPastTheEstimate(t *testing.T) {
	for _, tc := range []struct {
		estimate, spent string
		over            bool
	}{
		{"1h", "1h30m", true},
		{"1h", "45m", false},
		{"1h", "1h", false},
		{"", "3h", false},
		{"1h", "", false},
	} {
		if got := overrun(tc.estimate, tc.spent); got != tc.over {
			t.Errorf("%q against %q: over = %v, want %v", tc.spent, tc.estimate, got, tc.over)
		}
	}
}
