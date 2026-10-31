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
	for _, grab := range grabsOf(l) {
		if grab.Box.In(p) {
			taken := grab.Takes
			taken.On, taken.From, taken.Taken, taken.Was = true, p, l, was
			return taken, true
		}
	}
	return Widening{}, false
}

// Pointer is the shape the pointer takes over a point of the window.
type Pointer int

// The shapes the pointer takes: the plain arrow, sideways arrows for a gutter, up-and-down ones for the band's edge.
const (
	Arrow Pointer = iota
	ColResize
	RowResize
)

// Grab is a place a drag can take hold of, and the widening it starts there.
type Grab struct {
	Box   Box
	Takes Widening
}

func grabsOf(l Layout) []Grab {
	var grabs []Grab
	for at := range engine.Columns - 1 {
		grabs = append(grabs, Grab{Gutter(l, at), Widening{At: at}})
	}
	for at := range len(l.Lane) - 1 {
		grabs = append(grabs, Grab{LaneGutter(l, at), Widening{Lane: true, At: at}})
	}
	return append(grabs, Grab{BandEdge(l), Widening{Band: true}})
}

// Pointer is the shape the pointer takes while this gutter, or the band's edge, is held or hovered: up and down for the band, sideways for the rest.
func (g Widening) Pointer() Pointer {
	if g.Band {
		return RowResize
	}
	return ColResize
}

// Grabs are the places of the board a drag takes hold of: the gutters between columns and between lanes, and the band's edge.
func (w Window) Grabs() []Grab {
	if w.Blank() || w.tab != Tasks || w.screen != OnTheBoard {
		return nil
	}
	return grabsOf(w.Layout())
}

// PointerAt is the shape the pointer calls for at p: the one of the gutter or the edge being dragged wherever it is, otherwise the one of the grab it is over.
func (w Window) PointerAt(p Point) Pointer {
	if w.widening.On {
		return w.widening.Pointer()
	}
	for _, grab := range w.Grabs() {
		if grab.Box.In(p) {
			return grab.Takes.Pointer()
		}
	}
	return Arrow
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
