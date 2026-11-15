package gui

import (
	"cairn/internal/gui/geom"
	"strings"
)

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
func ItemOf(slot geom.Box, depth int) geom.Box {
	in := depth * indent
	return geom.Box{X: slot.X + in, Y: slot.Y, W: slot.W - in, H: slot.H - itemGap}
}

// ItemLine is an item's first line: its mark, its number and its title, a chip at its right; and the second, its progress, under the title.
func ItemLine(item geom.Box) (mark, first, second geom.Box) {
	mark = geom.Box{X: item.X + itemPad, Y: item.Y + itemTop + (lineHigh-markSide)/2, W: markSide, H: markSide}
	from := mark.X + markSide + markSpace
	first = geom.Box{X: from, Y: item.Y + itemTop, W: item.X + item.W - itemPad - from, H: lineHigh}
	second = geom.Box{X: from, Y: first.Y + lineHigh + 2, W: first.W, H: lineHigh}
	return mark, first, second
}

// JoinOf is the room left of an item where its connector runs: from under its parent's mark to the item.
func JoinOf(slot geom.Box, depth int) geom.Box {
	item := ItemOf(slot, depth)
	x := slot.X + itemPad + markSide/2
	return geom.Box{X: x, Y: slot.Y, W: item.X - x, H: slot.H}
}

// TileParts is where a card in tile layout writes: its first line, its track, its counts, and its notes under them.
func TileParts(panel geom.Box) (first, track, counts, notes geom.Box) {
	left, inside := panel.X+cardLeft, panel.W-cardLeft-cardRight
	first = geom.Box{X: left, Y: panel.Y + padding - 2, W: inside, H: lineHigh}
	track = geom.Box{X: left, Y: first.Y + lineHigh + 6, W: inside, H: 6}
	counts = geom.Box{X: left, Y: track.Y + track.H + 4, W: inside, H: lineHigh}
	notes = geom.Box{X: left, Y: counts.Y + lineHigh + 4, W: inside, H: 2 * lineHigh}
	return first, track, counts, notes
}

// SplitOf shares the board between the list of objectives and the detail of the one chosen: the list no wider than it needs, the detail the rest.
func SplitOf(l geom.Layout) (list, detail geom.Box) {
	last := l.Column[len(l.Column)-1]
	run := geom.Box{X: l.Column[0].X, Y: l.Head[0].Y, W: last.X + last.W - l.Column[0].X, H: l.Column[0].Y + l.Column[0].H - l.Head[0].Y}
	wide := min(splitTo, run.W*2/5)
	list = geom.Box{X: run.X, Y: run.Y, W: wide, H: run.H}
	detail = geom.Box{X: run.X + wide + geom.Gap, Y: run.Y, W: run.W - wide - geom.Gap, H: run.H}
	return list, detail
}

// ChevronIn is the chevron drawn in a parent's gutter, a little smaller than the gutter and in its middle.
func ChevronIn(gutter geom.Box) geom.Box {
	side := gutter.W / 2
	return geom.Box{X: gutter.X + (gutter.W-side)/2, Y: gutter.Y + (gutter.H-side)/2, W: side, H: side}
}

// SignalAt is where an item's second line writes what wants seeing, against its right end, and the dot before it when it is hidden.
func SignalAt(second geom.Box, said string, dot bool, m Measure) (dotAt, saidAt geom.Box) {
	wide := min(second.W, ends(m, said, small)+2)
	saidAt = geom.Box{X: second.X + second.W - wide, Y: second.Y, W: wide, H: second.H}
	dotAt = saidAt
	if dot {
		dotAt = geom.Box{X: saidAt.X - dotSide - pillGap, Y: second.Y + (second.H-dotSide)/2, W: dotSide, H: dotSide}
	}
	return dotAt, saidAt
}

// FoundIn is where what a search reached is marked in a title written in at, and whether it is written there to be marked.
func FoundIn(at geom.Box, written, found string, m Measure) (geom.Box, bool) {
	if found == "" {
		return geom.Box{}, false
	}
	from := strings.Index(strings.ToLower(written), strings.ToLower(found))
	if from < 0 {
		return geom.Box{}, false
	}
	x := at.X + ends(m, written[:from], reads)
	wide := min(ends(m, written[from:from+len(found)], reads), at.X+at.W-x)
	return geom.Box{X: x - 1, Y: at.Y + 2, W: wide + 2, H: at.H - 4}, wide > 0
}
