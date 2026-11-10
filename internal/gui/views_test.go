package gui

import (
	"testing"
	"time"

	"cairn/internal/engine"
)

func viewedHome(width, height int) Window {
	w := homeWindow(width, height).At(time.Now())
	w.tab = Tasks
	w.snap.Views = []engine.Viewing{
		{View: engine.View{Name: "Home", Place: 1, Holds: engine.Holds{Objective: []int{1}}}, Kept: []int{1, 2, 3}},
		{View: engine.View{Name: "Done", Place: 2, Holds: engine.Holds{State: "done"}}, Kept: []int{2}},
	}
	return w
}

func TestTheKeyboardWalksUpFromTheFiltersIntoTheViews(t *testing.T) {
	w := viewedHome(1200, 800).Key("shift+f")
	if w.ViewWalked() != 0 {
		t.Fatalf("the menu opens on the view %d", w.ViewWalked())
	}
	w = w.Key("up")
	if w.ViewWalked() != 2 || w.ViewRows()[1].On != true || w.ViewRows()[0].On {
		t.Fatalf("up from the first filter is on the view %d: %+v", w.ViewWalked(), w.ViewRows())
	}
	w = w.Key("up").Key("up")
	if w.ViewWalked() != 1 {
		t.Errorf("up past the first view is on %d", w.ViewWalked())
	}
	if back := w.Key("down").Key("down"); back.ViewWalked() != 0 || back.FilterOn() != 0 {
		t.Errorf("down past the last view is on the view %d and the filter %d", back.ViewWalked(), back.FilterOn())
	}
	opened := w.Key("enter")
	if opened.Viewing() != "Home" || opened.Screen() != OnTheBoard {
		t.Errorf("enter on the first view opens %q on %v", opened.Viewing(), opened.Screen())
	}
	if none := homeWindow(1200, 800).At(time.Now()).Key("shift+f").Key("up"); none.ViewWalked() != 0 {
		t.Error("with no view, up walks into a view")
	}
}

func TestAViewNamingAnObjectiveHoldingSubObjectivesDrawsTheNarrowingBar(t *testing.T) {
	w := viewedHome(1200, 800).Key("cmd+1")
	n := w.Narrowing()
	if n.Root != 1 || len(n.At) != 4 || n.On != 0 {
		t.Fatalf("the view draws the bar %+v", n)
	}
	sub := w.narrowedTo(1)
	if taken, _ := sub.objectiveTaken(); taken != "2" || !sub.ViewChanged() {
		t.Errorf("narrowed to Enclosure the filter on top is %q", taken)
	}
	if back := sub.narrowedTo(0); back.ViewChanged() || back.Narrowing().On != 0 {
		t.Errorf("narrowed back to all, the view is changed: %v", back.ViewChanged())
	}
	if other := w.Key("cmd+2"); other.Narrowing().Box.W != 0 {
		t.Error("a view naming no objective draws the bar")
	}
}

func TestTheObjectivesTabCountsOnlyWhatItShowsWhenAViewLeavesThemWhole(t *testing.T) {
	w := viewedHome(1200, 800).Key("cmd+2")
	if got := w.TabCount(Objectives); got != "4" {
		t.Errorf("under a view narrowing no objective the tab counts %q", got)
	}
}
