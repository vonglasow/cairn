package gui

import (
	"cairn/internal/board"
	"cairn/internal/gui/geom"
	"slices"
	"testing"

	"gioui.org/io/key"
)

func gathered(cards []board.Card, at ...geom.Point) Window {
	w := holding(newFake(), cards...)
	w = w.clicked(at[0])
	for _, p := range at[1:] {
		w = w.KeysDown(key.ModShortcut).clicked(p).KeysDown(0)
	}
	return w
}

func TestWhatIsTakenIsLetGoOf(t *testing.T) {
	cards := []board.Card{card(1, "First", "todo"), card(2, "Second", "todo"), card(3, "Third", "todo")}
	w := holding(newFake(), cards...)
	w = gathered(cards, middleOf(w, 1, 0), middleOf(w, 1, 1), middleOf(w, 1, 2))
	if got := w.Taken(); !slices.Equal(got, []int{1, 2, 3}) {
		t.Fatalf("three gathered are %v", got)
	}
	w = w.KeysDown(key.ModShortcut).clicked(middleOf(w, 1, 0)).KeysDown(0)
	if got := w.Taken(); !slices.Equal(got, []int{2, 3}) {
		t.Errorf("the first ⌘-clicked again left %v", got)
	}
	w.Looking = board.Written{Value: "spec"}
	w = w.Key("esc")
	if len(w.Gathered) != 0 || w.Looking.Value != "spec" {
		t.Errorf("esc left %v taken and the search at %q", w.Gathered, w.Looking.Value)
	}
	if w = w.Key("esc"); w.Looking.Value != "" {
		t.Errorf("esc with nothing taken left the search at %q", w.Looking.Value)
	}
	w = gathered(cards, middleOf(w, 1, 0), middleOf(w, 1, 1))
	if w = w.Key("esc"); len(w.Gathered) != 0 {
		t.Errorf("esc with no search left %v taken", w.Gathered)
	}
	w = w.clicked(middleOf(w, 1, 2)).KeysDown(key.ModShift).clicked(middleOf(w, 1, 0)).KeysDown(0)
	if got := w.Taken(); !slices.Equal(got, []int{1, 2, 3}) {
		t.Errorf("a stretch upward took %v", got)
	}
}

func TestAStretchStaysInItsColumn(t *testing.T) {
	cards := []board.Card{card(1, "First", "todo"), card(2, "Second", "todo"), card(3, "Elsewhere", "backlog")}
	w := holding(newFake(), cards...)
	w = w.clicked(middleOf(w, 1, 0)).KeysDown(key.ModShift).clicked(middleOf(w, 0, 0)).KeysDown(0)
	if got := w.Taken(); !slices.Equal(got, []int{1, 3}) || w.ChosenGoal {
		t.Errorf("a stretch across two columns took %v, chosen an objective: %v", got, w.ChosenGoal)
	}
	goals := New(newFake(), 1200, 800)
	goals.Snap = board.Snapshot{Goals: []board.Goal{{Number: 1, Title: "One", State: "active"}, {Number: 2, Title: "Two", State: "active"}}}
	goals.Tab = board.Objectives
	goals = goals.clicked(goalMiddle(goals, 1)).KeysDown(key.ModShift).clicked(goalMiddle(goals, 2)).KeysDown(0)
	if got := goals.Taken(); !slices.Equal(got, []int{1, 2}) || !goals.ChosenGoal {
		t.Errorf("a stretch over the objectives took %v, chosen an objective: %v", got, goals.ChosenGoal)
	}
}

func TestArchivingSeveralLeavesOutWhatIsAlreadyOff(t *testing.T) {
	f := newFake()
	cards := []board.Card{card(1, "First", "todo"), card(2, "Second", "todo"), {Number: 3, Title: "Gone", State: "todo", Archived: true}}
	w := holding(f, cards...)
	w.Archived = true
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
	asked.Choose("Archive")
	in, ok := f.gave("task", "archive")
	if !ok || in.Given["number"] != "1" || !slices.Equal(in.Listed["also"], []string{"2"}) {
		t.Errorf("archived %+v", in)
	}
}

