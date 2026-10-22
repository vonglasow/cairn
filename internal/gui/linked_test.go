package gui

import (
	"reflect"
	"testing"

	"cairn/internal/engine"
)

func TestStartingANewTaskCanCarryALink(t *testing.T) {
	w := holding(newFake(), card(1, "Ship it", "done"))
	w.chosen = 1
	got := w.startingLinked()
	if got.Screen() != OnForm || !got.form.Adding {
		t.Fatalf("the window is on %v, adding %v", got.Screen(), got.form.Adding)
	}
	if want := engine.LinkedTaskFields(1); !reflect.DeepEqual(got.form.Fields(), want) {
		t.Errorf("the form asks %+v, want %+v", got.form.Fields(), want)
	}
	if got.form.Changed() {
		t.Error("the preset link counts as a change")
	}
}

func TestOOpensANewLinkedTaskOnAChosenTaskOnly(t *testing.T) {
	w := holding(newFake(), card(1, "Ship it", "done"))
	if got := w.Key("o"); got.Screen() != OnTheBoard {
		t.Errorf("o with nothing chosen left the window on %v", got.Screen())
	}
	w.chosen, w.chosenGoal = 1, true
	if got := w.Key("o"); got.Screen() != OnTheBoard {
		t.Errorf("o on an objective left the window on %v", got.Screen())
	}
	w.chosenGoal = false
	got := w.Key("o")
	if got.Screen() != OnForm || !reflect.DeepEqual(got.form.Fields(), engine.LinkedTaskFields(1)) {
		t.Errorf("o on a task left the window on %v with %+v", got.Screen(), got.form.Fields())
	}
}
