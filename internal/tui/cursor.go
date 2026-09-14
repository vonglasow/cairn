package tui

import (
	"cairn/internal/reading"

	tea "charm.land/bubbletea/v2"
)

// Where the cursor is, and what moves it. In normal it stands on a column; in
// insert it holds a task. The two are different enough that the same key means
// two different journeys, and near enough that both are spelled hjkl.

// lastColumn is the rightmost, done. A function rather than a constant, so the
// gate can judge the arithmetic: a package-level constant sits outside every
// covered block, and a wrong one would go out unjudged.
func lastColumn() int { return reading.Columns - 1 }

// underTheBand says whether the stalled band lies beneath a column. It does
// not lie under the backlog, which runs the whole height instead.
func underTheBand(column int) bool { return column > 0 }

// moveCursor walks the cursor: between objectives on the second tab, and
// between columns — and down into the band or back out of it — on the first.
func (b Board) moveCursor(way string) Board {
	if b.tab == Objectives {
		return b.walkGoals(way)
	}
	switch way {
	case goLeft:
		return b.toColumn(b.column - 1)
	case goRight:
		return b.toColumn(b.column + 1)
	case goDown:
		if !b.inBand && underTheBand(b.column) {
			b.cameFrom = b.column
			b.inBand = true
			return b.follow()
		}
		return b
	case goUp:
		if b.inBand {
			b.inBand = false
			b.column = b.cameFrom
			return b.follow()
		}
		return b
	}
	return b
}

// walkGoals moves up and down the objectives. Left and right do nothing there:
// the list is one column, and a key that pointed sideways would be a key that
// looked broken.
func (b Board) walkGoals(way string) Board {
	goals := b.goals()
	if len(goals) == 0 {
		return b
	}
	to := b.goalPlace()
	if way == goUp {
		to--
	}
	if way == goDown {
		to++
	}
	if to < 0 || to >= len(goals) {
		return b
	}
	b.goal = goals[to].Number
	return b.follow()
}

// toColumn puts the cursor on a column, and refuses to walk off either end —
// a cursor that wrapped round would move the eye across the whole screen for
// one keystroke.
func (b Board) toColumn(to int) Board {
	if b.inBand {
		// Left and right inside the band walk the cards, not the columns: the
		// band is one place, however wide it is.
		return b
	}
	if to < 0 || to > lastColumn() {
		return b
	}
	b.column = to
	return b.follow()
}

// takeHold puts the cursor on a task, if there is one where it stands. On an
// empty column there is nothing to hold, so the board stays in normal rather
// than pretending.
func (b Board) takeHold() Board {
	if b.tab == Objectives {
		return b.openObjective()
	}
	here := b.here()
	if len(here) == 0 {
		return b
	}
	b.mode = Insert
	b.task = here[0].Number
	return b.follow()
}

// stepTo moves the cursor from one task to another.
func (b Board) stepTo(way string) Board {
	// The band is a grid, and a grid is walked as one. Everywhere else a column
	// is a list, and left and right change which list.
	if b.inBand {
		return b.stepInBand(way)
	}
	switch way {
	case goUp, goDown:
		return b.stepBy(way)
	case goLeft:
		return b.acrossTo(b.column - 1)
	case goRight:
		return b.acrossTo(b.column + 1)
	}
	return b
}

// wide is how many cards the band puts in a row.
func wide() int { return reading.Columns - 1 }

// stepInBand walks the band as what it is drawn as: a grid. Down goes down a
// row and right goes along one, because a key that pointed down and moved the
// cursor sideways would be a key nobody could trust.
//
// Up off the top row leaves the band, since that is what is above it.
func (b Board) stepInBand(way string) Board {
	band := b.here()
	at := b.at(band)
	if at < 0 {
		return b
	}
	to := at
	switch way {
	case goLeft:
		if at%wide() == 0 {
			return b
		}
		to = at - 1
	case goRight:
		if at%wide() == wide()-1 {
			return b
		}
		to = at + 1
	case goDown:
		to = at + wide()
	case goUp:
		if at < wide() {
			return b.outOfBand()
		}
		to = at - wide()
	}
	if to < 0 || to >= len(band) {
		return b
	}
	b.task = band[to].Number
	return b.follow()
}

