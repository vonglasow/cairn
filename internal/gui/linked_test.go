package gui

import (
	"cairn/internal/board"
	"reflect"
	"testing"
)

func TestOOpensANewLinkedTaskOnAChosenTaskOnly(t *testing.T) {
	w := holding(newFake(), card(1, "Ship it", "done"))
	if got := w.Key("o"); got.Screen != board.OnTheBoard {
		t.Errorf("o with nothing chosen left the window on %v", got.Screen)
	}
	w.Chosen, w.ChosenGoal = 1, true
	if got := w.Key("o"); got.Screen != board.OnTheBoard {
		t.Errorf("o on an objective left the window on %v", got.Screen)
	}
	w.ChosenGoal = false
	got := w.Key("o")
	if got.Screen != board.OnForm || !reflect.DeepEqual(got.Form.Fields, board.LinkedTaskFields(1)) {
		t.Errorf("o on a task left the window on %v with %+v", got.Screen, got.Form.Fields)
	}
}
