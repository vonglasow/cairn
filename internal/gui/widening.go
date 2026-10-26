package gui

import "cairn/internal/engine"

// Widening is a gutter, or the edge above the band, being dragged: which one, where the pointer took it, the layout it was taken in and the sizes the board had then.
type Widening struct {
	On    bool
	Lane  bool
	Band  bool
	At    int
	From  Point
	Taken Layout
	Was   Sizes
}

func gutterAt(l Layout, was Sizes, p Point) (Widening, bool) {
	taken := Widening{On: true, From: p, Taken: l, Was: was}
	for at := range engine.Columns - 1 {
		if Gutter(l, at).In(p) {
			taken.At = at
			return taken, true
		}
	}
	for at := range len(l.Lane) - 1 {
		if LaneGutter(l, at).In(p) {
			taken.Lane, taken.At = true, at
			return taken, true
		}
	}
	if BandEdge(l).In(p) {
		taken.Band = true
		return taken, true
	}
	return Widening{}, false
}

// To is the sizes the board takes with the pointer at p: the columns, the lanes, or the band, as they were when the gutter was taken, with the pair beside it given what the pointer went; everything else as it was.
func (g Widening) To(p Point) Sizes {
	by := g.Was
	switch {
	case g.Band:
		by.Band = Raised(g.Taken.Column[1].H, g.Taken.Band.H, g.From.Y-p.Y)
	case g.Lane:
		copy(by.Lane[:], Widened(g.Taken.Lane[:], g.At, p.X-g.From.X))
	default:
		copy(by.Column[:], Widened(g.Taken.Column[:], g.At, p.X-g.From.X))
	}
	return by
}

// Evened is the sizes the board had with what this gutter shares given back: every column, or every lane, as wide as the others, or the band as tall as it opens.
func (g Widening) Evened() Sizes {
	by := g.Was
	switch {
	case g.Band:
		by.Band = [2]int{}
	case g.Lane:
		by.Lane = [3]int{}
	default:
		by.Column = [engine.Columns]int{}
	}
	return by
}
