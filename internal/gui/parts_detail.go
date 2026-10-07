package gui

// Where the parts of an open task or objective go: its fields, what it
// serves, its description and its history.

// PanelOfferAt is where one of a panel's own offers sits: along its bottom,
// each in its own place so a pointer can land on one of them.
//
// A panel draws what can be done on it, and what is drawn has to work where it
// is drawn — words that look like buttons and answer only somewhere else are
// worse than no words at all.
func PanelOfferAt(box Box, at int) Box {
	// As tall as the bar's, so the screen that asks before losing work does
	// not hold the smallest targets drawn. It grows upward: the foot of the
	// panel is where it has always been.
	high := barHigh
	return Box{
		X: box.X + padding + at*offerWide,
		Y: box.Y + box.H - padding - high,
		W: offerWide,
		H: high,
	}
}

// AskingOver is a box asking a question of a few lines: centred across the
// window, a third of the way down, and as tall as its heading, its lines and
// its offers.
func AskingOver(width, height, lines int) Box {
	top := FirstRow(Box{})
	return Box{X: (width - askingWide) / 2, Y: height / 4, W: askingWide, H: top + lines*apart + barHigh + padding}
}

// askingWide is what a question takes across: room for a sentence of sixty
// characters and the two offers under it, and less than the narrowest window
// this board draws in.
const askingWide = 560

// FirstRow is where the first line inside a panel begins: under its heading.
func FirstRow(box Box) int {
	ruled := Ruled(box)
	return ruled.Y + ruled.H + padding
}

// ContextName is where an open task names what the line beside it says: at its
// foot, past the last place an offer can take. Not among the fields — the
// place is not something the task carries, and a thirteenth field took a row
// from the box of its parts on a window somebody works in, measured: three
// tiles down to two at 1200 by 800.
func ContextName(box Box) Box {
	after := PanelOfferAt(box, detailOffers)
	return Box{X: after.X + padding, Y: after.Y + (after.H-lineHigh)/2, W: contextNamed, H: lineHigh}
}

// ContextPath is where it says where what was written about it elsewhere
// lives: from the end of its name to the panel's edge.
func ContextPath(box Box) Box {
	name := ContextName(box)
	from := name.X + name.W
	return Box{X: from, Y: name.Y, W: box.X + box.W - padding - from, H: lineHigh}
}

// Chip is the patch drawn behind a key so it reads as a key and not as a word.
// It stands a little off the line it is on, because a chip flush with the text
// is a chip nobody sees.
func Chip(name Box) Box {
	return Box{X: name.X - chipAt, Y: name.Y - chipAt, W: name.W, H: name.H + chipAt + chipAt}
}

// Heading is where a panel says what it is, above everything it holds. Every
// panel has one: a panel that opens with no name is a panel somebody has to
// read through before knowing what they are looking at.
func Heading(box Box) Box {
	return Box{X: box.X + padding, Y: box.Y + padding, W: box.W - padding - padding, H: titleHi}
}

// Ruled is the line under that heading, which is what makes it a heading
// rather than the first row.
func Ruled(box Box) Box {
	head := Heading(box)
	return Box{X: head.X, Y: head.Y + head.H + chipAt, W: head.W, H: 1}
}

// Framed is the box drawn around what a field holds, tall enough for every
// line that field is given. Every field is as wide as every other, and each
// leaves at the end of its line the room a long one's bar takes: a title and a
// description end on the same line.
func Framed(value Box, rows int) Box {
	return Box{X: value.X, Y: value.Y - inset, W: value.W - barWide - between, H: rows*lineHigh + inset + inset}
}

// FormFields is the left column of a form opened on something already there.
func FormFields(box Box) Box {
	return Box{X: box.X, Y: box.Y, W: box.W * fieldsPart / ofPanel, H: box.H}
}

// FormBeside is the right column, beside the fields.
func FormBeside(box Box) Box {
	left := FormFields(box)
	return Box{X: left.X + left.W, Y: box.Y, W: box.W - left.W, H: box.H}
}

// BesideHead is where the column beside a form says what it is: what was done
// to a task, or the work serving an objective. One column for both: it is what
// was opened that decides what the column beside it has to say.
func BesideHead(box Box) Box {
	right := FormBeside(box)
	return Box{X: right.X + padding, Y: FirstRow(box), W: right.W - padding - padding, H: lineHigh}
}

// BesideBox is the frame that column is drawn in: from under its own heading
// down to just above what the panel offers, so it never reaches either.
func BesideBox(box Box) Box {
	head := BesideHead(box)
	top := head.Y + head.H + between
	return Box{
		X: head.X,
		Y: top,
		W: head.W - barWide - between,
		H: max(lineHigh, PanelOfferAt(box, 0).Y-padding-top),
	}
}

// Saying is a long field's frame and the bar beside its right edge, given the
// value box its row has and how many lines it is given. One shape for the
// detail and the form: the description is the same box on both, with the same
// bar in the same place, and the wheel reaches the same end.
//
// The bar is drawn whether or not there is anything to scroll — the same rule
// the history goes by.
func Saying(value Box, rows int) (frame, bar Box) {
	frame = Framed(value, rows)
	bar = Box{X: frame.X + frame.W + between, Y: frame.Y, W: barWide, H: frame.H}
	return frame, bar
}

// BesideBar is the track of the bar beside that frame.
func BesideBar(box Box) Box {
	frame := BesideBox(box)
	return Box{X: frame.X + frame.W + between, Y: frame.Y, W: barWide, H: frame.H}
}

// BesideRows is how many lines the frame beside a form shows at once.
func BesideRows(box Box) int { return max(1, Inside(BesideBox(box)).H/lineHigh) }

// Told is the history as the lines it takes, each act wrapped the way it is
// drawn, the last done first and one straight after another. Each act begins
// with when it was done, which is what says where it starts: a blank line
// between two of them read as an empty line between every line of a history
// whose acts each fit on one, and the oldest first put what was just done at
// the bottom of a box opened at its top.
func Told(acts []string, wrap func(string) Lines) []string {
	var lines []string
	for i := range acts {
		lines = append(lines, wrap(acts[len(acts)-1-i]).Said...)
	}
	return lines
}

// between is what one framed field leaves before the next. Small, because a
// frame already says where a field ends — the room a line needed when nothing
// was drawn around it is room a form no longer has to spend.
const between = 6

// apart is how far one line of a panel is from the next: the same step a form
// puts between two fields, which is lineHigh and two insets and the room one
// field leaves before the next. It was 26 where a field was 32, which is the
// same text written at two rhythms in one window — the help panel and the form
// beside it read as two pieces of software.
//
// Written as the number and not as that sum. A constant's arithmetic runs on
// no line, so no test can reach it and nothing would notice it changing; the
// test beside fieldHigh holds the two together instead.
const apart = 32

// The scale everything is spaced on, in steps of four: 4 inside a frame, 12
// between two things, 32 from one line of anything to the next. A number that
// is not on it is a number somebody will have to measure to match.

// longRows is how many rows a field that wants more than one gets. Four: a
// description is a paragraph, and a paragraph in one line is a slot the size of
// a word.
const longRows = 4
