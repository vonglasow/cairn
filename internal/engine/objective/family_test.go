package objective

import (
	"cairn/internal/engine/item"
	"strings"
	"testing"
	"time"
)

func named(number int, title string, under int) Objective {
	var a AsTold
	a.Numbered(number)
	a.Titled(title)
	a.Standing(Active)
	a.Under(under)
	return a.Objective()
}

func TestChildren_listsEachSubObjectiveWithItsCounts(t *testing.T) {
	got := Children(Family{Children: []Child{
		{Objective: named(5, "Enclosure", 4), Progress: item.Progress{Tasks: item.Tally{Total: 6, Done: 3}}, Spent: 8*time.Hour + 25*time.Minute},
		{Objective: named(7, "Features", 4)},
	}})
	want := "\nsub-objectives\n  #5  Enclosure  · active  · 3 of 6 tasks  · 8h25m\n  #7  Features  · active\n"
	if got != want {
		t.Errorf("Children =\n%q\nwant\n%q", got, want)
	}
	if Children(Family{}) != "" {
		t.Error("an objective with no sub-objective names none")
	}
}

func TestProposal_saysAParentMayBeReached(t *testing.T) {
	if got := Proposal(named(4, "Home automation", 0), true); got != "\nevery sub-objective is reached: it may be put reached too\n" {
		t.Errorf("Proposal on itself = %q", got)
	}
	if got := Proposal(named(4, "Home automation", 0), false); got != "\nevery sub-objective of #4 Home automation is reached: it may be put reached too\n" {
		t.Errorf("Proposal from a sub-objective = %q", got)
	}
}

func TestApart_splitsTheTasksOnItFromThoseInItsSubObjectives(t *testing.T) {
	got := Apart([]string{"#42  Wiring"}, []string{"#33  Print the case"})
	want := "\ntasks on it directly\n  #42  Wiring\n\ntasks in its sub-objectives\n  #33  Print the case\n"
	if got != want {
		t.Errorf("Apart =\n%q\nwant\n%q", got, want)
	}
	none := Apart(nil, nil)
	if none != "\nnothing is on it directly\n\nnothing is in its sub-objectives\n" {
		t.Errorf("Apart(nil, nil) = %q", none)
	}
}

func TestList_drawsTheTreeOnTwoLinesEach(t *testing.T) {
	got := List([]Objective{
		named(1, "Tidy", 0), named(4, "Home automation", 0), named(5, "Enclosure", 4), named(6, "Firmware", 4),
	}, map[int]item.Progress{4: {Tasks: item.Tally{Total: 4, Done: 1}}, 5: {Tasks: item.Tally{Total: 2, Done: 1}}},
		map[int]time.Duration{4: 90 * time.Minute})
	want := "#1  Tidy\n" +
		"    active\n" +
		"#4  Home automation\n" +
		"    active  · 1 of 4 tasks  · 25% done  · 1h30m\n" +
		"  ├ #5  Enclosure\n" +
		"  │   active  · 1 of 2 tasks  · 50% done\n" +
		"  └ #6  Firmware\n" +
		"      active\n"
	if got != want {
		t.Errorf("List =\n%s\nwant\n%s", got, want)
	}
}

func TestList_saysWhereASubObjectiveSitsWhenItsParentIsNotListed(t *testing.T) {
	got := List([]Objective{named(5, "Enclosure", 4)}, nil, nil)
	if got != "#5  Enclosure\n    active  · under objective 4\n" {
		t.Errorf("List = %q", got)
	}
}

func TestReport_splitsAMonthBySubObjective(t *testing.T) {
	in := Period{
		Since: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC), Until: time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC),
		Months: []Month{{Month: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC), Spent: 150 * time.Minute,
			Split: []Part{{Number: 2, Spent: 2 * time.Hour}, {Number: 3}, {Spent: 30 * time.Minute}}}},
	}
	got := Report(plan(t), item.Progress{}, Review{}, in)
	if !strings.Contains(got, "\n  2026-09  2h30m  · #2 2h  · #3 —  · itself 30m\n") {
		t.Errorf("the report does not split the month:\n%s", got)
	}
}
