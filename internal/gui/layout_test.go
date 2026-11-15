package gui

import (
	"cairn/internal/board"
	"cairn/internal/engine"
	"cairn/internal/gui/geom"
	"testing"
)

func TestEveryFieldOfAFormFitsAtTheLeastSize(t *testing.T) {
	w := holding(newFake(), board.Card{Number: 1, Title: "Write the spec", State: "todo", Steps: []engine.Part{{Number: 1, Title: "Draft"}}})
	w = w.Resized(geom.Narrowest, geom.Shortest).clicked(middleOf(w.Resized(geom.Narrowest, geom.Shortest), 1, 0)).Key("enter").Key("e")
	if w.Screen != board.OnForm {
		t.Fatalf("the form did not open at the least size: %v", w.Screen)
	}
	panel := w.formPanel()
	for at := range w.Form.Fields {
		if _, value, _ := FormPlace(w.formBox(), w.Form.Fields, at); value.Y+value.H > panel.Y+panel.H {
			t.Errorf("the field %q ends at %d, under the panel's foot at %d", w.Form.Fields[at].Name, value.Y+value.H, panel.Y+panel.H)
		}
	}
}
