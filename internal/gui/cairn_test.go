package gui

import (
	"testing"
	"time"

	"gioui.org/io/pointer"
)

// At the start every stone is above the window; at the end they stand on one
// another in the middle of it, the base first and widest.
func TestTheStonesFallIntoACairn(t *testing.T) {
	for i, stone := range StonesAt(1200, 800, 0) {
		if stone.Y+stone.H > 0 {
			t.Errorf("stone %d starts at %+v, inside the window", i, stone)
		}
	}
	built := StonesAt(1200, 800, 1)
	if base := built[0]; base != (Box{X: 520, Y: 436, W: 160, H: 44}) {
		t.Errorf("the base stands at %+v", base)
	}
	for i := 1; i < len(built); i++ {
		if built[i].Y+built[i].H != built[i-1].Y || built[i].X+built[i].W/2 != 600 {
			t.Errorf("stone %d, %+v, does not stand centred on the one below, %+v", i, built[i], built[i-1])
		}
	}
	// A stone slows as it lands: half its fall's time takes it three
	// quarters of the way — from 44 above the window to 436, 316.
	if got := StonesAt(1200, 800, 0.2)[0].Y; got != 316 {
		t.Errorf("half way through its fall the base is at %d, want 316", got)
	}
	// Half way, the base has landed further than the top stone has.
	half := StonesAt(1200, 800, 0.5)
	if half[0].Y <= half[4].Y {
		t.Errorf("half way, the base is at %d and the top stone at %d", half[0].Y, half[4].Y)
	}
}

// The cairn is built from the first frame, for as long as it lasts, and is
// ended by a press.
func TestTheCairnLastsABreath(t *testing.T) {
	start := time.Date(2026, 10, 12, 9, 0, 0, 0, time.UTC)
	if _, on := (Window{}).Intro(); on {
		t.Error("a window never drawn builds a cairn")
	}
	w := Window{}.Born(start)
	if progress, on := w.Intro(); !on || progress != 0 {
		t.Errorf("on its first frame the cairn is %v, %v", progress, on)
	}
	if progress, on := w.At(start.Add(IntroLasts() / 2)).Intro(); !on || progress != 0.5 {
		t.Errorf("half way through it is %v, %v", progress, on)
	}
	if _, on := w.At(start.Add(IntroLasts())).Intro(); on {
		t.Error("the cairn is still being built once its time is up")
	}
	if again := w.At(start.Add(time.Hour)).Born(start.Add(time.Hour)); again.born != start {
		t.Error("a later frame was taken for a birth")
	}
	pressed, _ := Pointing(w, Gesture{}, pointer.Press, Point{X: 1, Y: 1}, 0, start)
	if _, on := pressed.Intro(); on {
		t.Error("a press left the cairn building")
	}
	moved, _ := Pointing(w, Gesture{}, pointer.Move, Point{X: 1, Y: 1}, 0, start)
	if _, on := moved.Intro(); !on {
		t.Error("a pointer passing over ended the cairn")
	}
}

// While it is built the stones are drawn over the board.
func TestTheCairnIsDrawn(t *testing.T) {
	start := time.Date(2026, 10, 12, 9, 0, 0, 0, time.UTC)
	w := New(newFake(), 1200, 800).Born(start).At(start.Add(IntroLasts() * 9 / 10))
	stone := StonesAt(1200, 800, 0.9)[0]
	if got := ink(drawn(t, w), Box{X: stone.X + stone.W/2, Y: stone.Y + stone.H/2, W: 1, H: 1}, Dark().Back); got == 0 {
		t.Error("the base stone is not drawn")
	}
}
