package gui

import (
	"strings"

	"cairn/internal/domain"
	"cairn/internal/engine"
)

// TabAt is where one of the two tabs sits.
func TabAt(l Layout, at int) Box {
	return Box{X: l.Tabs.X + at*tabWide, Y: l.Tabs.Y + between, W: tabWide, H: l.Tabs.H - between}
}

// TabsLine is the line closing the band, run the width of the window.
func TabsLine(l Layout) Box {
	return Box{X: l.Tabs.X, Y: l.Tabs.Y + l.Tabs.H - 1, W: l.Tabs.W, H: 1}
}

// TabsLineAround is that line in the two pieces it is drawn in: the tab somebody is on breaks it, which is the whole of the folder.
func TabsLineAround(l Layout, at int) [2]Box {
	line, tab := TabsLine(l), TabAt(l, at)
	return [2]Box{
		{X: line.X, Y: line.Y, W: max(0, tab.X-line.X), H: line.H},
		{X: tab.X + tab.W, Y: line.Y, W: max(0, line.X+line.W-tab.X-tab.W), H: line.H},
	}
}

// TabSides is the three sides a tab is drawn with: a folder is open at the bottom, which is where the page it opens onto begins.
func TabSides(b Box) [3]Box {
	return [3]Box{
		{X: b.X, Y: b.Y, W: b.W, H: 1},
		{X: b.X, Y: b.Y, W: 1, H: b.H},
		{X: b.X + b.W - 1, Y: b.Y, W: 1, H: b.H},
	}
}

// TabName is where a tab's name and what it holds are written inside it.
func TabName(b Box) Box {
	return Box{X: b.X + padding, Y: b.Y + (b.H-lineHigh)/2, W: b.W - padding - padding, H: lineHigh}
}

// LaneName is where a lane's name is written, above the cards in it.
func LaneName(l Layout, lane int) Box {
	return Box{X: l.Lane[lane].X, Y: l.Band.Y, W: l.Lane[lane].W, H: laneHigh}
}

const nameDrop = 7

// ColumnName is the room a column's name is centred in: the whole width of its head, from a few pixels under its top edge.
func ColumnName(head Box) Box {
	return Box{X: head.X, Y: head.Y + nameDrop, W: head.W, H: head.H - nameDrop}
}

// OfferAt is where one of the bar's offers sits.
func OfferAt(l Layout, at int) Box {
	return Box{X: l.Bar.X + at*offerWide, Y: l.Bar.Y, W: offerWide, H: l.Bar.H}
}

// GoalsRun is the whole of the board the objectives are laid out in: every column's width, and the height the first of them runs to.
func GoalsRun(l Layout) Box {
	last := l.Column[engine.Columns-1]
	return Box{
		X: l.Column[0].X,
		Y: l.Column[0].Y,
		W: last.X + last.W - l.Column[0].X,
		H: l.Column[0].H,
	}
}

// GoalsAcross is how many objectives stand side by side.
func GoalsAcross() int { return engine.Columns }

// GoalAt is where the n-th objective sits in the grid, filled across and then down.
func GoalAt(l Layout, at, scrolled int) Box {
	across := GoalsAcross()
	column := l.Column[at%across]
	return Box{
		X: column.X,
		Y: l.Column[0].Y + (at/across)*CardHeight - scrolled,
		W: column.W,
		H: CardHeight,
	}
}

// GoalRows is how many rows a given number of objectives comes to.
func GoalRows(goals int) int {
	across := GoalsAcross()
	return (goals + across - 1) / across
}

// Searching is where what is being looked for is written: a line along the bottom, taking the rest of the bar beside the one offer that closes it.
func Searching(l Layout) Box {
	b := l.Bar
	kept := offerWide + padding
	b.X, b.W = b.X+kept, b.W-kept
	return b
}

// SearchWord is where the box says what it is, and SearchInto where what is being looked for is typed.
func SearchWord(b Box) Box {
	return Box{X: b.X + padding, Y: b.Y + (b.H-lineHigh)/2, W: searchWide, H: lineHigh}
}

// SearchInto is the rest of that box, which is what is typed into.
func SearchInto(b Box) Box {
	return Box{X: b.X + padding + searchWide, Y: b.Y + (b.H-lineHigh)/2, W: b.W - padding - padding - searchWide, H: lineHigh}
}

// WrongBox is where a refusal is written: two lines just above the bar, growing upwards so the foot of the window stays where it is.
func WrongBox(l Layout) Box {
	return Box{X: l.Bar.X, Y: l.Bar.Y - 2*lineHigh, W: l.Bar.W, H: 2 * lineHigh}
}

// Stamped splits a history line into the time it happened and what happened, and says whether it is a line that begins with one at all.
func Stamped(said string) (stamp, rest string, ok bool) {
	stamp, rest, found := strings.Cut(said, "  ")
	if !found {
		return "", said, false
	}
	if _, err := domain.ParseMoment(stamp); err != nil {
		return "", said, false
	}
	return stamp, rest, true
}

// LabelsRoom is how much of a card's bottom row the labels get: what is left once the facts written hard right have taken theirs, and the padding that keeps the two from touching.
func LabelsRoom(labels Box, said string) int {
	if said == "" {
		return labels.W
	}
	return max(0, labels.W-CountedAt(labels, said).W-padding)
}

// LabelsAt is where a card's labels are written: its bottom row, short of the facts written hard right.
func LabelsAt(labels Box, said string) Box {
	return Box{X: labels.X, Y: labels.Y, W: LabelsRoom(labels, said), H: labels.H}
}

// Crowded says a card's count would be written over what its number line already holds up to taken: its number, and its eye.
func Crowded(number Box, taken int, said string) bool {
	return CountedAt(number, said).X < taken+eyeSpace
}

// StampedOver is how far across a history line what was done is written: past the time it happened and the two spaces after it.
func StampedOver(stamp string) int {
	return (len([]rune(stamp)) + 2) * perSmallGlyph
}

// PastStamp is where what was done is written on a history line: past the time it happened, and no wider than what is left of the line.
func PastStamp(box Box, stamp string) Box {
	over := StampedOver(stamp)
	return Box{X: box.X + over, Y: box.Y, W: box.W - over, H: box.H}
}
