package gui

import (
	"cairn/internal/reading"
)

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

// DetailFields is the left column of the detail panel.
func DetailFields(box Box) Box {
	return Box{X: box.X, Y: box.Y, W: box.W * fieldsPart / ofPanel, H: box.H}
}

// DetailHistory is the right column, which the history has to itself.
func DetailHistory(box Box) Box {
	left := DetailFields(box)
	return Box{X: left.X + left.W, Y: box.Y, W: box.W - left.W, H: box.H}
}

// ServingHead is where the box under the fields says what it is: the left
// column, below the last field, in the room the fields did not take.
func ServingHead(box Box, fields []reading.Field) Box {
	left := DetailFields(box)
	return Box{
		X: left.X + padding,
		Y: FieldTop(box, fields, len(fields)),
		W: left.W - padding - padding,
		H: lineHigh,
	}
}

// ServingBox is the frame the work under an objective is drawn in: from under
// its own heading down to just above what the panel offers, so it reaches
// neither. Never smaller than one line, the way the history's frame is not —
// a frame that collapses is a frame somebody reads as missing.
func ServingBox(box Box, fields []reading.Field) Box {
	head := ServingHead(box, fields)
	top := head.Y + head.H + between
	return Box{
		X: head.X,
		Y: top,
		W: head.W - barWide - between,
		H: max(lineHigh, PanelOfferAt(box, 0).Y-padding-top),
	}
}

// ServingBar is the track beside that box, in the room its frame gave up for
// it — the same shape the history's bar has.
func ServingBar(box Box, fields []reading.Field) Box {
	frame := ServingBox(box, fields)
	return Box{X: frame.X + frame.W + between, Y: frame.Y, W: barWide, H: frame.H}
}

// HistoryHead is where that column says what it is.
func HistoryHead(box Box) Box {
	right := DetailHistory(box)
	return Box{X: right.X + padding, Y: FirstRow(box), W: right.W - padding - padding, H: lineHigh}
}

// HistoryBox is the frame the history is drawn in: from under its own heading
// down to just above what the panel offers, so it never reaches either.
func HistoryBox(box Box) Box {
	head := HistoryHead(box)
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

// SayingBox is where a task's description is read on the detail.
func SayingBox(box Box, fields []reading.Field, at int) Box {
	_, value := FieldRow(DetailFields(box), FieldTop(DetailFields(box), fields, at))
	frame, _ := Saying(value, LongRows(box, fields))
	return frame
}

// SayingBar is the track of the bar beside it.
func SayingBar(box Box, fields []reading.Field, at int) Box {
	_, value := FieldRow(DetailFields(box), FieldTop(DetailFields(box), fields, at))
	_, bar := Saying(value, LongRows(box, fields))
	return bar
}

// LongField says which of a panel's fields is the one read in that box, and
// whether there is one at all.
func LongField(fields []reading.Field) (int, bool) {
	for i, field := range fields {
		if field.Long {
			return i, true
		}
	}
	return 0, false
}

// HistoryBar is the track of the bar beside it.
func HistoryBar(box Box) Box {
	frame := HistoryBox(box)
	return Box{X: frame.X + frame.W + between, Y: frame.Y, W: barWide, H: frame.H}
}

// HistoryRows is how many lines of history the frame can show at once.
func HistoryRows(box Box) int { return max(1, Inside(HistoryBox(box)).H/lineHigh) }

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

// fieldHigh is what one field takes from the top of its frame to the top of
// the next: its lines, the room its frame keeps, and what it leaves after it.
func fieldHigh(box Box, fields []reading.Field, at int) int {
	return RowsOf(box, fields, at)*lineHigh + inset + inset + between
}

// The scale everything is spaced on, in steps of four: 4 inside a frame, 12
// between two things, 32 from one line of anything to the next. A number that
// is not on it is a number somebody will have to measure to match.

// longRows is how many rows a field that wants more than one gets. Four: a
// description is a paragraph, and a paragraph in one line is a slot the size of
// a word.
const longRows = 4

// RowsOf is how many rows a field takes.
func RowsOf(box Box, fields []reading.Field, at int) int {
	if fields[at].Long {
		return LongRows(box, fields)
	}
	return 1
}

// LongRows is how many lines the one field that wants more than one gets: four
// at least, and every line the panel has room for beyond that.
//
// It used to be four whatever the window was. A description of 2800 characters
// comes to seventy-eight lines once wrapped, so it was read a hundred and fifty
// characters at a time through a box the panel had two hundred unused points
// underneath — which reads as a text cut off rather than one to scroll.
//
// What it leaves behind is what the box under the fields needs to be worth
// drawing: its heading and four tiles. Below that the window is too short to
// give anything away, and the four lines stand.
//
// On a window somebody works in it leaves the four as they were: twelve fields
// take two thirds of the column and the parts box the rest. It is on a tall
// screen that it gives anything, and reading a description of eighty lines is
// what opening it on the whole panel is for.
func LongRows(box Box, fields []reading.Field) int {
	taken := 0
	for i := range fields {
		if !fields[i].Long {
			taken += lineHigh + inset + inset + between
		}
	}
	under := lineHigh + between + 4*SubHeight + padding
	room := PanelOfferAt(box, 0).Y - padding - FirstRow(box) - taken - under - inset - inset - between
	return max(longRows, room/lineHigh)
}

// FieldTop is where a field begins, counting what the fields above it took. A
// form is not a grid of equal rows once one of them wants four.
func FieldTop(box Box, fields []reading.Field, at int) int {
	top := FirstRow(box)
	// One guard and not two: a second condition on the same loop has a
	// boundary nothing can reach, since what is counted is only ever the
	// fields that exist.
	for i := range fields[:min(at, len(fields))] {
		top += fieldHigh(box, fields, i)
	}
	return top
}
