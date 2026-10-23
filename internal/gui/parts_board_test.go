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
