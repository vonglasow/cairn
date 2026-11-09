package gui

const (
	// ItemHigh is what one item of a list beside a detail takes: its two lines and the gap under it.
	ItemHigh = 50
	itemGap  = 4
	indent   = 24
	itemPad  = 14
	itemTop  = 6
	factName = 74
	// FiguresHigh is what a row of figures takes.
	FiguresHigh = titleHi + 4
	// TileHigh is what a card in tile layout takes: its line, its track, its counts and two lines of notes.
	TileHigh = 112
	tileWide = 160
	splitTo  = 420
)

// ItemOf is where an item stands in its slot: moved right by its depth, the gap under it left out.
func ItemOf(slot Box, depth int) Box {
	in := depth * indent
	return Box{X: slot.X + in, Y: slot.Y, W: slot.W - in, H: slot.H - itemGap}
}

// ItemLine is an item's first line: its mark, its number and its title, a chip at its right; and the second, its progress, under the title.
func ItemLine(item Box) (mark, first, second Box) {
	mark = Box{X: item.X + itemPad, Y: item.Y + itemTop + (lineHigh-markSide)/2, W: markSide, H: markSide}
	from := mark.X + markSide + markSpace
	first = Box{X: from, Y: item.Y + itemTop, W: item.X + item.W - itemPad - from, H: lineHigh}
	second = Box{X: from, Y: first.Y + lineHigh + 2, W: first.W, H: lineHigh}
	return mark, first, second
}

// JoinOf is the room left of an item where its connector runs: from under its parent's mark to the item.
func JoinOf(slot Box, depth int) Box {
	item := ItemOf(slot, depth)
	x := slot.X + itemPad + markSide/2
	return Box{X: x, Y: slot.Y, W: item.X - x, H: slot.H}
}

// TileParts is where a card in tile layout writes: its first line, its track, its counts, and its notes under them.
func TileParts(panel Box) (first, track, counts, notes Box) {
	left, inside := panel.X+cardLeft, panel.W-cardLeft-cardRight
	first = Box{X: left, Y: panel.Y + padding - 2, W: inside, H: lineHigh}
	track = Box{X: left, Y: first.Y + lineHigh + 6, W: inside, H: 6}
	counts = Box{X: left, Y: track.Y + track.H + 4, W: inside, H: lineHigh}
	notes = Box{X: left, Y: counts.Y + lineHigh + 4, W: inside, H: 2 * lineHigh}
	return first, track, counts, notes
}

// SplitOf shares the board between the list of objectives and the detail of the one chosen: the list no wider than it needs, the detail the rest.
func SplitOf(l Layout) (list, detail Box) {
	last := l.Column[len(l.Column)-1]
	run := Box{X: l.Column[0].X, Y: l.Head[0].Y, W: last.X + last.W - l.Column[0].X, H: l.Column[0].Y + l.Column[0].H - l.Head[0].Y}
	wide := min(splitTo, run.W*2/5)
	list = Box{X: run.X, Y: run.Y, W: wide, H: run.H}
	detail = Box{X: run.X + wide + gap, Y: run.Y, W: run.W - wide - gap, H: run.H}
	return list, detail
}
