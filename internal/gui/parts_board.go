package gui

import (
	"strings"

	"cairn/internal/domain"
)

// TabAt is where one of the two tabs sits.
func TabAt(l Layout, at int) Box {
	return Box{X: l.Tabs.X + at*tabWide, Y: l.Tabs.Y + between, W: tabWide, H: l.Tabs.H - between}
}

// TabName is where a tab's name and what it holds are written inside it.
func TabName(b Box) Box {
	return Box{X: b.X + padding, Y: b.Y + (b.H-lineHigh)/2, W: b.W - padding - padding, H: lineHigh}
}

// LaneName is where a lane's name is written, above the cards in it.
func LaneName(l Layout, lane int) Box {
	return NameIn(Box{X: l.Lane[lane].X, Y: l.Band.Y, W: l.Lane[lane].W, H: laneHigh})
}

const (
	nameDrop  = 7
	nameInset = 10
)

// NameIn is the room a heading is written in: its name at the left, its count at the right, a few pixels inside its edges, for a column and a lane alike.
func NameIn(head Box) Box {
	return Box{X: head.X + nameInset, Y: head.Y + nameDrop, W: head.W - 2*nameInset, H: head.H - nameDrop}
}

// LaneGround is the ground a lane of the band stands on: its head and its cards.
func LaneGround(l Layout, lane int) Box {
	return Box{X: l.Lane[lane].X, Y: l.Band.Y, W: l.Lane[lane].W, H: l.Lane[lane].Y + l.Lane[lane].H - l.Band.Y}
}

// EmptyAt is where a column or a lane holding nothing says so: where its first card would be.
func EmptyAt(slot Box) Box { return Box{X: slot.X, Y: slot.Y, W: slot.W, H: emptyHigh} }

const emptyHigh = 44

// OfferAt is where one of the bar's offers sits.
func OfferAt(l Layout, at int) Box {
	return Box{X: l.Bar.X + at*offerWide, Y: l.Bar.Y, W: offerWide, H: l.Bar.H}
}

// GoalsRun is where the objectives are listed: the list's ground under its head, which scrolls.
func GoalsRun(l Layout) Box {
	list, _ := SplitOf(l)
	return Box{X: list.X + gap, Y: list.Y + headHigh, W: list.W - 2*gap, H: list.H - headHigh - gap}
}

// GoalsHead is the head naming the list of objectives, with how many it holds.
func GoalsHead(l Layout) Box {
	list, _ := SplitOf(l)
	return NameIn(Box{X: list.X, Y: list.Y, W: list.W, H: headHigh})
}

// GoalAt is where the n-th objective of the tree has its slot, one under the other, right of the gutter its chevron stands in.
func GoalAt(l Layout, at, scrolled int) Box {
	run := GoalsRun(l)
	return Box{X: run.X + foldGutter, Y: run.Y + at*ItemHigh - scrolled, W: run.W - foldGutter, H: ItemHigh}
}

// FoldAt is where a parent's chevron stands: in the gutter left of its slot, level with its first line.
func FoldAt(slot Box) Box {
	_, first, _ := ItemLine(ItemOf(slot, 0))
	return Box{X: slot.X - foldGutter, Y: first.Y + (first.H-foldGutter)/2, W: foldGutter, H: foldGutter}
}

const foldGutter = 18

// Searching is where what is being looked for is written: the command bar itself, the search being what it turns into after a slash.
func Searching(l Layout) Box { return l.Bar }

// SearchWord is the chip at the left of that bar saying it searches, a slash, and SearchInto where what is being looked for is typed, short of the hints at its right.
func SearchWord(b Box) Box {
	return Box{X: b.X + barInset, Y: b.Y + (b.H-chipHigh)/2, W: KeyChipWide("/"), H: chipHigh}
}

// SearchInto is the rest of the bar's line after that chip, which is what is typed into.
func SearchInto(b Box) Box {
	word := SearchWord(b)
	from := word.X + word.W + hintSpace
	return Box{X: from, Y: b.Y + (b.H-lineHigh)/2, W: b.X + b.W - searchHints - from, H: lineHigh}
}

const searchHints = 240

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

// FactsAt is where a card's row writes its facts: at its right end, as wide as they are.
func FactsAt(row Box, wide int) Box { return Box{X: row.X + row.W - wide, Y: row.Y, W: wide, H: row.H} }

// StampedOver is how far across a history line what was done is written: past the time it happened and the two spaces after it.
func StampedOver(stamp string) int {
	return (len([]rune(stamp)) + 2) * perSmallGlyph
}

// PastStamp is where what was done is written on a history line: past the time it happened, and no wider than what is left of the line.
func PastStamp(box Box, stamp string) Box {
	over := StampedOver(stamp)
	return Box{X: box.X + over, Y: box.Y, W: box.W - over, H: box.H}
}
