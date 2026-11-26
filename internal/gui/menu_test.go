package gui

import (
	"cairn/internal/board"
	"cairn/internal/gui/geom"
	"slices"
	"strings"
	"testing"
)

func TestTheMenuIsWalkedBothWays(t *testing.T) {
	w := Window{Board: board.Board{Screen: board.OnMenu}, menuState: menuState{menu: []string{"Edit", "Delete"}}}
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

func TestACardsMenuChangesItsColumnAndPriority(t *testing.T) {
	f := newFake()
	w := holding(f, board.Card{Number: 1, Title: "A", State: "todo", Priority: "normal"})
	opened := w.ContextAt(middleOf(w, 1, 0))
	if got := opened.Menu(); !slices.Equal(got, []string{"Open", "Edit", "New linked task", "Move to…", "Priority…", "Labels…", "Archive", "Focus", StartSessionSays, "Delete"}) {
		t.Fatalf("a task's menu offers %v", got)
	}
	moving := opened.fromMenu(slices.Index(opened.Menu(), "Move to…"))
	if moving.Screen != board.OnMenu || !slices.Equal(moving.Menu(), []string{"backlog", "current", "blocked", "watching", "paused", "done"}) {
		t.Fatalf("Move to… left the window on %v offering %v", moving.Screen, moving.Menu())
	}
	moved := moving.fromMenu(slices.Index(moving.Menu(), "done"))
	if in, ok := f.gave("task", "edit"); !ok || in.Given["number"] != "1" || in.Given["state"] != "done" || moved.Screen != board.OnTheBoard {
		t.Errorf("done asked %+v, the window on %v", in, moved.Screen)
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
	if labelling.Screen != board.OnChoosing || labelling.ChoosingFor != "label" {
		t.Errorf("Labels… left the window on %v choosing %q", labelling.Screen, labelling.ChoosingFor)
	}
	if again := moving.Key("esc").ContextAt(middleOf(w, 1, 0)); again.Menu()[3] != "Move to…" {
		t.Errorf("opened again, the menu offers %v", again.Menu())
	}
}

func TestAFilterInForceIsAPillWhoseCrossClearsIt(t *testing.T) {
	w := holding(newFake(), board.Card{Number: 1, Title: "Write the spec", State: "todo", Priority: "high"})
	w = w.Key("shift+f")
	for w.Filters[w.FilterOn].Name != "priority" {
		w = w.Key("down")
	}
	w = w.Key("enter").Key("enter").Key("esc")
	var filters []FilterPill
	for _, pill := range w.Pills() {
		if pill.Kind == PillFilter {
			filters = append(filters, pill)
		}
	}
	if len(filters) != 1 || !strings.HasPrefix(filters[0].Said, "priority: ") {
		t.Fatalf("the filters in force are %+v", filters)
	}
	cross, _ := filters[0].Cross()
	w = w.clicked(geom.Point{X: cross.X + cross.W/2, Y: cross.Y + cross.H/2})
	if w.FilteredBy() != "" {
		t.Errorf("the cross left the board filtered by %q", w.FilteredBy())
	}
}

func TestOpenFirstInACardsMenuOpensIt(t *testing.T) {
	w := holding(newFake(), board.Card{Number: 1, Title: "A", State: "todo"})
	opened := w.ContextAt(middleOf(w, 1, 0))
	if MenuKey("Open") != "enter" {
		t.Errorf("Open says the key %q", MenuKey("Open"))
	}
	if got := opened.fromMenu(0); got.Screen != board.OnForm {
		t.Errorf("Open left the window on %v", got.Screen)
	}
}
