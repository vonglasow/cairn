package gui

import (
	"slices"
	"testing"

	"cairn/internal/engine"
)

func TestALabelSaysWhatCarriesIt(t *testing.T) {
	if got := LabelSays(engine.LabelCarried{Label: "sp26-20", Tasks: 2, Objectives: 1}); got != "sp26-20   2 tasks · 1 objective" {
		t.Errorf("a label says %q", got)
	}
}

func TestTheLabelsOfferWhatCanBeDoneToOne(t *testing.T) {
	w := Window{screen: OnLabels, labelsState: labelsState{carried: []engine.LabelCarried{{Label: "doc", Tasks: 1}}}}
	if offers := w.Offers(); !slices.Equal(offers, []string{"Rename", "Delete", "Cancel"}) {
		t.Errorf("the labels offer %v", offers)
	}
	w.renaming = true
	if offers := w.Offers(); !slices.Equal(offers, []string{"Take", "Cancel"}) {
		t.Errorf("a rename offers %v", offers)
	}
}
