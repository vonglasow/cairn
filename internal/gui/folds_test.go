package gui

import (
	"slices"
	"strings"
	"testing"
	"time"

	"cairn/test"
)

func treeNumbers(w Window) []int {
	var got []int
	for _, b := range w.Tree() {
		got = append(got, b.Goal.Number)
	}
	return got
}

func chooseGoal(w Window, n int) Window {
	w.chosen, w.chosenGoal = n, true
	return w
}

func TestAParentStartsFoldedAndItsChevronUnfoldsIt(t *testing.T) {
	w := homeWindow(1200, 800)
	if got := treeNumbers(w); !slices.Equal(got, []int{1, 4}) {
		t.Fatalf("the tree opens on %v", got)
	}
	if !w.Folded(1) || w.Folded(4) {
		t.Errorf("Folded(1) = %v, Folded(4) = %v", w.Folded(1), w.Folded(4))
	}
	chevron := FoldAt(GoalAt(w.Layout(), 0, 0))
	w = w.clicked(Point{X: chevron.X + chevron.W/2, Y: chevron.Y + chevron.H/2})
	if got := treeNumbers(w); !slices.Equal(got, []int{1, 2, 3, 4}) {
		t.Fatalf("the chevron unfolds to %v", got)
	}
	if w.chosen != 0 {
		t.Errorf("the chevron chose %d", w.chosen)
	}
	w = w.clicked(Point{X: chevron.X + chevron.W/2, Y: chevron.Y + chevron.H/2})
	if got := treeNumbers(w); !slices.Equal(got, []int{1, 4}) {
		t.Errorf("the chevron folds to %v", got)
	}
}

func TestTheArrowsUnfoldGoInClimbAndFold(t *testing.T) {
	w := chooseGoal(homeWindow(1200, 800), 1)
	w = w.Key("right")
	if w.Folded(1) || w.chosen != 1 {
		t.Fatalf("right leaves folded %v, chosen %d", w.Folded(1), w.chosen)
	}
	w = w.Key("right")
	if w.chosen != 2 {
		t.Fatalf("right again chose %d", w.chosen)
	}
	w = w.Key("right")
	if w.chosen != 2 {
		t.Errorf("right on a sub-objective chose %d", w.chosen)
	}
	w = w.Key("left")
	if w.chosen != 1 || w.Folded(1) {
		t.Fatalf("left on a sub-objective chose %d, folded %v", w.chosen, w.Folded(1))
	}
	w = w.Key("left")
	if !w.Folded(1) || w.chosen != 1 {
		t.Errorf("left on an open parent leaves folded %v, chosen %d", w.Folded(1), w.chosen)
	}
}

func TestZTogglesAndClimbsAndOptionArrowsDoThemAll(t *testing.T) {
	w := chooseGoal(homeWindow(1200, 800), 1)
	w = w.Key("z")
	if w.Folded(1) {
		t.Fatal("z leaves the parent folded")
	}
	w = chooseGoal(w, 3).Key("z")
	if !w.Folded(1) || w.chosen != 1 {
		t.Errorf("z on a sub-objective: folded %v, chosen %d", w.Folded(1), w.chosen)
	}
	w = w.Key("alt+right")
	if w.Folded(1) {
		t.Error("option right leaves a parent folded")
	}
	w = chooseGoal(w, 2).Key("alt+left")
	if !w.Folded(1) || w.chosen != 1 {
		t.Errorf("option left: folded %v, chosen %d", w.Folded(1), w.chosen)
	}
}

func TestDownWalksPastAFoldedParentsSubObjectives(t *testing.T) {
	w := chooseGoal(homeWindow(1200, 800), 1).Key("down")
	if w.chosen != 4 {
		t.Errorf("down chose %d", w.chosen)
	}
}

func TestAParentHoldingWhatIsChosenIsDrawnOpen(t *testing.T) {
	w := chooseGoal(homeWindow(1200, 800), 3)
	if w.Folded(1) {
		t.Error("the parent of the one chosen is folded")
	}
}

func TestASearchUnfoldsTheParentOfAMatchAndClearingItFoldsItBack(t *testing.T) {
	w := chooseGoal(homeWindow(1200, 800), 4)
	w.snap = test.Searchable(w.snap)
	w.looking = Written{Value: "firm", At: 4}
	tree := w.Tree()
	if len(tree) != 2 || tree[0].Goal.Number != 1 || !tree[0].Holds || tree[1].Goal.Number != 3 {
		t.Fatalf("the search draws %+v", tree)
	}
	if g, _ := w.Detailed(); g.Number != 3 {
		t.Errorf("the detail shows %d", g.Number)
	}
	if said := w.GoalsCount(); said != "1 of 4" {
		t.Errorf("the head counts %q", said)
	}
	w.looking = Written{}
	if got := treeNumbers(w); !slices.Equal(got, []int{1, 4}) {
		t.Errorf("the cleared search leaves %v", got)
	}
}

func TestASubObjectiveChosenInASearchKeepsItsParentOpen(t *testing.T) {
	w := homeWindow(1200, 800)
	w.snap = test.Searchable(w.snap)
	w.looking = Written{Value: "firm", At: 4}
	w = w.choosingAt(3)
	w.looking = Written{}
	if got := treeNumbers(w); !slices.Equal(got, []int{1, 2, 3, 4}) {
		t.Errorf("the cleared search leaves %v", got)
	}
}

func TestAFoldedParentSaysWhatItHides(t *testing.T) {
	w := homeWindow(1200, 800).At(time.Date(2026, 11, 10, 9, 0, 0, 0, time.Local))
	w.snap.Goals[2].WaitsOn = []int{2}
	if said, dot := w.SignalOf(w.Tree()[0]); said != "#3 waits on #2" || !dot {
		t.Errorf("the folded parent says %q, dot %v", said, dot)
	}
	w.snap.Goals[1].Deadline = time.Date(2026, 11, 6, 0, 0, 0, 0, time.Local)
	if said, _ := w.SignalOf(w.Tree()[0]); said != "#2 Enclosure late +1" {
		t.Errorf("the folded parent says %q", said)
	}
	if lines := w.HiddenSays(1); len(lines) != 4 || lines[1] != "#2 Enclosure, its deadline Fri 6 Nov past, not reached." || !strings.Contains(lines[2], "#3 Firmware waits on #2 Enclosure") {
		t.Errorf("the bubble says %q", lines)
	}
	w = w.Key("alt+right")
	if said, _ := w.SignalOf(w.Tree()[0]); said != "" {
		t.Errorf("the unfolded parent says %q", said)
	}
	if said, dot := w.SignalOf(w.Tree()[1]); said != "6 Nov, late" || dot {
		t.Errorf("the late one says %q", said)
	}
	if said, _ := w.SignalOf(w.Tree()[2]); said != "waits on Enclosure" {
		t.Errorf("the waiting one says %q", said)
	}
	narrow := w.Resized(720, 560).Key("alt+left")
	if said, _ := narrow.SignalOf(narrow.Tree()[0]); said != "#2 late +1" {
		t.Errorf("at 720 the folded parent says %q", said)
	}
}
