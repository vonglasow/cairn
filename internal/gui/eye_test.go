package gui

import (
	"cairn/internal/engine"
	"testing"
)

func TestACardInFocusHasAnEyeAfterItsNumber(t *testing.T) {
	number := Box{X: 100, Y: 200, W: 240, H: lineHigh}
	eye, ok := FocusEye(engine.Card{Number: 7, InFocus: true}, number, 30)
	if !ok {
		t.Fatal("a card in focus has no eye")
	}
	if eye.X != number.X+30+eyeSpace || eye.Y < number.Y || eye.Y+eye.H > number.Y+number.H || eye.W <= eye.H {
		t.Errorf("the eye is at %+v, after a number 30 wide at %+v", eye, number)
	}
	if _, ok := FocusEye(engine.Card{Number: 7}, number, 30); ok {
		t.Error("a card nobody is on has an eye")
	}
	if said := Numbered(engine.Card{Number: 7, InFocus: true}); said != "7" {
		t.Errorf("the number line reads %q", said)
	}
}

func TestTheEyeIsDrawn(t *testing.T) {
	on := engine.Card{Number: 1, Title: "Write the spec", State: "todo", InFocus: true}
	off := engine.Card{Number: 2, Title: "Ship it", State: "todo"}
	w := holding(newFake(), on, off)
	shot := drawn(t, w)
	l := w.Layout()
	if got := painted(shot, PartsOf(l.CardAt(1, 0, 0)).Number, Dark().Chosen); got == 0 {
		t.Error("the card in focus draws no eye")
	}
	if got := painted(shot, PartsOf(l.CardAt(1, 1, 0)).Number, Dark().Chosen); got != 0 {
		t.Errorf("the card nobody is on draws %d pixels of an eye", got)
	}
}
