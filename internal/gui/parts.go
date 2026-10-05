package gui

// Where every parcel of every panel goes.
//
// It is here and not in the drawing for a reason mutation found: arithmetic
// inside a drawing is arithmetic no assertion can judge. A title moved six
// pixels still looks like a title, and the only test that would notice is a
// pinned pixel frame — which goes red on another machine for a reason that is
// not a defect. Written as numbers, it is pinned like every other size.

// The room inside a card, and what each of its two lines takes.
const (
	edgeWide = 3  // the coloured strip down a card's left, for its priority
	numberAt = 6  // how far below a card's top its number sits
	titleAt  = 28 // and its title
	labelsAt = 54 // and the labels it carries, under the room the title has
	lineHigh = 18
	titleHi  = 22
)

// tabWide is how much of the strip one tab takes. Written in rather than
// measured: a tab is a word, and a word that changes length should not move the
// one beside it.
const tabWide = 140

// namesWide is how much of a form's row the name of a field takes, leaving the
// rest for what it holds.
const namesWide = 170

// Parts is where the pieces of a card go inside the box it was given.
type Parts struct {
	Panel, Edge, Number, Title, Labels Box
}

// PartsOf shares out a card's box. The gap at the bottom is what keeps two
// cards from touching; a card is CardHeight tall and draws in rather less.
func PartsOf(box Box) Parts {
	panel := Box{X: box.X, Y: box.Y, W: box.W, H: box.H - gap}
	inside := panel.W - padding - padding
	return Parts{
		Panel:  panel,
		Edge:   Box{X: panel.X, Y: panel.Y, W: edgeWide, H: panel.H},
		Number: Box{X: panel.X + padding, Y: panel.Y + numberAt, W: inside, H: lineHigh},
		Title:  Box{X: panel.X + padding, Y: panel.Y + titleAt, W: inside, H: titleHi},
		Labels: Box{X: panel.X + padding, Y: panel.Y + labelsAt, W: inside, H: lineHigh},
	}
}

// SubPanel is the body of one tile: its box less the gap that separates it
// from the next.
func SubPanel(b Box) Box {
	return Box{X: b.X, Y: b.Y, W: b.W, H: b.H - between}
}

// SubWriting is where a tile writes what it carries, which is the one line it
// has. Asked for by the tile and by the ghost of one in hand, so that what is
// carried is drawn where it would be drawn if it were put down — a ghost laid
// out any other way is a ghost that lies about where the thing will go.
func SubWriting(b Box) Box {
	panel := SubPanel(b)
	// Centred in the tile rather than pushed to its top: written an inset down
	// from a panel of twenty-two, a line of eighteen finished exactly on the
	// bottom edge. And after the box that says whether it is done.
	row := Row(panel, panel.Y+(panel.H-lineHigh)/2, lineHigh)
	past := checkSide + inset + inset
	return Box{X: row.X + past, Y: row.Y, W: row.W - past, H: row.H}
}

// checkSide is how big the box before a part is. Big enough to read as a box
// and not as a stray mark, small enough to sit inside a tile's line.
const checkSide = 12

// StruckAt is the line through a part done: across the middle of the line its
// name is written on, from where the name begins and as long as it is.
func StruckAt(line Box, wide int) Box {
	return Box{X: line.X, Y: line.Y + line.H/2, W: wide, H: 1}
}

// TickAt is where the tick is written inside a filled box: a line of text set so
// the mark sits in the box's middle. One point above the box's top, because a
// glyph's line carries room above the mark; three put it a touch too high,
// and two still did. And at the box's own edge across: a point in from it sat
// right of centre, a point out sat left.
func TickAt(check Box) Box {
	return Box{X: check.X, Y: check.Y - 1, W: check.W + 2, H: lineHigh}
}

// SubCheck is the box before a part's name, ticked when it is done: a shape,
// read without comparing two greys. That was all there was, and a part to do
// read as one finished with.
func SubCheck(b Box) Box {
	panel := SubPanel(b)
	return Box{X: panel.X + padding, Y: panel.Y + (panel.H-checkSide)/2, W: checkSide, H: checkSide}
}

