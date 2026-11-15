package gui

import (
	"cairn/internal/board"
	"cairn/internal/engine/label"
	"slices"
	"testing"
)

func TestALabelSaysWhatCarriesIt(t *testing.T) {
	if got := LabelSays(label.Carried{Label: "sp26-20", Tasks: 2, Objectives: 1}); got != "sp26-20   2 tasks · 1 objective" {
		t.Errorf("a label says %q", got)
	}
}

func TestTheLabelsOfferWhatCanBeDoneToOne(t *testing.T) {
	w := Window{Board: board.Board{LabelsState: board.LabelsState{Carried: []label.Carried{{Label: "doc", Tasks: 1}}}, Screen: board.OnLabels}}
	if offers := w.Offers(); !slices.Equal(offers, []string{"New label", "Close"}) {
		t.Errorf("the labels offer %v", offers)
	}
	w.Seeking = true
	if offers := w.Offers(); !slices.Equal(offers, []string{"Cancel"}) {
		t.Errorf("a search that finds labels offers %v", offers)
	}
	w.Sought = board.Written{}.Types("release")
	if offers := w.Offers(); !slices.Equal(offers, []string{"Create", "Cancel"}) {
		t.Errorf("a search that finds none offers %v", offers)
	}
	w.Seeking, w.Renaming = false, true
	if offers := w.Offers(); !slices.Equal(offers, []string{"Save", "Cancel"}) {
		t.Errorf("a rename offers %v", offers)
	}
}