// stepBy moves up or down the list the cursor is standing in.
//
// Off the bottom of a column it carries on into the band, and off the top of
// the band it comes back up: the band is under the columns on screen, so down
// out of one has to reach it. A column that stopped at its last card would
// leave what is stalled unreachable whenever a column happened to be short.
func (b Board) stepBy(way string) Board {
	here := b.here()
	at := b.at(here)
	if at < 0 {
		return b
	}
	to := at + 1
	if way == goUp {
		to = at - 1
	}
	if to >= 0 && to < len(here) {
		b.task = here[to].Number
		return b.follow()
	}
	// Off an end. Which end is the direction, not the number: a number would
	// carry a comparison nothing can tell the two sides of apart.
	if way == goUp {
		if b.inBand {
			return b.outOfBand()
		}
		return b
	}
	if !b.inBand && underTheBand(b.column) {
		return b.intoBand()
	}
	return b
}

// intoBand puts the cursor on the first thing that is not moving.
func (b Board) intoBand() Board {
	b.cameFrom = b.column
	b.inBand = true
	band := b.here()
	if len(band) == 0 {
		// Nothing is stalled, so there is nothing to hold: the cursor stands on
		// the band itself rather than on a card in it.
		b.mode, b.task = Normal, 0
		return b.follow()
	}
	b.task = band[0].Number
	return b.follow()
}

// outOfBand brings the cursor back up to the column it came down from.
func (b Board) outOfBand() Board {
	b.inBand = false
	b.column = b.cameFrom
	here := b.here()
	if len(here) == 0 {
		b.mode, b.task = Normal, 0
		return b.follow()
	}
	b.task = here[len(here)-1].Number
	return b.follow()
}

// follow scrolls whatever the cursor is standing in so that what it holds is
// wholly on screen. It is called after every move, so nothing the cursor can
// reach is ever hidden.
func (b Board) follow() Board {
	l := LayoutFor(b.width, b.height)
	if b.tab == Objectives {
		b.goalAt = Scrolled(b.goalAt, b.goalPlace(), l.Top+l.Rows+2)
		return b
	}
	here := b.here()
	at := max(0, b.at(here))
	if b.inBand {
		// The band lays its cards out three across, so what scrolls is the row
		// a card is in and not the card itself.
		b.bandAt = Scrolled(b.bandAt, at/(reading.Columns-1), l.Rows)
		return b
	}
	top := l.Top
	if b.inside != 0 {
		top = max(0, top-insideRows)
	}
	b.offsets[b.column] = Scrolled(b.offsets[b.column], at, top)
	return b
}

// readOn moves the help down or up, a line at a time. Something longer than the
// window has to be readable to its end, and the help is the one screen with
// nothing in it to put a cursor on.
func (b Board) readOn(way string) Board {
	if way == goDown {
		b.helpAt++
	}
	if way == goUp {
		b.helpAt = max(0, b.helpAt-1)
	}
	return b
}

// goalPlace is where in the list the objective under the cursor stands.
func (b Board) goalPlace() int {
	for i, g := range b.goals() {
		if g.Number == b.goal {
			return i
		}
	}
	return 0
}

// acrossTo moves to another column and stays at the same level in it. A column
// with nothing in it is still moved to, so the cursor can walk past an empty
// column rather than being stopped by one.
//
// The level is kept rather than reset to the top: walking sideways along a
// board is reading across it, and a cursor that jumped back to the first card
// every time would make the third task in one column and the third in the next
// two separate journeys.
func (b Board) acrossTo(to int) Board {
	if b.inBand || to < 0 || to > lastColumn() {
		return b
	}
	was := max(0, b.at(b.here()))
	b.column = to
	here := b.here()
	if len(here) == 0 {
		b.mode = Normal
		b.task = 0
		return b.follow()
	}
	b.task = here[min(was, len(here)-1)].Number
	return b.follow()
}

