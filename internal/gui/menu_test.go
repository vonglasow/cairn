package gui

import (
	"cairn/internal/engine"
	"slices"
	"strings"
	"testing"

	"gioui.org/io/pointer"
)

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

func TestTheSecondButtonOpensTheMenu(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w, _, _ = frames(t, w, pointer.Event{Kind: pointer.Press, Position: at(w, 1, 0), Buttons: pointer.ButtonSecondary})
	if w.Screen() != OnMenu || w.Chosen() != 1 {
		t.Errorf("a right press left the window on %v, choosing %d", w.Screen(), w.Chosen())
	}
}

func TestTheMenuIsDrawn(t *testing.T) {
	w := Window{screen: OnMenu, width: 1200, height: 800, menuState: menuState{menu: []string{"Edit", "Delete"}, menuAt: Point{X: 300, Y: 200}}}
	shot := drawn(t, w)
	corner := func(i int) Box {
		row := MenuRowAt(w.MenuBox(), i)
		return Box{X: row.X + row.W/2, Y: row.Y + 1, W: 1, H: 1}
	}
	if got := ink(shot, corner(0), Dark().Pill); got != 0 {
		t.Error("the line the keyboard is on is not lit")
	}
	if got := ink(shot, corner(1), Dark().Panel); got != 0 {
		t.Error("a line the keyboard is not on is lit")
	}
	if got := ink(shot, Edges(w.MenuBox())[0], Dark().Back); got == 0 {
		t.Error("the menu has no edge")
	}
}

func TestOnlyASecondPressOpensTheMenu(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w, _, _ = frames(t, w, pointer.Event{Kind: pointer.Move, Position: at(w, 1, 0), Buttons: pointer.ButtonSecondary})
	if w.Screen() != OnTheBoard {
		t.Errorf("a pointer moving with the second button held left the window on %v", w.Screen())
	}
}

func TestACardsMenuChangesItsColumnAndPriority(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "A", State: "todo", Priority: "normal"})
	opened := w.ContextAt(middleOf(w, 1, 0))
	if got := opened.Menu(); !slices.Equal(got, []string{"Open", "Edit", "New linked task", "Focus", "Move to…", "Priority…", "Labels…", "Archive", "Delete"}) {
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
	if again := moving.Key("esc").ContextAt(middleOf(w, 1, 0)); again.Menu()[4] != "Move to…" {
		t.Errorf("opened again, the menu offers %v", again.Menu())
	}
}

func TestTheMenusLitLineStandsOnAPillAndEachLineSaysItsKey(t *testing.T) {
	w := holding(newFake(), engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	at := middleOf(w, 1, 0)
	w = w.ContextAt(at)
	box := w.MenuBox()
	row := MenuRowAt(box, 0)
	shot := drawn(t, w)
	if painted(shot, Box{X: row.X + row.W/2, Y: row.Y + 1, W: 1, H: 1}, Dark().Pill) != 1 {
		t.Error("the menu's lit line does not stand on a pill")
	}
	if painted(shot, Box{X: box.X, Y: box.Y, W: 1, H: 1}, Dark().Panel) != 0 {
		t.Error("the menu's corner is not rounded")
	}
	if got := MenuKey("Edit"); got != "e" {
		t.Errorf("the menu's Edit line says the key %q", got)
	}
	if got := MenuKey("Delete"); got != "cmd+backspace" {
		t.Errorf("the menu's Delete line says the key %q", got)
	}
}

func TestAFilterInForceIsAPillWhoseCrossClearsIt(t *testing.T) {
	w := holding(newFake(), engine.Card{Number: 1, Title: "Write the spec", State: "todo", Priority: "high"})
	w = w.Key("shift+f")
	for w.Filters()[w.FilterOn()].Name != "priority" {
		w = w.Key("down")
	}
	w = w.Key("right").Key("esc")
	pills := w.FilterPills()
	if len(pills) != 1 || !strings.HasPrefix(pills[0].Said, "priority: ") {
		t.Fatalf("the filters in force are %+v", pills)
	}
	cross := pills[0].Box
	w = w.clicked(Point{X: cross.X + cross.W - pillCross/2, Y: cross.Y + cross.H/2})
	if len(w.FilterPills()) != 0 || w.FilteredBy() != "" {
		t.Errorf("the cross left the board filtered by %q", w.FilteredBy())
	}
}

func TestOpenFirstInACardsMenuOpensIt(t *testing.T) {
	w := holding(newFake(), engine.Card{Number: 1, Title: "A", State: "todo"})
	opened := w.ContextAt(middleOf(w, 1, 0))
	if MenuKey("Open") != "enter" {
		t.Errorf("Open says the key %q", MenuKey("Open"))
	}
	if got := opened.fromMenu(0); got.Screen() != OnForm {
		t.Errorf("Open left the window on %v", got.Screen())
	}
}