func TestTheMenuOfGatheredCardsActsOnAllOfThem(t *testing.T) {
	f := newFake()
	cards := []board.Card{card(1, "First", "todo"), card(2, "Second", "todo"), card(3, "Third", "todo")}
	w := holding(f, cards...)
	w = w.clicked(middleOf(w, 1, 0)).KeysDown(key.ModShortcut).clicked(middleOf(w, 1, 1)).KeysDown(0)
	opened := w.ContextAt(middleOf(w, 1, 1))
	want := []string{"Move to…", "Priority…", "Labels…", "Focus", "Archive", "Delete"}
	if !slices.Equal(opened.Taken(), []int{1, 2}) || !slices.Equal(opened.Menu(), want) || opened.MenuHead() != "2 tasks · #1 #2" {
		t.Fatalf("the menu of two gathered: taken %v, %v, %q", opened.Taken(), opened.Menu(), opened.MenuHead())
	}
	line := func(w Window, said string) Window { return w.fromMenu(slices.Index(w.Menu(), said)) }
	line(line(opened, "Priority…"), "high")
	if in, ok := f.gave("task", "edit"); !ok || in.Given["number"] != "1" || !slices.Equal(in.Listed["also"], []string{"2"}) || in.Given["priority"] != "high" {
		t.Errorf("Priority… high on two asked %+v", in)
	}
	line(opened, "Focus")
	if in, ok := f.gave("task", "focus"); !ok || !slices.Equal(in.Listed["also"], []string{"2"}) {
		t.Errorf("Focus on two asked %+v", in)
	}
	if asking := line(opened, "Archive"); asking.Question != board.AboutArchivingSeveral {
		t.Errorf("Archive on two asks %v", asking.Question)
	}
	if apart := w.ContextAt(middleOf(w, 1, 2)); len(apart.Taken()) != 0 {
		t.Errorf("a right-click outside the gathering kept %v", apart.Taken())
	}
}

func TestTheLabelsOfGatheredCardsSayWhatAllAndSomeCarry(t *testing.T) {
	f := newFake()
	cards := []board.Card{{Number: 1, Title: "First", State: "todo", Labels: []string{"bug", "ui"}}, {Number: 2, Title: "Second", State: "todo", Labels: []string{"bug"}}}
	w := holding(f, cards...)
	w = w.clicked(middleOf(w, 1, 0)).KeysDown(key.ModShortcut).clicked(middleOf(w, 1, 1)).KeysDown(0)
	w = w.ContextAt(middleOf(w, 1, 1))
	w = w.fromMenu(slices.Index(w.Menu(), "Labels…"))
	at := map[string]int{}
	for i, one := range w.Choosing() {
		at[one.Title] = i
	}
	bug, ui := w.Choosing()[at["bug"]].Number, w.Choosing()[at["ui"]].Number
	if !w.Ticked(bug) || w.Partly(bug) || w.Ticked(ui) || !w.Partly(ui) {
		t.Fatalf("bug ticked %v partly %v, ui ticked %v partly %v", w.Ticked(bug), w.Partly(bug), w.Ticked(ui), w.Partly(ui))
	}
	if !slices.Equal(w.Offers(), []string{"Done"}) {
		t.Errorf("the labels written at each click offer %v", w.Offers())
	}
	w.ChoosingOn = at["ui"]
	w = w.Key("space")
	if in, ok := f.gave("task", "edit"); !ok || in.Given["number"] != "1" || !slices.Equal(in.Listed["also"], []string{"2"}) || !slices.Equal(in.Listed["label"], []string{"+ui"}) {
		t.Errorf("a click on a label some carry asked %+v", in)
	}
	if !w.Ticked(ui) || w.Partly(ui) {
		t.Errorf("ui after the click: ticked %v partly %v", w.Ticked(ui), w.Partly(ui))
	}
}
