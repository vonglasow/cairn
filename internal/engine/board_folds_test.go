package engine

import (
	"slices"
	"testing"
	"time"
)

func foldedTree() Snapshot {
	s := Snapshot{Goals: []Goal{
		{Number: 1, Title: "Home automation", State: "active", Children: []int{2, 3, 4}},
		{Number: 2, Title: "Enclosure", State: "active", Parent: 1},
		{Number: 3, Title: "Firmware", State: "active", Parent: 1, WaitsOn: []int{2}},
		{Number: 4, Title: "Documentation", State: "active", Parent: 1, Deadline: time.Date(2026, 11, 6, 0, 0, 0, 0, time.Local)},
		{Number: 5, Title: "Tidy the workshop", State: "active"},
	}}
	for i, g := range s.Goals {
		s.Goals[i].Said = []string{g.Title}
	}
	return s
}

var foldDay = time.Date(2026, 11, 10, 9, 0, 0, 0, time.Local)

func TestHidden_putsTheLateFirstThenWhatWaits(t *testing.T) {
	hid := foldedTree().Hidden(1, foldDay)
	if len(hid) != 2 || hid[0].Goal.Number != 4 || !hid[0].Late || hid[1].Goal.Number != 3 || hid[1].Late {
		t.Fatalf("Hidden = %+v", hid)
	}
	if hid[1].On.Number != 2 {
		t.Errorf("what #3 waits on = %+v", hid[1].On)
	}
}

func TestHidden_saysNothingOfWhatIsReachedOrDropped(t *testing.T) {
	s := foldedTree()
	s.Goals[3].State = "reached"
	s.Goals[1].State = "reached"
	if hid := s.Hidden(1, foldDay); len(hid) != 0 {
		t.Errorf("Hidden = %+v", hid)
	}
	s.Goals[3].State = "dropped"
	if hid := s.Hidden(1, foldDay); len(hid) != 0 {
		t.Errorf("a dropped one is hidden as %+v", hid)
	}
}

func TestLateOn_isPastItsDeadlineAndNotReached(t *testing.T) {
	g := Goal{State: "active", Deadline: time.Date(2026, 11, 10, 0, 0, 0, 0, time.Local)}
	if g.LateOn(foldDay) {
		t.Error("an objective due today is late")
	}
	g.Deadline = g.Deadline.AddDate(0, 0, -1)
	if !g.LateOn(foldDay) {
		t.Error("an objective due yesterday is not late")
	}
	if (Goal{State: "active"}).LateOn(foldDay) {
		t.Error("an objective with no deadline is late")
	}
}

func TestFoundBranches_holdsTheParentOfAMatch(t *testing.T) {
	tree := FoundBranches(foldedTree().Goals, "firm")
	if len(tree) != 2 || tree[0].Goal.Number != 1 || !tree[0].Holds || tree[1].Goal.Number != 3 || !tree[1].Under || !tree[1].Last {
		t.Fatalf("FoundBranches = %+v", tree)
	}
	if tree[1].Holds {
		t.Error("the match itself is drawn as what holds it")
	}
}

func TestFoundBranches_aParentMatchingStandsAsItself(t *testing.T) {
	tree := FoundBranches(foldedTree().Goals, "home")
	if len(tree) != 1 || tree[0].Holds {
		t.Fatalf("FoundBranches = %+v", tree)
	}
	if every := FoundBranches(foldedTree().Goals, ""); len(every) != 5 {
		t.Errorf("nothing typed finds %d", len(every))
	}
}

func TestHidden_passesOverWhatIsReachedOrDroppedAndSaysEveryOther(t *testing.T) {
	today := time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC)
	s := Snapshot{Goals: []Goal{
		{Number: 1, Children: []int{2, 3, 4, 5, 6}},
		{Number: 2, Parent: 1, State: "reached", WaitsOn: []int{9}},
		{Number: 3, Parent: 1, State: "dropped", WaitsOn: []int{9}},
		{Number: 4, Parent: 1, Deadline: today.AddDate(0, 0, -1)},
		{Number: 5, Parent: 1, WaitsOn: []int{9}},
		{Number: 6, Parent: 1, Deadline: today.AddDate(0, 0, -2)},
		{Number: 9},
	}}
	var got []int
	for _, h := range s.Hidden(1, today) {
		got = append(got, h.Goal.Number)
	}
	if !slices.Equal(got, []int{4, 6, 5}) {
		t.Errorf("hidden %v, want the late 4 and 6, then 5 waiting", got)
	}
}

func TestFoundBranches_standsEachParentHoldingAMatchOnce(t *testing.T) {
	goals := []Goal{
		{Number: 1, Title: "match top"},
		{Number: 2, Title: "holder", Children: []int{3, 4}},
		{Number: 3, Title: "match a", Parent: 2},
		{Number: 4, Title: "match b", Parent: 2},
		{Number: 5, Title: "match parent", Children: []int{6}},
		{Number: 6, Title: "match child", Parent: 5},
	}
	var got []int
	for i := range goals {
		goals[i].Said = []string{goals[i].Title}
	}
	for _, b := range FoundBranches(goals, "match") {
		got = append(got, b.Goal.Number)
	}
	slices.Sort(got)
	if !slices.Equal(got, []int{1, 2, 3, 4, 5, 6}) {
		t.Errorf("found %v, want each once, the holder among them", got)
	}
}
