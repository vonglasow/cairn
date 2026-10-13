package gui

import (
	"testing"

	"cairn/internal/reading"
)

// A card in focus says so with an eye before its number, drawn rather than
// written: a dot read as nothing in particular. The number moves along to
// make room for it, and a card nobody is on has neither.
func TestACardInFocusHasAnEyeBeforeItsNumber(t *testing.T) {
	number := Box{X: 100, Y: 200, W: 240, H: lineHigh}
	eye, ok := FocusEye(reading.Card{Number: 7, InFocus: true}, number)
	if !ok {
		t.Fatal("a card in focus has no eye")
	}
	if eye.X != number.X || eye.Y < number.Y || eye.Y+eye.H > number.Y+number.H || eye.W <= eye.H {
		t.Errorf("the eye is at %+v, in a number line at %+v", eye, number)
	}
	if moved := NumberText(reading.Card{Number: 7, InFocus: true}, number); moved.X <= eye.X+eye.W || moved.X+moved.W != number.X+number.W {
		t.Errorf("beside the eye the number is written at %+v", moved)
	}
	if _, ok := FocusEye(reading.Card{Number: 7}, number); ok {
		t.Error("a card nobody is on has an eye")
	}
	if kept := NumberText(reading.Card{Number: 7}, number); kept != number {
		t.Errorf("without an eye the number moved to %+v", kept)
	}
	if said := Numbered(reading.Card{Number: 7, InFocus: true}); said != "7" {
		t.Errorf("the number line reads %q", said)
	}
}

// The eye is drawn, in the colour that says chosen, and nothing is drawn
// where a card nobody is on would have one.
func TestTheEyeIsDrawn(t *testing.T) {
	on := reading.Card{Number: 1, Title: "Write the spec", State: "todo", InFocus: true}
	off := reading.Card{Number: 2, Title: "Ship it", State: "todo"}
	w := holding(newFake(), on, off)
	shot := drawn(t, w)
	l := w.Layout()
	eye, _ := FocusEye(on, PartsOf(l.CardAt(1, 0, 0)).Number)
	if got := painted(shot, eye, Dark().Chosen); got == 0 {
		t.Error("the card in focus draws no eye")
	}
	where, _ := FocusEye(reading.Card{InFocus: true}, PartsOf(l.CardAt(1, 1, 0)).Number)
	if got := painted(shot, where, Dark().Chosen); got != 0 {
		t.Errorf("the card nobody is on draws %d pixels of an eye", got)
	}
}
