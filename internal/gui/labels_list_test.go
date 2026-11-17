package gui

import (
	"cairn/internal/board"
	"cairn/internal/gui/geom"
	"slices"
	"testing"
)

func labelsOpen(t *testing.T, carried ...[]string) Window {
	t.Helper()
	w := New(newFake(), 1200, 800)
	for i, labels := range carried {
		w.Snap.Cards = append(w.Snap.Cards, board.Card{Number: i + 1, Title: "Task", State: "todo", Labels: labels})
	}
	if len(carried) == 0 {
		w.Snap.Cards = []board.Card{{Number: 1, Title: "Task", State: "todo"}}
	}
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")
	for w.Form.Fields[w.Form.On()].Arg != "label" {
		w = w.Key("tab")
	}
	w = w.Key("enter")
	if w.Screen != board.OnChoosing {
		t.Fatalf("the labels field opened %v, not the list", w.Screen)
	}
	return w
}

func TestLabels_aRowSayingWhyTakesNothing(t *testing.T) {
	w := labelsOpen(t, nil, []string{"ado"}).Types("ado:1").Key("enter")
	if w.Ticked(1) {
		t.Error("enter on the row saying why ticked a label")
	}
	if len(w.LabelsOffered) != 1 {
		t.Errorf("enter on the row saying why created %v", w.LabelsOffered)
	}
}

func TestLabels_spaceTicksALabelAndEnterIsDone(t *testing.T) {
	w := labelsOpen(t, nil, []string{"ado"})
	if w = w.Key("space").Edits(0, 0, " "); !w.Ticked(1) || w.Filtering.Reads() != "" {
		t.Fatalf("space ticked %v and typed %q", w.Ticked(1), w.Filtering.Reads())
	}
	if w = w.Key("space"); w.Ticked(1) {
		t.Error("space again did not untick it")
	}
	if !slices.Equal(w.Offers(), []string{"Done", "Cancel"}) || w.shortcut("Done") != "enter" {
		t.Errorf("the list offers %v, Done on %q", w.Offers(), w.shortcut("Done"))
	}
	if w = w.Key("space").Key("enter"); w.Screen != board.OnForm || w.Form.Writing().Reads() != "ado" {
		t.Errorf("enter left %v with %q", w.Screen, w.Form.Writing().Reads())
	}
}

func TestLabels_withWordsTypedSpaceTypesAndEnterTakesTheLineFound(t *testing.T) {
	w := labelsOpen(t, nil, []string{"ado"})
	if typed := w.Types("a").Edits(1, 1, " "); typed.Filtering.Reads() != "a " {
		t.Errorf("a space after a word typed reads %q", typed.Filtering.Reads())
	}
	w = labelsOpen(t, nil, []string{"ado"})
	if w = w.Types("ado").Key("enter"); w.Screen != board.OnChoosing || !w.Ticked(1) {
		t.Errorf("enter on a label typed left %v, ticked %v", w.Screen, w.Ticked(1))
	}
}

func TestLabels_spaceOnNothingDoesNothing(t *testing.T) {
	w := labelsOpen(t)
	if got := w.Key("space"); got.Screen != board.OnChoosing || len(got.LabelsOffered) != 0 {
		t.Errorf("space on an empty list left %v offering %v", got.Screen, got.LabelsOffered)
	}
}

func TestLabels_aClickTicksALabelAndTheListStaysOpenForMore(t *testing.T) {
	w := labelsOpen(t, nil, []string{"ado"}, []string{"bug"})
	click := func(w Window, title string) Window {
		box := w.ChoosingBox()
		for i, one := range w.Choosing() {
			if one.Title == title {
				row := ChoosingRow(box, ChoosingFrom(box)+(i-w.ChoosingFromRow())*apart)
				return w.Click(geom.Point{X: row.X + row.W/2, Y: row.Y + row.H/2})
			}
		}
		t.Fatalf("the list offers no %q", title)
		return w
	}
	w = click(click(w, "ado"), "bug")
	if w.Screen != board.OnChoosing || !w.Ticked(1) || !w.Ticked(2) {
		t.Fatalf("two clicks leave %v, ticked %v", w.Screen, w.TickedSet)
	}
	if w = click(w, "ado"); w.Ticked(1) {
		t.Error("a second click on a label did not untick it")
	}
	if kept := w.clickedOff(geom.Point{X: 1, Y: 1}); kept.Form.Writing().Reads() != "bug" {
		t.Errorf("a click off the list kept %q", kept.Form.Writing().Reads())
	}
}
