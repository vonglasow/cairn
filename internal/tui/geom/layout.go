// Package geom says where the terminal lays things out, in cells: the columns, the band, how tall each is.
package geom

import "cairn/internal/board"

// Narrowest and Shortest are the two numbers written in rather than worked out.
const (
	Narrowest = 60
	Shortest  = 26
)

// CardHeight is the six rows every card takes, whatever it carries.
const CardHeight = 6

// Margin is the cell left at the board's edge, and Frame the cells a card's frame and padding take across.
const (
	Margin = 1
	Frame  = 4
)

// Fixed is the rows the tabs, the heads, the rules and the foot take, whatever the height.
const Fixed = 8

// InsideRows is the rows an objective opened takes above its columns.
const InsideRows = 3

// BandHead is the rows above the cards of a lane: the band's name, the lane's and its rule.
const BandHead = 3

// BandCards is how many cards a lane shows at most, one above the other.
const BandCards = 2

// Layout is every size the board draws from.
type Layout struct {
	Column   int
	Interior int
	Top      int
	Rows     int
}

// LayoutFor works out every size at a given window.
func LayoutFor(width, height int) Layout {
	column := (width - Margin - (board.Columns - 1)) / board.Columns
	return Layout{
		Column:   column,
		Interior: column - Frame,
		Top:      topRows(height),
		Rows:     max(1, bandRows(height)-1),
	}
}

func bandRows(height int) int {
	return min(BandCards*CardHeight+1, max(0, height-Fixed)/2)
}

func topRows(height int) int { return max(0, height-Fixed) - bandRows(height) }

// TooSmall says whether the board can be drawn in this window at all.
func TooSmall(width, height int) bool {
	return width < Narrowest || height < Shortest
}

// Scrolled says how far down a run of cards has to be pushed for the one at `at` to be wholly on screen, given where it is pushed to now.
func Scrolled(offset, at, rows int) int { return board.InSight(at, CardHeight, rows, offset) }
