package tui

import "cairn/internal/engine"

// Narrowest and Shortest are the two numbers written in rather than worked out.
const (
	Narrowest = 60
	Shortest  = 26
)

// CardHeight is the six rows every card takes, whatever it carries.
const CardHeight = 6

const (
	margin = 1
	frame  = 4
)

const fixed = 8

const insideRows = 3

const bandCards = 2

// Layout is every size the board draws from.
type Layout struct {
	Column   int
	Interior int
	Band     int
	Top      int
	Rows     int
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

func bandRows(height int) int {
	return min(bandCards*CardHeight, max(0, height-fixed)/2)
}

func topRows(height int) int { return max(0, height-fixed) - bandRows(height) }

// TooSmall says whether the board can be drawn in this window at all.
func TooSmall(width, height int) bool {
	return width < Narrowest || height < Shortest
}

// Showing is how many whole cards a run of rows has room for.
func Showing(rows int) int { return rows / CardHeight }

// Scrolled says how far down a run of cards has to be pushed for the one at `at` to be wholly on screen, given where it is pushed to now.
func Scrolled(offset, at, rows int) int {
	top := at * CardHeight
	return min(top, max(offset, top+CardHeight-rows))
}