// CountedAt is where an objective's count sits on its number line: hard right,
// so that it and the number are read as two separate things rather than as one
// run of digits.
func CountedAt(number Box, said string) Box {
	wide := len([]rune(said)) * perGlyph
	return Box{X: number.X + number.W - wide, Y: number.Y, W: wide, H: number.H}
}

// Over is the panel a detail or a form is drawn in: the middle of the window,
// with the board still showing around it.
//
// It takes rather more of the window than it used to. Once every field was
// drawn in a frame, a form of twelve fields ran off the bottom of the panel at
// the smallest window this board draws in — and a field off the bottom is a
// field nobody can fill in. Measured, not guessed: see the test that puts the
// last field of the longest form against what the panel offers.
func Over(width, height int) Box {
	return Box{X: width / 8, Y: height / 12, W: width * 3 / 4, H: height * 5 / 6}
}

// Row is one line inside a panel.
func Row(box Box, at, high int) Box {
	return Box{X: box.X + padding, Y: at, W: box.W - padding - padding, H: high}
}

// Message is where a window too small says so: across the middle of it.
func Message(width, height int) Box {
	return Box{X: margin, Y: height/2 - lineHigh, W: width - margin - margin, H: titleHi}
}

// Edges is the four sides of a box, each one row or one column thick. Written
// here rather than inside the drawing so the arithmetic that finds the far side
// of a box is arithmetic a test can pin.
func Edges(b Box) [4]Box {
	return [4]Box{
		{X: b.X, Y: b.Y, W: b.W, H: 1},
		{X: b.X, Y: b.Y + b.H - 1, W: b.W, H: 1},
		{X: b.X, Y: b.Y, W: 1, H: b.H},
		{X: b.X + b.W - 1, Y: b.Y, W: 1, H: b.H},
	}
}

// offerWide is how much of the bar one offer takes. Wide enough for the longest
// word the bar says, so an offer never moves because another changed length.
const offerWide = 110

// detailOffers is the most an open task offers at its foot: Done, Edit, Close.
const detailOffers = 3

// contextNamed is the room the word "context" takes before the place it names.
const contextNamed = 64

// searchWide is the room the word "Search" takes at the left of that box.
const searchWide = 76

// chipAt is how far a chip stands off the line it sits on.
const chipAt = 3

// inset is what a frame keeps between its edge and what it holds. A field is
// drawn in a box the way the search is drawn in a box: a place somebody types
// should look the same wherever it is, and a value floating on the panel's own
// ground looked like something written rather than something to write in.
const inset = 4

// The detail panel is two columns: what the task carries on the left, what was
// done to it on the right.
//
// Two columns and not one above the other, because the history is the part
// that never fits. Stacked, twelve framed fields take nearly the whole panel
// at the smallest window this board draws in, and the history gets what is
// left — which is nothing. Beside them it gets the panel's full height at
// every size.
const (
	// fieldsPart is how much of the panel's width the fields take.
	fieldsPart = 3
	ofPanel    = 5
	// barWide is the bar beside the history, which says how much of it is
	// showing and where in it somebody is.
	barWide = 10
	// leastThumb is the smallest the bar's grip is drawn. A grip in proportion
	// to a very long history would be a few pixels tall: too small to see, and
	// too small to take hold of.
	leastThumb = 24
)

// inSight moves a run so that a row is inside it, and answers where the run
// then stands.
//
// One arithmetic and not two: the columns and the objectives each had a copy,
// and the second copy of a boundary is where a boundary goes wrong unnoticed.
// Written without branches for a reason the mutation gate found: a run moved
// to exactly the row's top, and a row whose bottom lands exactly on the run's
// edge, both answer the same whichever branch is taken — so the two boundaries
// went out unjudged. As two clamps there are no boundaries to get wrong.
func inSight(at, high, run, where int) int {
	top := at * high
	return min(top, max(where, top+high-run))
}

// Inside is the room for the text within that frame.
func Inside(b Box) Box {
	return Box{X: b.X + inset, Y: b.Y + inset, W: b.W - inset - inset, H: b.H - inset - inset}
}
