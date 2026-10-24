package engine

import (
	"strings"
	"testing"
	"time"
)

func filterCards() []Card {
	day := func(s string) time.Time { d, _ := time.Parse("2006-01-02", s); return d }
	return []Card{
		{Number: 1, Title: "spec", State: "todo", Priority: "high", Labels: []string{"backend", "sprint"}, Objectives: []int{2}, Deadline: day("2026-10-25")},
		{Number: 2, Title: "review", State: "todo", Priority: "normal", Labels: []string{"backend"}},
		{Number: 3, Title: "ship", State: "current", Priority: "high", Labels: []string{"sprint"}, Deadline: day("2026-11-30")},
	}
}

func filteredNumbers(cards []Card) string {
	var out []string
	for _, c := range cards {
		out = append(out, Numbers([]int{c.Number}))
	}
	return strings.Join(out, " ")
}

func on(fs Filters, name, value string) Filters {
	for i := range fs {
		if fs[i].Name != name {
			continue
		}
		for at, v := range fs[i].Values {
			if v == value {
				fs[i].At = at
			}
		}
	}
	return fs
}

func TestFilters_offerEveryNarrowingOfAListingStartingAtAny(t *testing.T) {
	fs := BoardFilters(Snapshot{Cards: filterCards(), Goals: []Goal{{Number: 2, Title: "Ship 1.0"}}}, time.Date(2026, 10, 24, 0, 0, 0, 0, time.UTC))
	var names []string
	for _, f := range fs {
		names = append(names, f.Name)
		if f.Value() != Any {
			t.Fatalf("%s opens on %q, not %q", f.Name, f.Value(), Any)
		}
	}
	if got := strings.Join(names, ","); got != "column,priority,label,objective,due by" {
		t.Fatalf("the filters are %s", got)
	}
	if fs.Narrows() {
		t.Fatal("filters all on any narrow something")
	}
	if got := strings.Join(fs[2].Values, ","); got != "any,backend,sprint" {
		t.Fatalf("the labels offered are %s", got)
	}
	if got := strings.Join(fs[3].Shown, ","); got != "any,#2 Ship 1.0" {
		t.Fatalf("the objectives offered read %s", got)
	}
	if got := strings.Join(fs[4].Values, ","); got != "any,2026-10-24,2026-10-31,2026-11-23" {
		t.Fatalf("the days offered are %s", got)
	}
}

func TestFilters_keepTheCardsEveryFilterKeeps(t *testing.T) {
	cards := filterCards()
	fs := BoardFilters(Snapshot{Cards: cards, Goals: []Goal{{Number: 2}}}, time.Date(2026, 10, 24, 0, 0, 0, 0, time.UTC))
	for _, c := range []struct {
		name, value, want string
	}{
		{"column", "todo", "1 2"},
		{"priority", "high", "1 3"},
		{"label", "sprint", "1 3"},
		{"objective", "2", "1"},
		{"due by", "2026-10-31", "1"},
	} {
		kept, err := on(BoardFilters(Snapshot{Cards: cards, Goals: []Goal{{Number: 2}}}, time.Date(2026, 10, 24, 0, 0, 0, 0, time.UTC)), c.name, c.value).Kept(cards)
		if err != nil {
			t.Fatal(err)
		}
		if got := filteredNumbers(kept); got != c.want {
			t.Errorf("%s %s keeps %q, want %q", c.name, c.value, got, c.want)
		}
	}
	both, err := on(on(fs, "column", "todo"), "priority", "high").Kept(cards)
	if err != nil {
		t.Fatal(err)
	}
	if got := filteredNumbers(both); got != "1" {
		t.Fatalf("todo and high together keep %q", got)
	}
}

func TestFilters_walkWrapsAndClearingPutsEveryOneBackOnAny(t *testing.T) {
	fs := BoardFilters(Snapshot{Cards: filterCards()}, time.Date(2026, 10, 24, 0, 0, 0, 0, time.UTC))
	walked := fs.Walked(1, -1)
	if got := walked[1].Value(); got != "critical" {
		t.Fatalf("walking back from any reaches %q", got)
	}
	if fs[1].Value() != Any {
		t.Fatal("walking changed the filters it was asked of")
	}
	if !walked.Narrows() || walked.Said() != "priority critical" {
		t.Fatalf("walked says %q", walked.Said())
	}
	if walked.Walked(1, 1)[1].Value() != Any {
		t.Fatal("walking on from the last value does not come back to any")
	}
	two := on(walked, "label", "backend")
	if two.Said() != "priority critical · label backend" {
		t.Fatalf("two filters say %q", two.Said())
	}
	if two.Cleared().Narrows() {
		t.Fatal("cleared filters still narrow")
	}
}

func TestFilters_keepWhatTheyHeldWhenTheBoardIsReadAgain(t *testing.T) {
	today := time.Date(2026, 10, 24, 0, 0, 0, 0, time.UTC)
	held := on(BoardFilters(Snapshot{Cards: filterCards()}, today), "label", "sprint")
	again := BoardFilters(Snapshot{Cards: filterCards()[:2]}, today).Keeping(held)
	if again[2].Value() != "sprint" {
		t.Fatalf("a label still held by a card was let go: %q", again[2].Value())
	}
	gone := BoardFilters(Snapshot{Cards: filterCards()[1:2]}, today).Keeping(held)
	if gone[2].Value() != Any {
		t.Fatalf("a label no card holds any more is still kept: %q", gone[2].Value())
	}
}
