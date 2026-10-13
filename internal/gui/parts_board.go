package gui

import (
	"strings"

	"cairn/internal/domain"
	"cairn/internal/engine"
)

// Where the parts of the board itself go: the tabs, the lanes, the
// objectives, the search box and the card dragged over it.

// TabAt is where one of the two tabs sits. It stands a little below the top of
// the band, so the band is something the tabs sit in rather than something
// they fill.
func TabAt(l Layout, at int) Box {
	return Box{X: l.Tabs.X + at*tabWide, Y: l.Tabs.Y + between, W: tabWide, H: l.Tabs.H - between}
}

// The tabs are folders, not two words one of which is coloured.
//
// A band closed by a line runs the width of the window, and the tab somebody
// is on is a block whose bottom breaks that line — it does not stand out by
// its colour, it stands out by opening onto the page. That is what says "you
// are on this page" rather than "this word is selected", and a colour contrast
// is what it replaced: the two words sat in the same emptiness and neither of
// them looked like a page.

// TabsLine is the line closing the band, run the width of the window.
func TabsLine(l Layout) Box {
	return Box{X: l.Tabs.X, Y: l.Tabs.Y + l.Tabs.H - 1, W: l.Tabs.W, H: 1}
}

// TabsLineAround is that line in the two pieces it is drawn in: the tab
// somebody is on breaks it, which is the whole of the folder.
func TabsLineAround(l Layout, at int) [2]Box {
	line, tab := TabsLine(l), TabAt(l, at)
	return [2]Box{
		{X: line.X, Y: line.Y, W: max(0, tab.X-line.X), H: line.H},
		{X: tab.X + tab.W, Y: line.Y, W: max(0, line.X+line.W-tab.X-tab.W), H: line.H},
	}
}

// TabSides is the three sides a tab is drawn with: a folder is open at the
// bottom, which is where the page it opens onto begins.
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

// OfferAt is where one of the bar's offers sits. They are places a pointer can
// land, which is what makes anything at all reachable without a shortcut.
func OfferAt(l Layout, at int) Box {
	return Box{X: l.Bar.X + at*offerWide, Y: l.Bar.Y, W: offerWide, H: l.Bar.H}
}

// GoalsRun is the whole of the board the objectives are laid out in: every
// column's width, and the height the first of them runs to.
//
// Written once and used three times — to cut the grid to, to say how far it
// scrolls, and to bring what is chosen into sight. Cut to one column's width
// instead, three quarters of the grid went missing.
func GoalsRun(l Layout) Box {
	last := l.Column[engine.Columns-1]
	return Box{
		X: l.Column[0].X,
		Y: l.Column[0].Y,
		W: last.X + last.W - l.Column[0].X,
		H: l.Column[0].H,
	}
}

// GoalsAcross is how many objectives stand side by side. The board's own
// columns, so an objective is the width of a card and the grid reflows with
// the window without a sum of its own.
func GoalsAcross() int { return engine.Columns }

// GoalAt is where the n-th objective sits in the grid, filled across and then
// down.
//
// A grid and not a list. One to a row, each the width of the whole board, an
// objective was a card's height stretched over four columns of nothing — and
// a dozen of them ran off the bottom of the window, so the ones that mattered
// least were the only ones in sight. Across, a dozen fit without scrolling at
// the smallest window this board draws in.
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

// Searching is where what is being looked for is written: a line along the
// bottom, taking the rest of the bar beside the one offer that closes it.
//
// Along the bottom and not across the middle, which is where a window says it
// is too small — a place nobody types, and a place that covers the very board
// a search is narrowing.
// It leaves the right of the bar to the offers: taking the whole of it left a
// search with no way out but a key, and no way at all to act on the card it
// had just found.
func Searching(l Layout) Box {
	b := l.Bar
	// It begins after the one offer the bar keeps, where every offer is drawn.
	// Taking the whole bar left a key as the only way out of a search, and no
	// way at all to act on the card the search had just found.
	kept := offerWide + padding
	b.X, b.W = b.X+kept, b.W-kept
	return b
}

// SearchWord is where the box says what it is, and SearchInto where what is
// being looked for is typed.
//
// Said in a word and not in a slash: the key that opens a search used to be
// drawn where the text goes, so the first thing anybody saw was a character
// they had not typed, in a place that looked like nothing in particular.
func SearchWord(b Box) Box {
	return Box{X: b.X + padding, Y: b.Y + (b.H-lineHigh)/2, W: searchWide, H: lineHigh}
}

// SearchInto is the rest of that box, which is what is typed into.
func SearchInto(b Box) Box {
	return Box{X: b.X + padding + searchWide, Y: b.Y + (b.H-lineHigh)/2, W: b.W - padding - padding - searchWide, H: lineHigh}
}

// WrongBox is where a refusal is written: two lines just above the bar,
// growing upwards so the foot of the window stays where it is.
//
// Two, because the most instructive refusal cairn has runs past one — "it is
// still blocked, by #3 and by #4. take them away first: task 3, task 4" was cut
// just before the part that says what to do. Above the bar and not over it: it
// used to take the bar's place, and every offer went with it until the next
// good read.
func WrongBox(l Layout) Box {
	return Box{X: l.Bar.X, Y: l.Bar.Y - 2*lineHigh, W: l.Bar.W, H: 2 * lineHigh}
}

// Stamped splits a history line into the time it happened and what happened,
// and says whether it is a line that begins with one at all. A line a long act
// wrapped onto carries no time, and writing part of it in one ink and the rest
// in another would make one sentence read as two.
//
// The history was thirty lines of the faintest ink the window has — a thing to
// read, written in the ink kept for the things beside it.
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

// LabelsRoom is how much of a card's bottom row the labels get: what is left
// once the facts written hard right have taken theirs, and the padding that
// keeps the two from touching. Nothing at all when there is nothing left,
// since a width below zero is not a narrower line but a wider one.
func LabelsRoom(labels Box, said string) int {
	if said == "" {
		return labels.W
	}
	return max(0, labels.W-CountedAt(labels, said).W-padding)
}

// StampedOver is how far across a history line what was done is written: past
// the time it happened and the two spaces after it. Here rather than in the
// drawing, for the reason this file opens with.
func StampedOver(stamp string) int {
	return (len([]rune(stamp)) + 2) * perSmallGlyph
}

// PastStamp is where what was done is written on a history line: past the time
// it happened, and no wider than what is left of the line.
func PastStamp(box Box, stamp string) Box {
	over := StampedOver(stamp)
	return Box{X: box.X + over, Y: box.Y, W: box.W - over, H: box.H}
}
