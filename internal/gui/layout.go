// Package gui draws cairn in a window a mouse can reach.
package gui

import (
	"cairn/internal/engine"
	"fmt"
	"strings"
)

// Narrowest and Shortest are the window this board needs.
const (
	Narrowest = 1024
	Shortest  = 640
)

// CardHeight is what every card takes, whatever it carries.
const CardHeight = 96

// SubHeight is what a part of the work takes in the box under a task's fields: one line and the padding round it.
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

// LayoutFor works out every size at a given window.
func LayoutFor(width, height int) Layout {
	column := (width - margin - margin - (engine.Columns-1)*gap) / engine.Columns

	usable := max(0, height-margin-margin-tabsHigh-headHigh-barHigh-gap-gap)
	band := bandHigh
	top := usable - band - gap

	var l Layout
	l.Tabs = Box{X: margin, Y: margin, W: width - margin - margin, H: tabsHigh}
	for i := range engine.Columns {
		x := margin + i*(column+gap)
		l.Head[i] = Box{X: x, Y: margin + tabsHigh + gap, W: column, H: headHigh}
		l.Column[i] = Box{X: x, Y: l.Head[i].Y + headHigh, W: column, H: top}
	}
	l.Band = Box{
		X: l.Column[1].X,
		Y: l.Column[1].Y + top + gap,
		W: column*(engine.Columns-1) + (engine.Columns-2)*gap,
		H: band,
	}
	l.Column[0].H = top + gap + band

	lane := (l.Band.W - (len(engine.Stalled)-1)*gap) / len(engine.Stalled)
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

// Showing is how many whole cards a column has room for.
func (l Layout) Showing(column int) int { return l.Column[column].H / CardHeight }

// Written is the whole shape as a person can read it, so a test can pin it and a sum changed by accident shows up as a number that moved.
func (l Layout) Written() string {
	var out strings.Builder
	say := func(name string, b Box) {
		fmt.Fprintf(&out, "%-12s x=%-5d y=%-5d w=%-5d h=%d\n", name, b.X, b.Y, b.W, b.H)
	}
	say("tabs", l.Tabs)
	for i := range engine.Columns {
		say(fmt.Sprintf("head %d", i), l.Head[i])
		say(fmt.Sprintf("column %d", i), l.Column[i])
	}
	say("band", l.Band)
	for i, name := range engine.Stalled {
		say("lane "+name, l.Lane[i])
	}
	say("bar", l.Bar)
	return out.String()
}
