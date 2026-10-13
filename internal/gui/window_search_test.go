package gui

import (
	"testing"

	"cairn/internal/engine"
	"cairn/test"
)

// Searching from the window: what a search narrows, how it is left, and what
// it reaches.

// A search narrows what is drawn to what it reaches, and reaches what a card
// does not show.
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

// Leaving a search puts everything back, because a search is a way of looking
// and not a way of changing.
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

// A search that reaches nothing draws nothing rather than everything: an empty
// answer is an answer.
func TestASearchThatReachesNothingShowsNothing(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "Write the spec", State: "backlog"})
	w = w.Key("/").Types("nothing like it")
	if got := len(w.Showing().Cards); got != 0 {
		t.Errorf("a search that reaches nothing shows %d tasks", got)
	}
}

// A search reaches objectives too, on the tab that shows them.
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

// Backspace in a search takes back one character of what is being looked for,
// and one character is one rune.
func TestBackspaceInASearchTakesBackOneCharacter(t *testing.T) {
	f := newFake()
	w := holding(f, engine.Card{Number: 1, Title: "réunion", State: "backlog"})

	w = w.Key("/").Types("réu")
	if got := w.Key("backspace").Looking(); got != "ré" {
		t.Fatalf("taking one back left %q", got)
	}
	if got := w.Key("backspace").Key("backspace").Looking(); got != "r" {
		t.Errorf("taking two back left %q", got)
	}
	// Taking back what is not there leaves nothing rather than failing.
	empty := w.Key("esc").Key("/")
	for range 3 {
		empty = empty.Key("backspace")
	}
	if got := empty.Looking(); got != "" {
		t.Errorf("an empty search reads %q", got)
	}
	// And backspace on the board itself is not a search at all.
	if got := w.Key("esc").Key("backspace").Looking(); got != "" {
		t.Errorf("backspace on the board left a search of %q", got)
	}
}

// A search can be left with the pointer. The box took the whole bar, so the
// one way out was a key — and the card the search had just found could not be
// acted on at all.
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

// Escape closes the search box and widens the board again, by the key as well
// as by the offer: the two are the same act and were written apart.
func TestEscapeOutOfTheSearchByTheKey(t *testing.T) {
	w := holding(newFake(), engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	w = w.Key("/").Types("spec")
	if w.Screen() != OnSearch || w.Looking() == "" {
		t.Fatalf("the search did not open: %v %q", w.Screen(), w.Looking())
	}
	w = w.Key("esc")
	if w.Screen() != OnTheBoard {
		t.Fatalf("escape left the window on %v", w.Screen())
	}
	if w.Looking() != "" {
		t.Fatalf("escape left the board narrowed to %q", w.Looking())
	}
}
