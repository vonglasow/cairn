// Package gui draws cairn in a window a mouse can reach.
package gui

import (
	"cairn/internal/engine"
)

// Narrowest and Shortest are the window this board needs.
const (
	Narrowest = 1024
	Shortest  = 640
)

// CardHeight is what every card takes, whatever it carries.
const CardHeight = 96

// SubHeight is what a step takes in its box among a task's fields, and a serving task in the box beside an objective's: one line and the padding round it.
const SubHeight = 28

const (
	margin    = 12
	gap       = 12
	tabsHigh  = 40
	headHigh  = 28
	laneHigh  = 22
	barHigh   = 32
	bandCards = 2
)

// Point is somewhere in the window.
type Point struct{ X, Y int }

// Box is a rectangle of the window, in dp.
type Box struct{ X, Y, W, H int }

// In says whether a point falls inside.
func (b Box) In(p Point) bool {
	return p.X >= b.X && p.X < b.X+b.W && p.Y >= b.Y && p.Y < b.Y+b.H
}

// Layout is every size the window draws from.
type Layout struct {
	Tabs   Box
	Head   [engine.Columns]Box
	Column [engine.Columns]Box
	Band   Box
	Lane   [3]Box
	Bar    Box
}

// TooSmall says whether the board can be drawn in this window at all.
func TooSmall(width, height int) bool {
	return width < Narrowest || height < Shortest
}

// Sizes is how wide a person dragged each column and each lane, and how tall the columns above the band and the band, in dp at the size the window then had: what the board keeps of them is their shares, so a window resized keeps them in proportion. Zero, every column and every lane is as wide as the others, and the band as tall as it opens.
type Sizes struct {
	Column [engine.Columns]int `json:"columns"`
	Lane   [3]int              `json:"lanes"`
	Band   [2]int              `json:"band"`
}

// NarrowestColumn is the narrowest a gutter dragged leaves a column or a lane: room for a card's title, even when the window is wide enough for more.
const NarrowestColumn = 120

// LayoutFor works out every size at a given window, every column as wide as the others.
func LayoutFor(width, height int) Layout { return LaidOut(width, height, Sizes{}) }

// LaidOut works out every size at a given window, each column and lane given its share of the widths a person dragged.
func LaidOut(width, height int, by Sizes) Layout {
	room := width - margin - margin - (engine.Columns-1)*gap
	columns := sharesOf(room, by.Column[:])

	usable := max(0, height-margin-margin-tabsHigh-headHigh-barHigh-gap-gap)
	top, band := split(usable-gap, by.Band)

	var l Layout
	l.Tabs = Box{X: margin, Y: margin, W: width - margin - margin, H: tabsHigh}
	x := margin
	for i := range engine.Columns {
		l.Head[i] = Box{X: x, Y: margin + tabsHigh + gap, W: columns[i], H: headHigh}
		l.Column[i] = Box{X: x, Y: l.Head[i].Y + headHigh, W: columns[i], H: top}
		x += columns[i] + gap
	}
	last := l.Column[engine.Columns-1]
	l.Band = Box{
		X: l.Column[1].X,
		Y: l.Column[1].Y + top + gap,
		W: last.X + last.W - l.Column[1].X,
		H: band,
	}
	l.Column[0].H = top + gap + band

	lanes := sharesOf(l.Band.W-(len(engine.Stalled)-1)*gap, by.Lane[:])
	x = l.Band.X
	for i := range l.Lane {
		l.Lane[i] = Box{X: x, Y: l.Band.Y + laneHigh, W: lanes[i], H: max(0, band-laneHigh)}
		x += lanes[i] + gap
	}
	l.Bar = Box{X: margin, Y: height - margin - barHigh, W: width - margin - margin, H: barHigh}
	return l
}

// LowestBand is the shortest the edge above the band, dragged, leaves it: its lanes' names and a card's room under them. The columns above it keep a card's room too.
const LowestBand = laneHigh + CardHeight

