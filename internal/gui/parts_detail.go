package gui

// PanelOfferAt is where one of a panel's own offers sits: along its bottom, each in its own place so a pointer can land on one of them.
func PanelOfferAt(box Box, at int) Box {
	high := barHigh
	return Box{
		X: box.X + padding + at*offerWide,
		Y: box.Y + box.H - padding - high,
		W: offerWide,
		H: high,
	}
}

// AskingOver is a box asking a question of a few lines: centred across the window, a third of the way down, and as tall as its heading, its lines and its offers.
func AskingOver(width, height, lines int) Box {
	top := FirstRow(Box{})
	return Box{X: (width - askingWide) / 2, Y: height / 4, W: askingWide, H: top + lines*apart + barHigh + padding}
}

const askingWide = 560

// FirstRow is where the first line inside a panel begins: under its heading.
func FirstRow(box Box) int {
	ruled := Ruled(box)
	return ruled.Y + ruled.H + padding
}

// ContextName is where an open task names what the line beside it says: at its foot, past the last place an offer can take.
func ContextName(box Box) Box {
	after := PanelOfferAt(box, detailOffers)
	return Box{X: after.X + padding, Y: after.Y + (after.H-lineHigh)/2, W: contextNamed, H: lineHigh}
}

// ContextPath is where it says where what was written about it elsewhere lives: from the end of its name to the panel's edge.
func ContextPath(box Box) Box {
	name := ContextName(box)
	from := name.X + name.W
	return Box{X: from, Y: name.Y, W: box.X + box.W - padding - from, H: lineHigh}
}

// Chip is the patch drawn behind a key so it reads as a key and not as a word.
func Chip(name Box) Box {
	return Box{X: name.X - chipAt, Y: name.Y - chipAt, W: name.W, H: name.H + chipAt + chipAt}
}

// Heading is where a panel says what it is, above everything it holds.
func Heading(box Box) Box {
	return Box{X: box.X + padding, Y: box.Y + padding, W: box.W - padding - padding, H: titleHi}
}

// Ruled is the line under that heading, which is what makes it a heading rather than the first row.
func Ruled(box Box) Box {
	head := Heading(box)
	return Box{X: head.X, Y: head.Y + head.H + chipAt, W: head.W, H: 1}
}

// Framed is the box drawn around what a field holds, tall enough for every line that field is given.
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

// BesideHead is where the column beside a form says what it is: what was done to a task, or the work serving an objective.
func BesideHead(box Box) Box {
	right := FormBeside(box)
	return Box{X: right.X + padding, Y: FirstRow(box), W: right.W - padding - padding, H: lineHigh}
}

// BesideBox is the frame that column is drawn in: from under its own heading down to just above what the panel offers, so it never reaches either.
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

// Saying is a long field's frame and the bar beside its right edge, given the value box its row has and how many lines it is given.
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

// Told is the history as the lines it takes, each act wrapped the way it is drawn, the last done first and one straight after another.
func Told(acts []string, wrap func(string) Lines) []string {
	var lines []string
	for i := range acts {
		lines = append(lines, wrap(acts[len(acts)-1-i]).Said...)
	}
	return lines
}

const between = 6

const apart = 32

const longRows = 4

const closeWide = 24

// FormClose is the box at the right end of a form's heading that closes it, no smaller than a pointer can be sure of.
func FormClose(box Box) Box {
	head := Heading(box)
	return Box{X: head.X + head.W - closeWide, Y: head.Y + (head.H-closeWide)/2, W: closeWide, H: closeWide}
}

// FormHeading is a form's heading less the box that closes it, so its words never run under the cross.
func FormHeading(box Box) Box {
	head := Heading(box)
	return Box{X: head.X, Y: head.Y, W: head.W - closeWide - between, H: head.H}
}
