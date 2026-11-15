package board

// Spot is where a card sits: its column, or the lane of the band it is in, and its rank there.
type Spot struct {
	Column, At int
	InBand     bool
}

// Walk is the card the choice lands on, moving across and down from the chosen one: across the columns or the band's lanes, down a column into the lane under it and up out of a lane into the column above; with nowhere to go it does not move.
func Walk(cards []Card, withArchived bool, chosen, across, down int) (Card, Spot, bool) {
	columns, lanes := heldIn(cards, withArchived)
	from, ok := spotIn(columns, lanes, chosen)
	if from.InBand {
		return walkingBand(columns, lanes, from, across, down)
	}
	if !ok {
		from, ok = firstSpot(columns)
	}
	if !ok {
		return Card{}, Spot{}, false
	}
	return walkingColumns(columns, lanes, from, across, down)
}

// SpotOf is where a card sits on the board, when it is there.
func SpotOf(cards []Card, withArchived bool, number int) (Spot, bool) {
	columns, lanes := heldIn(cards, withArchived)
	return spotIn(columns, lanes, number)
}

func heldIn(cards []Card, withArchived bool) (columns, lanes [][]Card) {
	dealt, stalled := Deal(cards, withArchived)
	columns = make([][]Card, len(dealt))
	for i, held := range dealt {
		columns[i] = held.Cards
	}
	return columns, InLanes(stalled)
}

func spotIn(columns, lanes [][]Card, number int) (Spot, bool) {
	if s, ok := spotOf(lanes, number); ok {
		s.InBand = true
		return s, true
	}
	return spotOf(columns, number)
}

func walkingColumns(columns, lanes [][]Card, from Spot, across, down int) (Card, Spot, bool) {
	column, at := from.Column, from.At
	if down > 0 && column > 0 && at == len(columns[column])-1 && len(lanes[column-1]) > 0 {
		return landed(lanes, Spot{Column: column - 1, InBand: true})
	}
	if across != 0 {
		for next := column + across; next >= 0 && next < len(columns); next += across {
			if len(columns[next]) > 0 {
				return landed(columns, Spot{Column: next, At: min(at, len(columns[next])-1)})
			}
		}
		return Card{}, from, false
	}
	return landed(columns, Spot{Column: column, At: min(len(columns[column])-1, max(0, at+down))})
}

func walkingBand(columns, lanes [][]Card, from Spot, across, down int) (Card, Spot, bool) {
	lane, at := from.Column, from.At
	if down < 0 && at == 0 {
		if above := lane + 1; above < len(columns) && len(columns[above]) > 0 {
			return landed(columns, Spot{Column: above, At: len(columns[above]) - 1})
		}
		return Card{}, from, false
	}
	if across != 0 {
		for next := lane + across; next >= 0 && next < len(lanes); next += across {
			if len(lanes[next]) > 0 {
				return landed(lanes, Spot{Column: next, At: min(at, len(lanes[next])-1), InBand: true})
			}
		}
		return Card{}, from, false
	}
	return landed(lanes, Spot{Column: lane, At: min(len(lanes[lane])-1, max(0, at+down)), InBand: true})
}

func landed(held [][]Card, s Spot) (Card, Spot, bool) { return held[s.Column][s.At], s, true }

func spotOf(held [][]Card, number int) (Spot, bool) {
	for i, cards := range held {
		for j, c := range cards {
			if number != 0 && c.Number == number {
				return Spot{Column: i, At: j}, true
			}
		}
	}
	return Spot{}, false
}

func firstSpot(columns [][]Card) (Spot, bool) {
	for i, cards := range columns {
		if len(cards) > 0 {
			return Spot{Column: i}, true
		}
	}
	return Spot{}, false
}

// InSight is where a run scrolled to where shows the element at, each high, the run being run long: the least scroll that shows it whole.
func InSight(at, high, run, where int) int {
	top := at * high
	return min(top, max(where, top+high-run))
}
