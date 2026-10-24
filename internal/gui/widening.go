package gui

import "cairn/internal/engine"

// Widening is a gutter being dragged: which one, where the pointer took it, and the layout it was taken in.
type Widening struct {
	On    bool
	Lane  bool
	At    int
	From  Point
	Taken Layout
}

func gutterAt(l Layout, p Point) (Widening, bool) {
	for at := range engine.Columns - 1 {
		if Gutter(l, at).In(p) {
			return Widening{On: true, At: at, From: p, Taken: l}, true
		}
	}
	for at := range len(l.Lane) - 1 {
		if LaneGutter(l, at).In(p) {
			return Widening{On: true, Lane: true, At: at, From: p, Taken: l}, true
		}
	}
	return Widening{}, false
}

// To is the widths the board takes with the pointer at p: the columns, or the lanes, as they were when the gutter was taken, with the pair beside it widened by how far the pointer went.
func (g Widening) To(p Point) Widths {
	var by Widths
	columns, lanes := g.Taken.Column[:], g.Taken.Lane[:]
	if g.Lane {
		copy(by.Lane[:], Widened(lanes, g.At, p.X-g.From.X))
		copy(by.Column[:], Widened(columns, 0, 0))
		return by
	}
	copy(by.Column[:], Widened(columns, g.At, p.X-g.From.X))
	copy(by.Lane[:], Widened(lanes, 0, 0))
	return by
}
