package gui

import (
	"cairn/internal/board"
	"cairn/internal/engine/calendar"
	"cairn/internal/gui/geom"
	"slices"
	"strings"
)

// TabAt is where one of the two tabs sits.
func TabAt(l geom.Layout, at int) geom.Box {
	return geom.Box{X: l.Tabs.X + at*tabWide, Y: l.Tabs.Y + between, W: tabWide, H: l.Tabs.H - between}
}

// TabName is where a tab's name and what it holds are written inside it.
func TabName(b geom.Box) geom.Box {
	return geom.Box{X: b.X + padding, Y: b.Y + (b.H-lineHigh)/2, W: b.W - padding - padding, H: lineHigh}
}

// LaneName is where a lane's name is written, above the cards in it.
func LaneName(l geom.Layout, lane int) geom.Box {
	return NameIn(geom.Box{X: l.Lane[lane].X, Y: l.Band.Y, W: l.Lane[lane].W, H: geom.LaneHigh})
}

const (
	nameDrop  = 7
	nameInset = 10
)

// NameIn is the room a heading is written in: its name at the left, its count at the right, a few pixels inside its edges, for a column and a lane alike.
func NameIn(head geom.Box) geom.Box {
	return geom.Box{X: head.X + nameInset, Y: head.Y + nameDrop, W: head.W - 2*nameInset, H: head.H - nameDrop}
}

// LaneGround is the ground a lane of the band stands on: its head and its cards.
func LaneGround(l geom.Layout, lane int) geom.Box {
	return geom.Box{X: l.Lane[lane].X, Y: l.Band.Y, W: l.Lane[lane].W, H: l.Lane[lane].Y + l.Lane[lane].H - l.Band.Y}
}

// EmptyAt is where a column or a lane holding nothing says so: where its first card would be.
func EmptyAt(slot geom.Box) geom.Box { return geom.Box{X: slot.X, Y: slot.Y, W: slot.W, H: emptyHigh} }

const emptyHigh = 44

// OfferAt is where one of the bar's offers sits.
func OfferAt(l geom.Layout, at int) geom.Box {
	return geom.Box{X: l.Bar.X + at*offerWide, Y: l.Bar.Y, W: offerWide, H: l.Bar.H}
}

// GoalsRun is where the objectives are listed: the list's ground under its head, which scrolls.
func GoalsRun(l geom.Layout) geom.Box {
	list, _ := SplitOf(l)
	return geom.Box{X: list.X + geom.Gap, Y: list.Y + geom.HeadHigh, W: list.W - 2*geom.Gap, H: list.H - geom.HeadHigh - geom.Gap}
}

// GoalsHead is the head naming the list of objectives, with how many it holds.
func GoalsHead(l geom.Layout) geom.Box {
	list, _ := SplitOf(l)
	return NameIn(geom.Box{X: list.X, Y: list.Y, W: list.W, H: geom.HeadHigh})
}

// GoalAt is where the n-th objective of the tree has its slot, one under the other, right of the gutter its chevron stands in.
func GoalAt(l geom.Layout, at, scrolled int) geom.Box {
	run := GoalsRun(l)
	return geom.Box{X: run.X + foldGutter, Y: run.Y + at*ItemHigh - scrolled, W: run.W - foldGutter - l.Listed, H: ItemHigh}
}

// FoldAt is where a parent's chevron stands: in the gutter left of its slot, level with its first line.
func FoldAt(slot geom.Box) geom.Box {
	_, first, _ := ItemLine(ItemOf(slot, 0))
	return geom.Box{X: slot.X - foldGutter, Y: first.Y + (first.H-foldGutter)/2, W: foldGutter, H: foldGutter}
}

const foldGutter = 18

// Searching is where what is being looked for is written: the command bar itself, the search being what it turns into after a slash.
func Searching(l geom.Layout) geom.Box { return l.Bar }

// SearchWord is the chip at the left of that bar saying it searches, a slash, and SearchInto where what is being looked for is typed, short of the hints at its right.
func SearchWord(b geom.Box) geom.Box {
	return geom.Box{X: b.X + barInset, Y: b.Y + (b.H-chipHigh)/2, W: KeyChipWide("/"), H: chipHigh}
}

// SearchInto is the rest of the bar's line after that chip, which is what is typed into.
func SearchInto(b geom.Box) geom.Box {
	word := SearchWord(b)
	from := word.X + word.W + hintSpace
	return geom.Box{X: from, Y: b.Y + (b.H-lineHigh)/2, W: b.X + b.W - searchHints - from, H: lineHigh}
}

const searchHints = 240

// WrongBox is where a refusal is written: two lines just above the bar, growing upwards so the foot of the window stays where it is.
func WrongBox(l geom.Layout) geom.Box {
	return geom.Box{X: l.Bar.X, Y: l.Bar.Y - 2*lineHigh, W: l.Bar.W, H: 2 * lineHigh}
}

// Stamped splits a history line into the time it happened and what happened, and says whether it is a line that begins with one at all.
func Stamped(said string) (stamp, rest string, ok bool) {
	stamp, rest, found := strings.Cut(said, "  ")
	if !found {
		return "", said, false
	}
	if _, err := calendar.ParseMoment(stamp); err != nil {
		return "", said, false
	}
	return stamp, rest, true
}

// FactsAt is where a card's row writes its facts: at its right end, as wide as they are.
func FactsAt(row geom.Box, wide int) geom.Box {
	return geom.Box{X: row.X + row.W - wide, Y: row.Y, W: wide, H: row.H}
}

// StampedOver is how far across a history line what was done is written: past the time it happened and the two spaces after it.
func StampedOver(stamp string) int {
	return (len([]rune(stamp)) + 2) * perSmallGlyph
}

// PastStamp is where what was done is written on a history line: past the time it happened, and no wider than what is left of the line.
func PastStamp(box geom.Box, stamp string) geom.Box {
	over := StampedOver(stamp)
	return geom.Box{X: box.X + over, Y: box.Y, W: box.W - over, H: box.H}
}

// TabsAt is where each tab shown stands, as wide as its name, its count and its alert need, at most tabWide, then where the room after them starts, as one more box.
func (w Window) TabsAt() []geom.Box {
	l, m := w.Layout(), w.measured()
	at := TabAt(l, 0)
	var boxes []geom.Box
	for _, page := range w.Shown() {
		wide := 2*padding + ends(m, w.TabSaid(page), reads) + countGap + ends(m, w.TabCount(page), small) + tabSlack + tabArrowRoom(page)
		if w.TabAlert(page) > 0 {
			wide += alertSide + tabSlack
		}
		at.W = min(tabWide+tabArrowRoom(page)+ends(m, w.TabSaid(page), reads)-ends(m, Titled(Pages[page]), reads), wide)
		boxes = append(boxes, at)
		at.X += at.W
	}
	if tasks := slices.Index(w.Shown(), board.Tasks); tasks >= 0 {
		named := ends(m, w.TabSaid(board.Tasks), reads) - ends(m, Titled(Pages[board.Tasks]), reads)
		over := min(named, at.X-(MenuButton(l).X-between))
		if over > 0 {
			boxes[tasks].W -= over
			for i := tasks + 1; i < len(boxes); i++ {
				boxes[i].X -= over
			}
			at.X -= over
		}
	}
	at.W = 0
	return append(boxes, at)
}

const (
	tabSlack  = 8
	alertSide = 16
)
