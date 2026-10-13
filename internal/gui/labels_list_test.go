package gui

import (
	"cairn/internal/engine"
	"testing"
)

// labelsOpen opens the list of labels over the form of task 1, in a window
// whose tasks carry the labels given.
func labelsOpen(t *testing.T, carried ...[]string) Window {
	t.Helper()
	w := New(newFake(), 1200, 800)
	for i, labels := range carried {
		w.snap.Cards = append(w.snap.Cards, engine.Card{Number: i + 1, Title: "Task", State: "todo", Labels: labels})
	}
	if len(carried) == 0 {
		w.snap.Cards = []engine.Card{{Number: 1, Title: "Task", State: "todo"}}
	}
	w = w.clicked(middleOf(w, 1, 0)).Choose("Edit")
	for w.Form().Fields()[w.Form().On()].Arg != "label" {
		w = w.Key("tab")
	}
	w = w.Key("enter")
	if w.Screen() != OnChoosing {
		t.Fatalf("the labels field opened %v, not the list", w.Screen())
	}
	return w
}

func TestLabels_aRowSayingWhyTakesNothing(t *testing.T) {
	w := labelsOpen(t, nil, []string{"ado"}).Types("ado:1").Key("enter")
	if w.Ticked(1) {
		t.Error("enter on the row saying why ticked a label")
	}
	if said := w.RowSays(w.Choosing()[0]); said[:1] != " " {
		t.Errorf("the row saying why reads %q once enter was pressed on it, as if taken", said)
	}
	if len(w.labelsOffered) != 1 {
		t.Errorf("enter on the row saying why created %v", w.labelsOffered)
	}
}

func TestLabels_theCreatingRowCreatesTickedAndClearsTheLine(t *testing.T) {
	w := labelsOpen(t, nil, []string{"ado"}).Types("Front End").Key("enter")
	if len(w.labelsOffered) != 2 || w.labelsOffered[1] != "front-end" || !w.Ticked(2) {
		t.Errorf("after creating, the list offers %v with %v ticked", w.labelsOffered, w.ticked)
	}
	if w.Filtering().Reads() != "" {
		t.Errorf("the line still reads %q", w.Filtering().Reads())
	}
}

func TestLabels_enterTicksALabelAndUnticksIt(t *testing.T) {
	w := labelsOpen(t, nil, []string{"ado"})
	if w = w.Key("enter"); !w.Ticked(1) {
		t.Fatal("enter did not tick the label")
	}
	if w = w.Key("enter"); w.Ticked(1) {
		t.Error("enter again did not untick it")
	}
}

func TestLabels_enterOnNothingDoesNothing(t *testing.T) {
	w := labelsOpen(t)
	if got := w.Key("enter"); got.Screen() != OnChoosing || len(got.labelsOffered) != 0 {
		t.Errorf("enter on an empty list left %v offering %v", got.Screen(), got.labelsOffered)
	}
}

func TestLabels_anEmptyListSaysHowToBegin(t *testing.T) {
	if got := labelsOpen(t).EmptyListSays(); got != LabelsNoneSays {
		t.Errorf("an empty list of labels says %q", got)
	}
	if got := New(newFake(), 1200, 800).EmptyListSays(); got != ChoosingSays {
		t.Errorf("an empty list of objectives says %q", got)
	}
}
