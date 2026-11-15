package board

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

func TestFoundBranches_aParentMatchingStandsAsItself(t *testing.T) {
	tree := FoundBranches(foldedTree().Goals, "home")
	if len(tree) != 1 || tree[0].Holds {
		t.Fatalf("FoundBranches = %+v", tree)
	}
	if every := FoundBranches(foldedTree().Goals, ""); len(every) != 5 {
		t.Errorf("nothing typed finds %d", len(every))
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
