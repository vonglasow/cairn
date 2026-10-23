package gui

import "testing"

func TestColumnName_spansItsHeadClearOfItsTopEdge(t *testing.T) {
	head := Box{X: 100, Y: 40, W: 200, H: 28}
	got := ColumnName(head)
	if got.X != head.X || got.W != head.W {
		t.Errorf("the name runs from %d over %d, want the head's width", got.X, got.W)
	}
	if got.Y < head.Y+nameDrop || got.Y+got.H > head.Y+head.H {
		t.Errorf("the name stands from %d to %d in a head from %d to %d", got.Y, got.Y+got.H, head.Y, head.Y+head.H)
	}
}

func TestShifted_movesAPanelAndHoldsItInsideTheWindow(t *testing.T) {
	panel := Box{X: 100, Y: 60, W: 600, H: 300}
	if got := Shifted(panel, Point{X: 80, Y: 40}, 1200, 800); got != (Box{X: 180, Y: 100, W: 600, H: 300}) {
		t.Errorf("moved by 80 and 40, the panel stands at %v", got)
	}
	if got := Shifted(panel, Point{X: 5000, Y: -5000}, 1200, 800); got != (Box{X: 600, Y: 0, W: 600, H: 300}) {
		t.Errorf("moved far, the panel stands at %v, want held inside the window", got)
	}
}
