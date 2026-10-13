package gui

import (
	"cairn/internal/engine"
	"slices"
	"testing"

	"gioui.org/io/pointer"
)

// The menu opens down and to the right of the point clicked, and is kept
// inside the window near an edge.
func TestTheMenuIsLaidOutBesideThePointer(t *testing.T) {
	w := Window{width: 1200, height: 800, menuState: menuState{menu: []string{"Edit", "Focus", "Archive", "Delete"}}}
	w.menuAt = Point{X: 300, Y: 200}
	if got, want := w.MenuBox(), (Box{X: 300, Y: 200, W: 180, H: 124}); got != want {
		t.Errorf("the menu is at %+v, want %+v", got, want)
	}
	if got, want := MenuRowAt(w.MenuBox(), 2), (Box{X: 306, Y: 262, W: 168, H: 28}); got != want {
		t.Errorf("its third line is at %+v, want %+v", got, want)
	}
	w.menuAt = Point{X: 1150, Y: 790}
	if got, want := w.MenuBox(), (Box{X: 1020, Y: 676, W: 180, H: 124}); got != want {
		t.Errorf("near the corner the menu is at %+v, want %+v", got, want)
	}
}

// The keys walk the menu both ways and stop at its ends.
func TestTheMenuIsWalkedBothWays(t *testing.T) {
	w := Window{screen: OnMenu, menuState: menuState{menu: []string{"Edit", "Delete"}}}
	if got := w.onMenu("up").MenuOn(); got != 0 {
		t.Errorf("up from the first line is on %d", got)
	}
	down := w.onMenu("j").onMenu("down")
	if got := down.MenuOn(); got != 1 {
		t.Errorf("down past the last line is on %d", got)
	}
	if got := down.onMenu("k").MenuOn(); got != 0 {
		t.Errorf("k is on %d, want 0", got)
	}
}

// A right press in the loop opens the menu on the card under it.
func TestTheSecondButtonOpensTheMenu(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w, _, _ = frames(t, w, pointer.Event{Kind: pointer.Press, Position: at(w, 1, 0), Buttons: pointer.ButtonSecondary})
	if w.Screen() != OnMenu || w.Chosen() != 1 {
		t.Errorf("a right press left the window on %v, choosing %d", w.Screen(), w.Chosen())
	}
}

// The menu is drawn, the line the keyboard is on lit.
func TestTheMenuIsDrawn(t *testing.T) {
	w := Window{screen: OnMenu, width: 1200, height: 800, menuState: menuState{menu: []string{"Edit", "Delete"}, menuAt: Point{X: 300, Y: 200}}}
	shot := drawn(t, w)
	corner := func(i int) Box {
		row := MenuRowAt(w.MenuBox(), i)
		return Box{X: row.X + 1, Y: row.Y + 1, W: 1, H: 1}
	}
	if got := ink(shot, corner(0), Dark().Back); got != 0 {
		t.Error("the line the keyboard is on is not lit")
	}
	if got := ink(shot, corner(1), Dark().Panel); got != 0 {
		t.Error("a line the keyboard is not on is lit")
	}
	if got := ink(shot, Edges(w.MenuBox())[0], Dark().Back); got == 0 {
		t.Error("the menu has no edge")
	}
}

// Only the press opens it: a pointer moving with the second button held, or
// coming in with it, opens nothing.
func TestOnlyASecondPressOpensTheMenu(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w, _, _ = frames(t, w, pointer.Event{Kind: pointer.Move, Position: at(w, 1, 0), Buttons: pointer.ButtonSecondary})
	if w.Screen() != OnTheBoard {
		t.Errorf("a pointer moving with the second button held left the window on %v", w.Screen())
	}
}

// A task's menu reaches what the keys and a drag do to it: its column and its
// priority, offered in the menu itself, every value but the one it holds,
// and taken at once; its labels in the list a form opens for them.
func TestACardsMenuChangesItsColumnAndPriority(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "A", State: "todo", Priority: "normal"})
	opened := w.ContextAt(middleOf(w, 1, 0))
	if got := opened.Menu(); !slices.Equal(got, []string{"Edit", "Focus", "Move to…", "Priority…", "Labels…", "Archive", "Delete"}) {
		t.Fatalf("a task's menu offers %v", got)
	}
	moving := opened.fromMenu(slices.Index(opened.Menu(), "Move to…"))
	if moving.Screen() != OnMenu || !slices.Equal(moving.Menu(), []string{"backlog", "current", "blocked", "watching", "paused", "done"}) {
		t.Fatalf("Move to… left the window on %v offering %v", moving.Screen(), moving.Menu())
	}
	moved := moving.fromMenu(slices.Index(moving.Menu(), "done"))
	if in, ok := f.gave("task", "edit"); !ok || in.Given["number"] != "1" || in.Given["state"] != "done" || moved.Screen() != OnTheBoard {
		t.Errorf("done asked %+v, the window on %v", in, moved.Screen())
	}
	rating := opened.fromMenu(slices.Index(opened.Menu(), "Priority…"))
	if !slices.Equal(rating.Menu(), []string{"low", "high", "critical"}) {
		t.Fatalf("Priority… offers %v", rating.Menu())
	}
	rating.fromMenu(slices.Index(rating.Menu(), "high"))
	if in, _ := f.gave("task", "edit"); in.Given["priority"] != "high" {
		t.Errorf("high asked %+v", in)
	}
	labelling := opened.fromMenu(slices.Index(opened.Menu(), "Labels…"))
	if labelling.Screen() != OnChoosing || labelling.choosingFor != "label" {
		t.Errorf("Labels… left the window on %v choosing %q", labelling.Screen(), labelling.choosingFor)
	}
	// A menu put away and opened again is the card's own again.
	if again := moving.Key("esc").ContextAt(middleOf(w, 1, 0)); again.Menu()[2] != "Move to…" {
		t.Errorf("opened again, the menu offers %v", again.Menu())
	}
}
