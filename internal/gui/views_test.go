package gui

import (
	"cairn/internal/board"
	"testing"
	"time"

	"cairn/internal/engine"
)

func viewedHome(width, height int) Window {
	w := homeWindow(width, height).At(time.Now())
	w.Tab = board.Tasks
	w.Snap.Views = []board.Viewing{
		{View: engine.View{Name: "Home", Place: 1, Holds: engine.Holds{Objective: []int{1}}}, Kept: []int{1, 2, 3}},
		{View: engine.View{Name: "Done", Place: 2, Holds: engine.Holds{State: "done"}}, Kept: []int{2}},
	}
	return w
}

func TestTheKeyboardWalksUpFromTheFiltersIntoTheViews(t *testing.T) {
	w := viewedHome(1200, 800).Key("shift+f")
	if w.ViewWalked != 0 {
		t.Fatalf("the menu opens on the view %d", w.ViewWalked)
	}
	w = w.Key("up")
	if w.ViewWalked != 2 || w.ViewRows()[1].On != true || w.ViewRows()[0].On {
		t.Fatalf("up from the first filter is on the view %d: %+v", w.ViewWalked, w.ViewRows())
	}
	w = w.Key("up").Key("up")
	if w.ViewWalked != 1 {
		t.Errorf("up past the first view is on %d", w.ViewWalked)
	}
	if back := w.Key("down").Key("down"); back.ViewWalked != 0 || back.FilterOn != 0 {
		t.Errorf("down past the last view is on the view %d and the filter %d", back.ViewWalked, back.FilterOn)
	}
	opened := w.Key("enter")
	if opened.Viewing != "Home" || opened.Screen != board.OnTheBoard {
		t.Errorf("enter on the first view opens %q on %v", opened.Viewing, opened.Screen)
	}
	if none := homeWindow(1200, 800).At(time.Now()).Key("shift+f").Key("up"); none.ViewWalked != 0 {
		t.Error("with no view, up walks into a view")
	}
}
