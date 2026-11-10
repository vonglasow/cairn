package engine

import (
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
