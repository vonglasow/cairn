package gui

const (
	edgeWide = 5
	numberAt = 6
	titleAt  = 28
	labelsAt = 54
	lineHigh = 18
	titleHi  = 22
)

const tabWide = 140

const namesWide = 170

// Parts is where the pieces of a card go inside the box it was given.
type Parts struct {
	Panel, Edge, Number, Title, Labels Box
}

// PartsOf shares out a card's box.
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

// SubPanel is the body of one tile: its box less the gap that separates it from the next.
func SubPanel(b Box) Box {
	return Box{X: b.X, Y: b.Y, W: b.W, H: b.H - between}
}

// SubWriting is where a tile writes what it carries, which is the one line it has.
func SubWriting(b Box) Box {
	panel := SubPanel(b)
	row := Row(panel, panel.Y+(panel.H-lineHigh)/2, lineHigh)
	past := checkSide + inset + inset
	return Box{X: row.X + past, Y: row.Y, W: row.W - past, H: row.H}
}

const checkSide = 12

// StruckAt is the line through a step done: across the middle of the line its name is written on, from where the name begins and as long as it is.
func StruckAt(line Box, wide int) Box {
	return Box{X: line.X, Y: line.Y + line.H/2, W: wide, H: 1}
}

// TickAt is where the tick is written inside a filled box: a line of text set so the mark sits in the box's middle.
func TickAt(check Box) Box {
	return Box{X: check.X, Y: check.Y - 1, W: check.W + 2, H: lineHigh}
}

// SubCheck is the box before a step's name, ticked when it is done: a shape, read without comparing two greys.
func SubCheck(b Box) Box {
	panel := SubPanel(b)
	return Box{X: panel.X + padding, Y: panel.Y + (panel.H-checkSide)/2, W: checkSide, H: checkSide}
}

// CountedAt is where an objective's count sits on its number line: hard right, so that it and the number are read as two separate things rather than as one run of digits.
func CountedAt(number Box, said string) Box {
	wide := len([]rune(said)) * perGlyph
	return Box{X: number.X + number.W - wide, Y: number.Y, W: wide, H: number.H}
}

// Over is the panel a detail or a form is drawn in: the middle of the window, with the board still showing around it.
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

// Edges is the four sides of a box, each one row or one column thick.
func Edges(b Box) [4]Box {
	return [4]Box{
		{X: b.X, Y: b.Y, W: b.W, H: 1},
		{X: b.X, Y: b.Y + b.H - 1, W: b.W, H: 1},
		{X: b.X, Y: b.Y, W: 1, H: b.H},
		{X: b.X + b.W - 1, Y: b.Y, W: 1, H: b.H},
	}
}

const offerWide = 110

const detailOffers = 3

const contextNamed = 64

const searchWide = 76

const chipAt = 3

const inset = 4

const (
	fieldsPart = 5
	ofPanel    = 8
	barWide    = 10
	leastThumb = 24
)

func inSight(at, high, run, where int) int {
	top := at * high
	return min(top, max(where, top+high-run))
}

// Inside is the room for the text within that frame.
func Inside(b Box) Box {
	return Box{X: b.X + inset, Y: b.Y + inset, W: b.W - inset - inset, H: b.H - inset - inset}
}
