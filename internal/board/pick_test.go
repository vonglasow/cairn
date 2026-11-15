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

func TestFoundCards_isWhatASearchReachesAmongTheOtherTasksNewestFirst(t *testing.T) {
	cards := []Card{
		{Number: 1, Title: "Fix the login", Said: []string{"Fix the login"}},
		{Number: 2, Title: "Write the docs", Said: []string{"Write the docs", "login"}},
		{Number: 3, Title: "Plan it", Said: []string{"Plan it"}},
	}
	for _, c := range []struct {
		query  string
		except int
		want   []int
	}{
		{"", 0, []int{3, 2, 1}},
		{"", 3, []int{2, 1}},
		{"login", 0, []int{2, 1}},
		{"login", 2, []int{1}},
		{"nothing", 0, nil},
	} {
		var got []int
		for _, card := range FoundCards(cards, c.query, c.except) {
			got = append(got, card.Number)
		}
		if !slices.Equal(got, c.want) {
			t.Errorf("FoundCards(%q, except %d) = %v, want %v", c.query, c.except, got, c.want)
		}
	}
}
