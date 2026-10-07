package gui

import (
	"slices"
	"testing"

	"gioui.org/io/key"

	"cairn/internal/reading"
)

// gathered is a board with the cards given, the first of the second column
// chosen and the others ⌘-clicked in, in the order given.
func gathered(cards []reading.Card, at ...Point) Window {
	w := holding(newFake(), cards...)
	w = w.clicked(at[0])
	for _, p := range at[1:] {
		w = w.KeysDown(key.ModShortcut).clicked(p).KeysDown(0)
	}
	return w
}

// A card ⌘-clicked a second time comes back out of what is taken, the first
// one taken as much as any other; and esc lets go of what is taken before it
// clears a search.
func TestWhatIsTakenIsLetGoOf(t *testing.T) {
	cards := []reading.Card{card(1, "First", "todo"), card(2, "Second", "todo"), card(3, "Third", "todo")}
	w := holding(newFake(), cards...)
	w = gathered(cards, middleOf(w, 1, 0), middleOf(w, 1, 1), middleOf(w, 1, 2))
	if got := w.Taken(); !slices.Equal(got, []int{1, 2, 3}) {
		t.Fatalf("three gathered are %v", got)
	}
	w = w.KeysDown(key.ModShortcut).clicked(middleOf(w, 1, 0)).KeysDown(0)
	if got := w.Taken(); !slices.Equal(got, []int{2, 3}) {
		t.Errorf("the first ⌘-clicked again left %v", got)
	}
	w.looking = Written{Value: "spec"}
	w = w.Key("esc")
	if len(w.taken) != 0 || w.looking.Value != "spec" {
		t.Errorf("esc left %v taken and the search at %q", w.taken, w.looking.Value)
	}
	if w = w.Key("esc"); w.looking.Value != "" {
		t.Errorf("esc with nothing taken left the search at %q", w.looking.Value)
	}
	// And with no search to clear, esc still lets go of what is taken.
	w = gathered(cards, middleOf(w, 1, 0), middleOf(w, 1, 1))
	if w = w.Key("esc"); len(w.taken) != 0 {
		t.Errorf("esc with no search left %v taken", w.taken)
	}
	// A stretch runs upward as well as down.
	w = w.clicked(middleOf(w, 1, 2)).KeysDown(key.ModShift).clicked(middleOf(w, 1, 0)).KeysDown(0)
	if got := w.Taken(); !slices.Equal(got, []int{1, 2, 3}) {
		t.Errorf("a stretch upward took %v", got)
	}
}

// ⇧ stretches within one column, and across two there is no between: the
// click gathers instead. On the objectives page what is chosen stays an
// objective.
func TestAStretchStaysInItsColumn(t *testing.T) {
	cards := []reading.Card{card(1, "First", "todo"), card(2, "Second", "todo"), card(3, "Elsewhere", "backlog")}
	w := holding(newFake(), cards...)
	w = w.clicked(middleOf(w, 1, 0)).KeysDown(key.ModShift).clicked(middleOf(w, 0, 0)).KeysDown(0)
	if got := w.Taken(); !slices.Equal(got, []int{1, 3}) || w.ChosenGoal() {
		t.Errorf("a stretch across two columns took %v, chosen an objective: %v", got, w.ChosenGoal())
	}
	goals := New(newFake(), 1200, 800)
	goals.snap = reading.Snapshot{Goals: []reading.Goal{{Number: 1, Title: "One", State: "active"}, {Number: 2, Title: "Two", State: "active"}}}
	goals.tab = Objectives
	goals = goals.clicked(goalMiddle(goals, 1)).KeysDown(key.ModShift).clicked(goalMiddle(goals, 2)).KeysDown(0)
	if got := goals.Taken(); !slices.Equal(got, []int{1, 2}) || !goals.ChosenGoal() {
		t.Errorf("a stretch over the objectives took %v, chosen an objective: %v", got, goals.ChosenGoal())
	}
}

// Archiving several names what is already off the board, and leaves it out:
// the rest goes in one act.
func TestArchivingSeveralLeavesOutWhatIsAlreadyOff(t *testing.T) {
	f := newFake()
	cards := []reading.Card{card(1, "First", "todo"), card(2, "Second", "todo"), {Number: 3, Title: "Gone", State: "todo", Archived: true}}
	w := holding(f, cards...)
	w.archived = true
	w = w.clicked(middleOf(w, 1, 0))
	for _, at := range []int{1, 2} {
		w = w.KeysDown(key.ModShortcut).clicked(middleOf(w, 1, at)).KeysDown(0)
	}
	asked := w.Key("x")
	if got := Heads(asked); got != "Archive 3 tasks?" {
		t.Errorf("the question is headed %q", got)
	}
	if offers := asked.Offers(); slices.Contains(offers, "Skip flagged") {
		t.Errorf("only what cannot go was named, and the question offers %v", offers)
	}
	asked.Choose("All possible")
	in, ok := f.gave("task", "archive")
	if !ok || in.Given["number"] != "1" || !slices.Equal(in.Listed["also"], []string{"2"}) {
		t.Errorf("archived %+v", in)
	}
}