func split(room int, given [2]int) (top, band int) {
	band = bandHigh
	if total := given[0] + given[1]; total > 0 {
		band = room * given[1] / total
	}
	band = max(min(band, room-CardHeight), LowestBand)
	return room - band, band
}

func sharesOf(room int, widths []int) []int {
	out := make([]int, len(widths))
	total := 0
	for _, w := range widths {
		total += w
	}
	if total == 0 {
		for i := range out {
			out[i] = room / len(out)
		}
		return out
	}
	given := 0
	for i, w := range widths {
		out[i] = room * w / total
		given += out[i]
	}
	out[len(out)-1] += room - given
	return out
}

// Gutter is the gap right of a column, from the top of its head to its foot: what is dragged to widen it and narrow the one beside it. Right of the backlog it runs beside the band too, since the band begins where the todo does.
func Gutter(l Layout, column int) Box {
	left, right := l.Column[column], l.Column[column+1]
	return Box{X: left.X + left.W, Y: l.Head[column].Y, W: right.X - left.X - left.W, H: left.Y + left.H - l.Head[column].Y}
}

// LaneGutter is the gap right of a lane of the band, the band's height.
func LaneGutter(l Layout, lane int) Box {
	left, right := l.Lane[lane], l.Lane[lane+1]
	return Box{X: left.X + left.W, Y: l.Band.Y, W: right.X - left.X - left.W, H: l.Band.H}
}

// BandEdge is the gap between the columns and the band, as wide as the band: what is dragged to make the band taller and the columns above it shorter.
func BandEdge(l Layout) Box {
	return Box{X: l.Band.X, Y: l.Column[1].Y + l.Column[1].H, W: l.Band.W, H: l.Band.Y - l.Column[1].Y - l.Column[1].H}
}

// Raised is the heights of the columns above the band and of the band after the edge between them was dragged up by so much: the band takes what the columns give, neither going under what it must hold.
func Raised(top, band, by int) [2]int {
	pair := top + band
	band = max(min(band+by, pair-CardHeight), LowestBand)
	return [2]int{pair - band, band}
}

// Widened is the widths after the gutter right of one of these boxes was dragged by so much: the box on its left takes what the one on its right gives, neither going under NarrowestColumn, and every other keeps what it had.
func Widened(boxes []Box, at, by int) []int {
	widths := make([]int, len(boxes))
	for i, b := range boxes {
		widths[i] = b.W
	}
	pair := widths[at] + widths[at+1]
	widths[at] = max(min(widths[at]+by, pair-NarrowestColumn), NarrowestColumn)
	widths[at+1] = pair - widths[at]
	return widths
}

const bandHigh = 214

// CardAt is where the n-th card of a column sits, given how far that column has been scrolled.
func (l Layout) CardAt(column, at, scrolled int) Box {
	body := l.Column[column]
	return Box{X: body.X, Y: body.Y + at*CardHeight - scrolled, W: body.W, H: CardHeight}
}

// LaneCardAt is the same for a card in one of the band's three lanes.
func (l Layout) LaneCardAt(lane, at, scrolled int) Box {
	body := l.Lane[lane]
	return Box{X: body.X, Y: body.Y + at*CardHeight - scrolled, W: body.W, H: CardHeight}
}

// Under says which column a point falls in and which card of it, given how far that column is scrolled.
func (l Layout) Under(p Point, scrolled int) (column, at int, found bool) {
	for i := range engine.Columns {
		if !l.Column[i].In(p) {
			continue
		}
		return i, (p.Y - l.Column[i].Y + scrolled) / CardHeight, true
	}
	return 0, 0, false
}

// LaneUnder says which of the band's lanes a point falls in.
func (l Layout) LaneUnder(p Point) (lane int, found bool) {
	for i := range l.Lane {
		if l.Lane[i].In(p) {
			return i, true
		}
	}
	return 0, false
}
