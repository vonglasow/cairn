package gui

import (
	"testing"

	"cairn/internal/engine"
	"cairn/test"
)

func TestSearchingNarrowsTheBoardToWhatItReaches(t *testing.T) {
	f := newFake()
	w := holding(f,
		engine.Card{Number: 1, Title: "Write the spec", State: "backlog", Description: "two columns"},
		engine.Card{Number: 2, Title: "Ship the board", State: "backlog"})

	w = w.Key("/")
	if w.Screen() != OnSearch {
		t.Fatalf("/ left the window on %v", w.Screen())
	}
	w = w.Types("spec")
	if got := len(w.Showing().Cards); got != 1 {
		t.Fatalf("a search for \"spec\" reaches %d tasks", got)
	}

	deep := w.Key("esc").Key("/").Types("columns")
	if got := len(deep.Showing().Cards); got != 1 {
		t.Errorf("a search of the description reaches %d tasks", got)
	}
	if got := deep.Showing().Cards[0].Number; got != 1 {
		t.Errorf("it reached task %d", got)
	}
}

func TestLeavingASearchPutsEverythingBack(t *testing.T) {
	f := newFake()
	w := holding(f,
		engine.Card{Number: 1, Title: "Write the spec", State: "backlog"},
		engine.Card{Number: 2, Title: "Ship the board", State: "backlog"})

	w = w.Key("/").Types("spec")
	if got := len(w.Showing().Cards); got != 1 {
		t.Fatalf("the search reaches %d", got)
	}
	w = w.Key("esc")
	if got := len(w.Showing().Cards); got != 2 {
		t.Errorf("leaving the search left %d tasks showing", got)
	}
	if w.Screen() != OnTheBoard {
		t.Errorf("leaving the search left the window on %v", w.Screen())
	}
}

func TestASearchThatReachesNothingShowsNothing(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "backlog"})
	w = w.Key("/").Types("nothing like it")
	if got := len(w.Showing().Cards); got != 0 {
		t.Errorf("a search that reaches nothing shows %d tasks", got)
	}
}

func TestASearchReachesObjectives(t *testing.T) {
	f := newFake()
	w := New(f, 1200, 800)
	w.snap = test.Searchable(engine.Snapshot{Goals: []engine.Goal{
		{Number: 3, Title: "Ship cairn", State: "active"},
		{Number: 4, Title: "Keep it honest", State: "active"},
	}})
	w.tab = Objectives

	w = w.Key("/").Types("honest")
	if got := len(w.Showing().Goals); got != 1 {
		t.Fatalf("a search of the objectives reaches %d", got)
	}
	if got := w.Showing().Goals[0].Number; got != 4 {
		t.Errorf("it reached objective %d", got)
	}
}

func TestBackspaceInASearchTakesBackOneCharacter(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "réunion", State: "backlog"})

	w = w.Key("/").Types("réu")
	if got := w.Key("backspace").looking.Reads(); got != "ré" {
		t.Fatalf("taking one back left %q", got)
	}
	if got := w.Key("backspace").Key("backspace").looking.Reads(); got != "r" {
		t.Errorf("taking two back left %q", got)
	}
	empty := w.Key("esc").Key("/")
	for range 3 {
		empty = empty.Key("backspace")
	}
	if got := empty.looking.Reads(); got != "" {
		t.Errorf("an empty search reads %q", got)
	}
	if got := w.Key("esc").Key("backspace").looking.Reads(); got != "" {
		t.Errorf("backspace on the board left a search of %q", got)
	}
}

func TestASearchCanBeLeftWithThePointer(t *testing.T) {
	f := newFake()
	w := holding(f, card(1, "Write the spec", "todo")).Choose("Search")
	if w.Screen() != OnSearch {
		t.Fatalf("the window is on %v", w.Screen())
	}
	if len(w.Offers()) == 0 {
		t.Fatal("a search offers nothing, so nothing can be pointed at")
	}
	at := OfferAt(w.Layout(), len(w.Offers())-1)
	got := w.clicked(Point{X: at.X + 10, Y: at.Y + at.H/2})
	if got.Screen() == OnSearch {
		t.Error("clicking the last offer left the window in the search")
	}
}

func TestEscapeOutOfTheSearchByTheKey(t *testing.T) {
	w := holding(newFake(), engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w = w.Key("/").Types("spec")
	if w.Screen() != OnSearch || w.looking.Reads() == "" {
		t.Fatalf("the search did not open: %v %q", w.Screen(), w.looking.Reads())
	}
	w = w.Key("esc")
	if w.Screen() != OnTheBoard {
		t.Fatalf("escape left the window on %v", w.Screen())
	}
	if w.looking.Reads() != "" {
		t.Fatalf("escape left the board narrowed to %q", w.looking.Reads())
	}
}
