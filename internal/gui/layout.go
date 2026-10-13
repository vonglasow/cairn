// Package gui draws cairn in a window a mouse can reach. It is a surface like
// the command line, the MCP server and the terminal board: it reads a gesture,
// calls one thing in internal/engine, and draws what comes back — it decides
// nothing.
//
// Nothing here touches a screen. Every size is worked out from the window that
// was reported, a gesture turns one state into another, and drawing reads that
// state without changing it — which is what lets a scenario click on a card and
// read what happened with no window anywhere.
package gui

import (
	"cairn/internal/engine"
	"fmt"
	"strings"
)

// Everything is in dp rather than pixels, so a Retina screen and a 96-dpi one
// give the same physical thing. Gio is told the ratio once and multiplies.

// Narrowest and Shortest are the window this board needs. Below either, it says
// so rather than drawing something unreadable.
//
// Four columns at 240dp is the width, and the height is what two rows of the
// band and three cards above them come to. Both are written in rather than
// worked out, because what makes a card readable is not something a sum knows.
const (
	Narrowest = 1024
	Shortest  = 640
)

// Columns is how many stand side by side, read off reading rather than counted
// again here.
const Columns = engine.Columns

// Lanes are the three the band holds, in the order it draws them. Naming them
// is what the terminal needs a glyph for: a card says which kind of stopped it
// is by which lane it sits in.
var Lanes = engine.Stalled

// CardHeight is what every card takes, whatever it carries. Fixed on purpose:
// it is what lets a column say how many fit and which one is under the pointer
// without drawing a thing.
const CardHeight = 96

// SubHeight is what a part of the work takes in the box under a task's fields:
// one line and the padding round it.
//
// A card is 96 and uses 50 — the rest was room for a deadline, a progress bar
// and a run of labels, none of which a sub-task carries any more: it is its
// number, its title, and whether it is done. Those fit on one line, and five
// of them fit in a box where one card did.
//
// Measured before it was chosen. The box under twelve fields is 120 pixels in
// the smallest window this board draws in, 287 in a large one; at 96 that is
// one tile and two, at 24 it is five and eleven. The same measurement said a
// second line would buy nine characters of title and cost two tiles of sight,
// so there is one.
//
// Written as the number it is rather than as lineHigh + padding: the sum says
// no more than the number does, and it is one more piece of arithmetic for a
// gate to ask about and nothing to answer with.
const SubHeight = 28

// The room everything else takes before a card is drawn.
const (
	margin    = 12 // around the whole board
	gap       = 12 // between two columns, and between the columns and the band
	tabsHigh  = 40 // the two tabs across the top
	headHigh  = 28 // a column's name and its count
	laneHigh  = 22 // a lane's name inside the band
	barHigh   = 32 // what is offered, along the bottom
	bandCards = 2  // rows of cards the band asks for when the window can spare them
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
	// Tabs is the strip across the top.
	Tabs Box
	// Head is each column's name and count; Column is the run of cards under it.
	Head   [Columns]Box
	Column [Columns]Box
	// Band is the whole stalled strip, Lane each of the three inside it.
	Band Box
	Lane [3]Box
	// Bar is what is offered, along the bottom.
	Bar Box
}

// TooSmall says whether the board can be drawn in this window at all.
func TooSmall(width, height int) bool {
	return width < Narrowest || height < Shortest
}

// LayoutFor works out every size at a given window.
func LayoutFor(width, height int) Layout {
	column := (width - margin - margin - (Columns-1)*gap) / Columns

	// What is left once the tabs, the headers and the bar have taken theirs.
	// Two gaps: one between the band of tabs and the columns, one between the
	// columns and the stalled band. The first is what keeps the tabs from
	// reading as the first row of the board.
	usable := max(0, height-margin-margin-tabsHigh-headHigh-barHigh-gap-gap)
	band := bandHigh
	top := usable - band - gap

	var l Layout
	l.Tabs = Box{X: margin, Y: margin, W: width - margin - margin, H: tabsHigh}
	for i := range Columns {
		x := margin + i*(column+gap)
		l.Head[i] = Box{X: x, Y: margin + tabsHigh + gap, W: column, H: headHigh}
		l.Column[i] = Box{X: x, Y: l.Head[i].Y + headHigh, W: column, H: top}
	}
	l.Band = Box{
		X: l.Column[1].X,
		Y: l.Column[1].Y + top + gap,
		W: column*(Columns-1) + (Columns-2)*gap,
		H: band,
	}
	// The backlog runs past where the band stops the other three: it is the
	// heap everything lands in, and it holds more than they do together.
	l.Column[0].H = top + gap + band

	lane := (l.Band.W - (len(Lanes)-1)*gap) / len(Lanes)
	for i := range l.Lane {
		l.Lane[i] = Box{
			X: l.Band.X + i*(lane+gap),
			Y: l.Band.Y + laneHigh,
			W: lane,
			H: max(0, band-laneHigh),
		}
	}
	l.Bar = Box{X: margin, Y: height - margin - barHigh, W: width - margin - margin, H: barHigh}
	return l
}

// bandHigh is what the band gets: a name for its two rows of cards and the
// lane names above them.
//
// It was a minimum of that and half of what was left, the way the terminal
// board writes it, and the mutation gate showed the half was unreachable. The
// arithmetic, so nobody puts it back: the smallest window this board draws in
// is Shortest, which leaves 640-148 = 492 once the tabs, the headers, the bar
// and the two gaps have taken theirs — and half of 492 is 246, more than the
// 214 the band asks for. A window where the half would win is a window that says it is
// too small instead.
// Written out rather than summed, for the reason test/harness.go gives: the
// arithmetic of a package-level constant is never inside a covered block, so a
// sum here would go out unjudged. It is laneHigh + bandCards*CardHeight, which
// is 22 + 2*96.
const bandHigh = 214

// CardAt is where the n-th card of a column sits, given how far that column has
// been scrolled.
func (l Layout) CardAt(column, at, scrolled int) Box {
	body := l.Column[column]
	return Box{X: body.X, Y: body.Y + at*CardHeight - scrolled, W: body.W, H: CardHeight}
}

// LaneCardAt is the same for a card in one of the band's three lanes.
func (l Layout) LaneCardAt(lane, at, scrolled int) Box {
	body := l.Lane[lane]
	return Box{X: body.X, Y: body.Y + at*CardHeight - scrolled, W: body.W, H: CardHeight}
}

// Under says which column a point falls in and which card of it, given how far
// that column is scrolled. It answers false for a point that is in no column at
// all — the gap between two of them, the bar, the tabs.
func (l Layout) Under(p Point, scrolled int) (column, at int, found bool) {
	for i := range Columns {
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

// Showing is how many whole cards a column has room for.
func (l Layout) Showing(column int) int { return l.Column[column].H / CardHeight }

// Written is the whole shape as a person can read it, so a test can pin it and
// a sum changed by accident shows up as a number that moved.
func (l Layout) Written() string {
	var out strings.Builder
	say := func(name string, b Box) {
		fmt.Fprintf(&out, "%-12s x=%-5d y=%-5d w=%-5d h=%d\n", name, b.X, b.Y, b.W, b.H)
	}
	say("tabs", l.Tabs)
	for i := range Columns {
		say(fmt.Sprintf("head %d", i), l.Head[i])
		say(fmt.Sprintf("column %d", i), l.Column[i])
	}
	say("band", l.Band)
	for i, name := range Lanes {
		say("lane "+name, l.Lane[i])
	}
	say("bar", l.Bar)
	return out.String()
}
