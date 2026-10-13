package engine

import "strconv"

// Step is what moving a card by one place among the ones it sits with asks of the engine: to go behind the card that will then be before it, or to the head of them.
func Step(here []Card, number, by int) (map[string]string, bool) {
	if by == 0 {
		return nil, false
	}
	at := -1
	for i, c := range here {
		if c.Number == number {
			at = i
		}
	}
	to := at + by
	if at < 0 || to < 0 || to >= len(here) {
		return nil, false
	}
	if to == 0 {
		return map[string]string{"first": "true"}, true
	}
	behind := to - 1 + min(max(by, 0), 1)
	return map[string]string{"after": strconv.Itoa(here[behind].Number)}, true
}

// Beside is the state of the column one over from where a card stands, and whether there is one.
func Beside(column int, inBand bool, by int) (string, bool) {
	if inBand {
		if by < 0 {
			return StateAt(0), true
		}
		return "", false
	}
	if by == 0 {
		return "", false
	}
	to := column + by
	if to < 0 || to >= Columns {
		return "", false
	}
	return StateAt(to), true
}

// Where finds a card on the board: the column it stands in, or the band, and its place there.
func Where(cards []Card, withArchived bool, number int) (column int, inBand bool, at int, found bool) {
	all, band := Deal(cards, withArchived)
	for i, col := range all {
		for j, c := range col.Cards {
			if c.Number == number {
				return i, false, j, true
			}
		}
	}
	for j, c := range band {
		if c.Number == number {
			return 0, true, j, true
		}
	}
	return 0, false, 0, false
}

// InLanes lays the band out in its lanes: one per stopped state, in the order the band holds them, each in the order its cards were dealt.
func InLanes(band []Card) [][]Card {
	lanes := make([][]Card, len(Stalled))
	for _, c := range band {
		for i, state := range Stalled {
			if c.State == state {
				lanes[i] = append(lanes[i], c)
			}
		}
	}
	return lanes
}