// toEnd puts the cursor on the last of whatever it is standing in, and toStart
// on the first. `G` and `gg`, which is what anybody who has used vim will try.
func (b Board) toEnd() Board { return b.toEdge(func(n int) int { return n - 1 }) }

// toStart puts it on the first.
func (b Board) toStart() Board { return b.toEdge(func(int) int { return 0 }) }

// toEdge is both of them: it asks where to go given how many there are.
func (b Board) toEdge(which func(int) int) Board {
	if b.tab == Objectives {
		goals := b.goals()
		if len(goals) == 0 {
			return b
		}
		b.goal = goals[which(len(goals))].Number
		return b.follow()
	}
	here := b.here()
	if len(here) == 0 {
		return b
	}
	b.task = here[which(len(here))].Number
	return b.follow()
}

// at answers where in a list the cursor's task sits, or that it is not in it.
func (b Board) at(here []reading.Card) int {
	for i, c := range here {
		if c.Number == b.task {
			return i
		}
	}
	return -1
}

// held answers the card the cursor is holding.
func (b Board) held() (reading.Card, bool) {
	for _, c := range b.cards() {
		if c.Number == b.task {
			return c, true
		}
	}
	return reading.Card{}, false
}

// push moves the task itself. Sideways it changes column, and the state follows
// — where a task sits and what state it is in are the same fact drawn two ways.
// Up and down it changes place among the ones it sits with, which is a
// different thing and never touches the column.
func (b Board) push(way string) (tea.Model, tea.Cmd) {
	if _, ok := b.held(); !ok {
		return b, nil
	}
	switch way {
	case goDown:
		return b.pushDown()
	case goUp:
		return b.pushUp()
	case goLeft, goRight:
		return b.sideways(way)
	}
	return b, nil
}

// pushDown puts the task one place lower among the ones it sits with. From the
// last of them there is nowhere lower in the column, and what is below the
// column is the band — so that is the one case where it asks which kind of
// stopped the task has become.
func (b Board) pushDown() (tea.Model, tea.Cmd) {
	if b.inBand {
		return b, nil
	}
	here := b.here()
	at := b.at(here)
	if at < 0 {
		return b, nil
	}
	if at < len(here)-1 {
		return b, b.moveBehind(here[at+1].Number)
	}
	// Which of the three kinds of stopped this is cannot be worked out, only
	// asked. Nothing is written until the answer comes back.
	b.choosing, b.choice = true, 0
	return b, nil
}

// pushUp puts the task one place higher, and out of the band back into the
// column it was pushed down from.
func (b Board) pushUp() (tea.Model, tea.Cmd) {
	if b.inBand {
		return b.land(reading.StateAt(b.cameFrom), b.cameFrom, false)
	}
	here := b.here()
	at := b.at(here)
	if at < 1 {
		return b, nil
	}
	if at == 1 {
		return b, b.moveFirst()
	}
	return b, b.moveBehind(here[at-2].Number)
}

// sideways moves a task one column over. Out of the band it reaches the
// backlog, which the band does not lie under but which is still what is to its
// left; there is nothing to the band's right, so nothing happens there.
func (b Board) sideways(way string) (tea.Model, tea.Cmd) {
	if b.inBand {
		if way == goRight {
			return b, nil
		}
		return b.land(reading.StateAt(0), 0, false)
	}
	to := b.column - 1
	if way == goRight {
		to = b.column + 1
	}
	if to < 0 || to > lastColumn() {
		return b, nil
	}
	return b.land(reading.StateAt(to), to, false)
}

// land writes the new state and puts the cursor where the task went.
func (b Board) land(state string, column int, inBand bool) (tea.Model, tea.Cmd) {
	b.column, b.inBand = column, inBand
	return b, b.setState(state)
}
