package tui

import (
	"cairn/internal/engine"

	tea "charm.land/bubbletea/v2"
)

func lastColumn() int { return engine.Columns - 1 }

func underTheBand(column int) bool { return column > 0 }

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
	if b.inBand {
		return b
	}
	if to < 0 || to > lastColumn() {
		return b
	}
	b.column = to
	return b.follow()
}

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

func (b Board) stepTo(way string) Board {
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

func wide() int { return engine.Columns - 1 }

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

func (b Board) intoBand() Board {
	b.cameFrom = b.column
	b.inBand = true
	band := b.here()
	if len(band) == 0 {
		b.mode, b.task = Normal, 0
		return b.follow()
	}
	b.task = band[0].Number
	return b.follow()
}

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

func (b Board) follow() Board {
	l := LayoutFor(b.width, b.height)
	if b.tab == Objectives {
		b.goalAt = Scrolled(b.goalAt, b.goalPlace(), l.Top+l.Rows+2)
		return b
	}
	here := b.here()
	at := max(0, b.at(here))
	if b.inBand {
		b.bandAt = Scrolled(b.bandAt, at/(engine.Columns-1), l.Rows)
		return b
	}
	top := l.Top
	if b.inside != 0 {
		top = max(0, top-insideRows)
	}
	b.offsets[b.column] = Scrolled(b.offsets[b.column], at, top)
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

func (b Board) goalPlace() int {
	for i, g := range b.goals() {
		if g.Number == b.goal {
			return i
		}
	}
	return 0
}

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

func (b Board) toEnd() Board { return b.toEdge(func(n int) int { return n - 1 }) }

func (b Board) toStart() Board { return b.toEdge(func(int) int { return 0 }) }

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

func (b Board) at(here []engine.Card) int {
	for i, c := range here {
		if c.Number == b.task {
			return i
		}
	}
	return -1
}

func (b Board) held() (engine.Card, bool) {
	for _, c := range b.cards() {
		if c.Number == b.task {
			return c, true
		}
	}
	return engine.Card{}, false
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
		return b, nil
	}
	here := b.here()
	if where, ok := engine.Step(here, b.task, 1); ok {
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
		return b.land(engine.StateAt(b.cameFrom), b.cameFrom, false)
	}
	if where, ok := engine.Step(b.here(), b.task, -1); ok {
		return b, b.moving(where)
	}
	return b, nil
}

func (b Board) sideways(way string) (tea.Model, tea.Cmd) {
	by := -1
	if way == goRight {
		by = 1
	}
	state, ok := engine.Beside(b.column, b.inBand, by)
	if !ok {
		return b, nil
	}
	return b.land(state, engine.IndexOf(state), false)
}

func (b Board) land(state string, column int, inBand bool) (tea.Model, tea.Cmd) {
	b.column, b.inBand = column, inBand
	return b, b.setState(state)
}
