package board

import (
	"slices"
	"testing"
)

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
	if !slices.Equal(got, []int{1, 2, 3}) {
		t.Errorf("nothing looked for lists %v, want every one, in the order given", got)
	}
	got = nil
	for _, g := range FoundGoals(goals, "docs") {
		got = append(got, g.Number)
	}
	if !slices.Equal(got, []int{2}) {
		t.Errorf("looking for a label found %v, want 2", got)
	}
}
