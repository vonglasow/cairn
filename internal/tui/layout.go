package tui

import "cairn/internal/engine"

// How wide and how tall everything is, worked out from the window the terminal
// reports and worked out again every time it changes. Nothing about a size is
// written into a card or a column — a board that carried its own sizes would be
// a board that lies the moment somebody drags a window.

// Narrowest and Shortest are the two numbers written in rather than worked out.
// Below them four columns leave ten characters for a title and the band has
// nowhere to be, so the board says what it needs instead of drawing something
// unreadable.
const (
	Narrowest = 60
	Shortest  = 26
)

// CardHeight is the six rows every card takes, whatever it carries. Fixed on
// purpose: it makes moving the cursor arithmetic rather than measurement, and
// it lets a column say how much of itself is off screen without drawing one.
const CardHeight = 6

// margin is the one blank column down the left of the whole board, and frame is
// what a card's border and padding take out of its width. Named rather than
// spelled into the sums, because the sums are what has to be checked against a
// drawing.
const (
	margin = 1
	frame  = 4 // two borders and two paddings
)

// fixed is every row the board spends before a single card is drawn: the tabs,
// a blank line, the column headers and their rule, the band's header and its
// rule, another blank line, and the bar.
const fixed = 8

// insideRows is what an objective's own card costs the columns when its board
// is open: two rows for the card and the blank line under it. Named, because it
// is subtracted in two places and a board that took two off in one and three in
// the other would draw one row more than the window has.
const insideRows = 3

// bandCards is how many rows of cards the band shows when the window can spare
// them. Two rather than one: three cards across is a row, and one row makes the
// band a strip that has to be scrolled to be read at all.
const bandCards = 2

// Layout is every size the board draws from.
type Layout struct {
	// Column is the full width of one column, borders included.
	Column int
	// Interior is what a card has left for its own text.
	Interior int
	// Band is how wide the stalled band is: the last three columns and the gaps
	// between them, because it lies under those three and not under the
	// backlog.
	Band int
	// Top is how many rows the columns get, and Rows how many the band gets.
	// The band is never given nothing: what is not moving is the thing most
	// worth seeing, and a band pushed off the bottom of the screen is a band
	// that might as well not be there.
	Top  int
	Rows int
}

// LayoutFor works out every size at a given window.
func LayoutFor(width, height int) Layout {
	column := (width - margin - (engine.Columns - 1)) / engine.Columns
	return Layout{
		Column:   column,
		Interior: column - frame,
		Band:     column*(engine.Columns-1) + (engine.Columns - 2),
		Top:      topRows(height),
		Rows:     bandRows(height),
	}
}

// bandRows is how many rows the band gets: one card's worth when the window can
// spare it, and half of what is left when it cannot. Never all of it, and never
// none.
func bandRows(height int) int {
	// Written as a minimum rather than a branch: the branch would have had a
	// boundary nothing could see, since at exactly the share it wants both
	// sides of it answer the same thing.
	//
	// It asks for two rows of cards and settles for half the window, so the
	// columns always keep the other half however small the window is.
	return min(bandCards*CardHeight, max(0, height-fixed)/2)
}

// topRows is what the columns get: everything the band did not take.
func topRows(height int) int { return max(0, height-fixed) - bandRows(height) }

// TooSmall says whether the board can be drawn in this window at all.
func TooSmall(width, height int) bool {
	return width < Narrowest || height < Shortest
}

// Showing is how many whole cards a run of rows has room for.
func Showing(rows int) int { return rows / CardHeight }

// Scrolled says how far down a run of cards has to be pushed for the one at
// `at` to be wholly on screen, given where it is pushed to now.
//
// Written without a branch, and not for brevity: every way of spelling this
// with ifs has a boundary that both sides of agree on — at exactly one card's
// worth, at exactly the top of the window — so a wrong comparison would go out
// unjudged. A minimum and a maximum have no boundary to get wrong.
func Scrolled(offset, at, rows int) int {
	top := at * CardHeight
	return min(top, max(offset, top+CardHeight-rows))
}
