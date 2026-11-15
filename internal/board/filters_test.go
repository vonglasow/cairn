package board

import (
	"slices"
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

func TestFilters_keepTheCardsEveryFilterKeeps(t *testing.T) {
	cards := filterCards()
	fs := FiltersOf(Snapshot{Cards: cards, Goals: []Goal{{Number: 2}}}, time.Date(2026, 10, 24, 0, 0, 0, 0, time.UTC))
	for _, c := range []struct {
		name, value, want string
	}{
		{"column", "todo", "1 2"},
		{"priority", "high", "1 3"},
		{"label", "sprint", "1 3"},
		{"objective", "2", "1"},
		{"deadline", "2026-10-31", "1"},
	} {
		kept, err := on(FiltersOf(Snapshot{Cards: cards, Goals: []Goal{{Number: 2}}}, time.Date(2026, 10, 24, 0, 0, 0, 0, time.UTC)), c.name, c.value).Kept(cards)
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
	fs := FiltersOf(Snapshot{Cards: filterCards()}, time.Date(2026, 10, 24, 0, 0, 0, 0, time.UTC))
	walked := fs.Walked(1, -1)
	if got := walked[1].value(); got != "critical" {
		t.Fatalf("walking back from any reaches %q", got)
	}
	if fs[1].value() != anyValue {
		t.Fatal("walking changed the filters it was asked of")
	}
	if !walked.Narrows() || walked.Said() != "priority critical" {
		t.Fatalf("walked says %q", walked.Said())
	}
	if walked.Walked(1, 1)[1].value() != anyValue {
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

func labelFilterAt(fs Filters) int {
	for i, f := range fs {
		if f.Arg == "label" {
			return i
		}
	}
	return -1
}

func TestFilters_takeSeveralLabelsTogether(t *testing.T) {
	fs := FiltersOf(Snapshot{Cards: filterCards()}, time.Date(2026, 10, 24, 0, 0, 0, 0, time.UTC))
	at := labelFilterAt(fs)
	if !fs[at].Many {
		t.Fatal("the label filter takes one label only")
	}
	fs = fs.Toggled(at, "backend").Toggled(at, "sprint")
	if got := strings.Join(fs[at].Taken(), ","); got != "backend,sprint" {
		t.Fatalf("the label filter took %s", got)
	}
	kept, err := fs.Kept(filterCards())
	if err != nil || filteredNumbers(kept) != "1" {
		t.Fatalf("both labels kept %s (%v)", filteredNumbers(kept), err)
	}
	if got := fs.Said(); got != "label backend + sprint" {
		t.Fatalf("the filters say %q", got)
	}
	fs = fs.Toggled(at, "backend")
	if got := strings.Join(fs[at].Taken(), ","); got != "sprint" {
		t.Fatalf("toggled off, the label filter took %s", got)
	}
	if fs = fs.Toggled(at, "sprint"); fs.Narrows() {
		t.Fatal("every label let go of still narrows")
	}
	if fs = fs.Toggled(at, "backend").Cleared(); fs.Narrows() || len(fs[at].Taken()) > 0 {
		t.Fatal("cleared, the labels taken stay")
	}
	walked := fs.Toggled(at, "backend").Toggled(at, "sprint").Walked(at, 1)
	if got := strings.Join(walked[at].Taken(), ","); got != "backend" {
		t.Fatalf("walked, the label filter took %s", got)
	}
	held := FiltersOf(Snapshot{Cards: filterCards()[1:2]}, time.Date(2026, 10, 24, 0, 0, 0, 0, time.UTC)).Keeping(fs.Toggled(at, "backend").Toggled(at, "sprint"))
	if got := strings.Join(held[at].Taken(), ","); got != "backend" {
		t.Fatalf("kept over a board where sprint is carried no more, the labels taken are %s", got)
	}
}

func TestFilters_countWhatEachValueKeepsUnderTheOthers(t *testing.T) {
	fs := on(FiltersOf(Snapshot{Cards: filterCards()}, time.Date(2026, 10, 24, 0, 0, 0, 0, time.UTC)), "priority", "high")
	at := labelFilterAt(fs)
	counts, err := fs.Counts(filterCards(), at)
	if err != nil {
		t.Fatal(err)
	}
	if got := Numbers(counts); got != "2 1 2" {
		t.Fatalf("under priority high, any, backend and sprint keep %s", got)
	}
}

func TestFilters_narrowTheObjectivesByWhatTheyCarry(t *testing.T) {
	goals := []Goal{
		{Number: 1, Title: "Ship", Priority: "high", Labels: []string{"backend"}},
		{Number: 2, Title: "Plan", Priority: "normal", Labels: []string{"backend", "sprint"}},
		{Number: 3, Title: "Rest", Priority: "high"},
	}
	fs := FiltersOf(Snapshot{Cards: filterCards(), Goals: goals}, time.Date(2026, 10, 24, 0, 0, 0, 0, time.UTC))
	at := labelFilterAt(fs)
	if got := goalNumbers(fs.Toggled(at, "backend").KeptGoals(goals)); got != "1 2" {
		t.Fatalf("label backend keeps the objectives %s", got)
	}
	if got := goalNumbers(on(slices.Clone(fs), "priority", "high").KeptGoals(goals)); got != "1 3" {
		t.Fatalf("priority high keeps the objectives %s", got)
	}
	if got := goalNumbers(on(slices.Clone(fs), "objective", "2").KeptGoals(goals)); got != "2" {
		t.Fatalf("objective 2 keeps the objectives %s", got)
	}
	if got := goalNumbers(on(slices.Clone(fs), "column", "todo").KeptGoals(goals)); got != "1 2 3" {
		t.Fatalf("a column, which an objective has not, keeps the objectives %s", got)
	}
	for _, f := range fs {
		if reaches := f.ReachesGoals(); reaches != (f.Arg == "label" || f.Arg == "priority" || f.Arg == "objective") {
			t.Errorf("%s reaches the objectives: %v", f.Name, reaches)
		}
	}
}

func goalNumbers(goals []Goal) string {
	var out []int
	for _, g := range goals {
		out = append(out, g.Number)
	}
	return Numbers(out)
}

func TestBoardFilters_keepWhatIsInFocusOrWhatIsNot(t *testing.T) {
	cards := []Card{{Number: 1}, {Number: 2, InFocus: true}}
	fs := FiltersOf(Snapshot{Cards: cards}, time.Now())
	if last := fs[len(fs)-1]; last.Arg != "focus" || !slices.Equal(last.Shown, []string{anyValue, "in focus", "not in focus"}) {
		t.Fatalf("the last filter is %+v, want focus: in focus, not in focus", last)
	}
	for value, want := range map[string]int{"in": 2, "out": 1} {
		kept, err := fs.On("focus", value).Kept(cards)
		if err != nil || len(kept) != 1 || kept[0].Number != want {
			t.Errorf("focus %s kept %+v (%v), want task %d", value, kept, err, want)
		}
	}
	if said := fs.On("focus", "in").Said(); said != "focus in focus" {
		t.Errorf("the filters say %q", said)
	}
}

func TestAFilterSaysTheValuesItTakesAsTheyRead(t *testing.T) {
	f := Filter{Values: []string{"api", "doc"}, Shown: []string{"API", "Doc"}, Ticked: []string{"api", "doc", "gone"}}
	if got := f.TakenSays(); got != "API + Doc + gone" {
		t.Errorf("TakenSays = %q", got)
	}
}

func TestFilters_keepOneValueHeldAsTheOneTheyAreOn(t *testing.T) {
	fs := Filters{several(offered("label", "label", []string{"api", "doc"}, nil))}
	one := fs.Keeping(Filters{{Arg: "label", Ticked: []string{"doc", "gone"}}})
	if one[0].At != 2 || one[0].Ticked != nil {
		t.Errorf("one value held: at %d, ticked %v", one[0].At, one[0].Ticked)
	}
	two := fs.Keeping(Filters{{Arg: "label", Ticked: []string{"doc", "api"}}})
	if two[0].At != 0 || !slices.Equal(two[0].Ticked, []string{"doc", "api"}) {
		t.Errorf("two values held: at %d, ticked %v", two[0].At, two[0].Ticked)
	}
	none := fs.Keeping(Filters{{Arg: "label", Ticked: []string{"gone"}}})
	if none[0].At != 0 || none[0].Ticked != nil {
		t.Errorf("nothing held still offered: at %d, ticked %v", none[0].At, none[0].Ticked)
	}
}
