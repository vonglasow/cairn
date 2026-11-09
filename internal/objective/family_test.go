package objective

import (
	"errors"
	"strings"
	"testing"
	"time"

	"cairn/internal/domain"
)

func named(number int, title string, under int) Objective {
	var a AsTold
	a.Numbered(number)
	a.Titled(title)
	a.Standing(Active)
	a.Under(under)
	return a.Objective()
}

func TestAsTold_isToldWhatItSitsUnder(t *testing.T) {
	if got := named(5, "Enclosure", 4).Carried().Parent; got != 4 {
		t.Errorf("Parent = %d, want 4", got)
	}
	if got := Made(Carried{Number: 5, Parent: 4}).Carried().Parent; got != 4 {
		t.Errorf("Made lost the parent: %d", got)
	}
}

func TestMayHang_takesALevelOneParent(t *testing.T) {
	if err := MayHang(named(5, "Enclosure", 0), named(4, "Home automation", 0), false); err != nil {
		t.Errorf("MayHang: %v, want it allowed", err)
	}
}

func TestMayHang_refusesAThirdLevel(t *testing.T) {
	err := MayHang(named(7, "Bootloader", 0), named(6, "Firmware", 4), false)
	if !errors.Is(err, ErrTooDeep) || !strings.Contains(err.Error(), "#6 Firmware already sits under #4") {
		t.Errorf("MayHang = %v, want #6 refused as already under #4", err)
	}
}

func TestMayHang_refusesAParentAParent(t *testing.T) {
	err := MayHang(named(4, "Home automation", 0), named(9, "Garden", 0), true)
	if !errors.Is(err, ErrTooDeep) || !strings.Contains(err.Error(), "#4 Home automation holds sub-objectives") {
		t.Errorf("MayHang = %v, want #4 refused for holding sub-objectives", err)
	}
}

func TestMayHang_refusesItself(t *testing.T) {
	if err := MayHang(named(4, "Home automation", 0), named(4, "Home automation", 0), false); !errors.Is(err, ErrUnderItself) {
		t.Errorf("MayHang = %v, want ErrUnderItself", err)
	}
}

func TestCountedOnce_letsTwoSubObjectivesOfOneParent(t *testing.T) {
	if err := CountedOnce(44, []Objective{named(5, "Enclosure", 4), named(6, "Firmware", 4)}); err != nil {
		t.Errorf("CountedOnce: %v, want it allowed", err)
	}
}

func TestCountedOnce_refusesAParentBesideItsSubObjectives(t *testing.T) {
	err := CountedOnce(44, []Objective{
		named(4, "Home automation", 0), named(5, "Enclosure", 4), named(6, "Firmware", 4), named(9, "Garden", 0),
	})
	want := "#44 is already counted in #4 Home automation through #5 Enclosure and #6 Firmware; nothing was changed"
	if !errors.Is(err, ErrCountedTwice) || err == nil || err.Error() != want {
		t.Errorf("CountedOnce = %v, want %q", err, want)
	}
}

func TestCountedOnce_namesThreeWithCommas(t *testing.T) {
	err := CountedOnce(0, []Objective{
		named(7, "Features", 4), named(4, "Home automation", 0), named(5, "Enclosure", 4), named(6, "Firmware", 4),
	})
	want := "the task is already counted in #4 Home automation through #5 Enclosure, #6 Firmware and #7 Features; nothing was changed"
	if err == nil || err.Error() != want {
		t.Errorf("CountedOnce = %v, want %q", err, want)
	}
}

func TestCountedOnce_namesTheFirstTreeCountedTwice(t *testing.T) {
	err := CountedOnce(1, []Objective{
		named(9, "Garden", 0), named(10, "Valves", 9), named(4, "Home", 0), named(5, "Enclosure", 4),
	})
	if err == nil || !strings.Contains(err.Error(), "in #4 Home through #5 Enclosure;") {
		t.Errorf("CountedOnce = %v, want the tree of #4 named first", err)
	}
}

func TestProposedReached_onlyWhenEverySubObjectiveIsAndItIsNot(t *testing.T) {
	reached := func(o Objective) Objective {
		a := From(o)
		a.Standing(Reached)
		return a.Objective()
	}
	parent := named(4, "Home automation", 0)
	for _, c := range []struct {
		parent   Objective
		children []Objective
		want     bool
	}{
		{parent, nil, false},
		{parent, []Objective{reached(named(5, "a", 4)), named(6, "b", 4)}, false},
		{parent, []Objective{reached(named(5, "a", 4)), reached(named(6, "b", 4))}, true},
		{reached(parent), []Objective{reached(named(5, "a", 4))}, false},
	} {
		if got := ProposedReached(c.parent, c.children); got != c.want {
			t.Errorf("ProposedReached(%v, %d children) = %v, want %v", c.parent.Carried().State, len(c.children), got, c.want)
		}
	}
}

func TestText_saysWhatItSitsUnder(t *testing.T) {
	got := Text(named(5, "Enclosure", 4), domain.Progress{}, Review{}, Family{Parent: named(4, "Home automation", 0)})
	if !strings.Contains(got, "parent       #4 Home automation\n") {
		t.Errorf("text does not say its parent:\n%s", got)
	}
	alone := Text(plan(t), domain.Progress{}, Review{}, Family{})
	if !strings.Contains(alone, "parent       —\n") {
		t.Errorf("text does not say it sits under nothing:\n%s", alone)
	}
}

func TestChildren_listsEachSubObjectiveWithItsCounts(t *testing.T) {
	got := Children(Family{Children: []Child{
		{Objective: named(5, "Enclosure", 4), Progress: domain.Progress{Tasks: domain.Tally{Total: 6, Done: 3}}, Spent: 8*time.Hour + 25*time.Minute},
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
	}, map[int]domain.Progress{4: {Tasks: domain.Tally{Total: 4, Done: 1}}, 5: {Tasks: domain.Tally{Total: 2, Done: 1}}},
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
	got := Report(plan(t), domain.Progress{}, Review{}, in)
	if !strings.Contains(got, "\n  2026-09  2h30m  · #2 2h  · #3 —  · itself 30m\n") {
		t.Errorf("the report does not split the month:\n%s", got)
	}
}
