package tui

import (
	"cairn/internal/board"
	"cairn/internal/tui/geom"

	tea "charm.land/bubbletea/v2"
)

func lastColumn() int { return board.Columns - 1 }

func underTheBand(column int) bool { return column > 0 }

func (b Board) moveCursor(way string) Board {
	if b.tab == board.Objectives {
		return b.walkGoals(way)
	}
	if b.tab == board.Series {
		return b.walkSeries(way)
	}
	return b.stepTo(way)
}

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

func (b Board) toColumn(to int) Board {
	if to < 0 || to > lastColumn() {
		return b
	}
	was := max(0, b.at(b.here()))
	b.column, b.inBand = to, false
	b.task = b.standingAt(was)
	return b.follow()
}

func (b Board) stepTo(way string) Board {
	if b.task == 0 {
		return b.fromNothing(way)
	}
	across, down := map[string]int{goLeft: -1, goRight: 1}[way], map[string]int{goUp: -1, goDown: 1}[way]
	c, spot, moved := board.Walk(b.cards(), b.archived, b.task, across, down)
	if !moved {
		return b
	}
	if spot.InBand && !b.inBand {
		b.cameFrom = b.column
	}
	b.task = c.Number
	return b.followed().follow()
}

func (b Board) fromNothing(way string) Board {
	switch way {
	case goLeft:
		if b.inBand {
			return b.toLane(b.column - 1)
		}
		return b.toColumn(b.column - 1)
	case goRight:
		if b.inBand {
			return b.toLane(b.column + 1)
		}
		return b.toColumn(b.column + 1)
	case goDown:
		return b.intoBand()
	case goUp:
		return b.outOfBand()
	}
	return b
}

func (b Board) toLane(column int) Board {
	if column < 1 || column > lastColumn() {
		return b
	}
	b.column = column
	b.task = b.standingAt(0)
	return b.follow()
}

func (b Board) intoBand() Board {
	if b.inBand {
		return b
	}
	b.cameFrom = b.column
	b.inBand, b.column = true, max(1, b.column)
	b.task = b.standingAt(0)
	return b.follow()
}

func (b Board) outOfBand() Board {
	if !b.inBand {
		return b
	}
	b.inBand = false
	b.task = b.standingAt(len(b.here()) - 1)
	return b.follow()
}

func (b Board) follow() Board {
	l := geom.LayoutFor(b.width, b.height)
	if b.tab == board.Objectives {
		b.goalAt = board.InSight(b.goalPlace(), goalHeight, l.Top+l.Rows+2, b.goalAt)
		return b
	}
	here := b.here()
	at := max(0, b.at(here))
	if b.inBand {
		b.laneAt[b.column-1] = geom.Scrolled(b.laneAt[b.column-1], at, l.Rows)
		return b
	}
	top := l.Top
	if b.inside != 0 {
		top = max(0, top-geom.InsideRows)
	}
	b.offsets[b.column] = geom.Scrolled(b.offsets[b.column], at, top)
	return b
}

func (b Board) readOn(way string) Board {
	if way == goDown {
		b.helpAt++
	}
	if way == goUp {
		b.helpAt = max(0, b.helpAt-1)
	}
	return b
}

const goalHeight = 2

func (b Board) goalPlace() int {
	for i, g := range b.goals() {
		if g.Number == b.goal {
			return i
		}
	}
	return 0
}

func (b Board) toEnd() Board { return b.toEdge(func(n int) int { return n - 1 }) }

func (b Board) toStart() Board { return b.toEdge(func(int) int { return 0 }) }

func (b Board) toEdge(which func(int) int) Board {
	if b.tab == board.Objectives {
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

func (b Board) at(here []board.Card) int {
	for i, c := range here {
		if c.Number == b.task {
			return i
		}
	}
	return -1
}

func (b Board) held() (board.Card, bool) {
	if b.tab == board.Series {
		return b.snap.Card(b.task)
	}
	for _, c := range b.cards() {
		if c.Number == b.task {
			return c, true
		}
	}
	return board.Card{}, false
}

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

func (b Board) pushDown() (tea.Model, tea.Cmd) {
	if b.inBand {
		if where, ok := board.Step(b.here(), b.task, 1); ok {
			return b, b.moving(where)
		}
		return b, nil
	}
	here := b.here()
	if where, ok := board.Step(here, b.task, 1); ok {
		return b, b.moving(where)
	}
	if b.at(here) < 0 {
		return b, nil
	}
	b.choosing, b.choice = true, 0
	return b, nil
}

func (b Board) pushUp() (tea.Model, tea.Cmd) {
	if b.inBand {
		if where, ok := board.Step(b.here(), b.task, -1); ok {
			return b, b.moving(where)
		}
		return b.land(board.StateAt(b.cameFrom), b.cameFrom, false)
	}
	if where, ok := board.Step(b.here(), b.task, -1); ok {
		return b, b.moving(where)
	}
	return b, nil
}

func (b Board) sideways(way string) (tea.Model, tea.Cmd) {
	by := -1
	if way == goRight {
		by = 1
	}
	state, ok := board.Beside(b.column, b.inBand, by)
	if !ok {
		return b, nil
	}
	return b.land(state, board.IndexOf(state), false)
}

type landing struct {
	column int
	inBand bool
}

func (b Board) land(state string, column int, inBand bool) (tea.Model, tea.Cmd) {
	b.leftFrom = landing{column: b.column, inBand: b.inBand}
	b.column, b.inBand = column, inBand
	return b, b.setState(state)
}
