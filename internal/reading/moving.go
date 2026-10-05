package reading

// Where a card goes when it is pushed. Both boards move a card the same way —
// a place up or down among the ones it sits with, a column over sideways — and
// each had worked it out on its own. The window's copy asked to put a card
// behind itself whenever it went up from anywhere but second, and took a card
// in the band for the first card of the first column.

import "strconv"

// Step is what moving a card by one place among the ones it sits with asks of
// the engine: to go behind the card that will then be before it, or to the head
// of them. False when there is nowhere to go — the head going up, the last
// going down, a card that is not among them, a step of nothing.
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
	// Going down, the card lands behind the one it passes; going up, behind
	// the one before the one it passes. Written as one sum rather than a test
	// on the direction: with a step of nothing refused above, `by < 0` and
	// `by <= 0` say the same thing, and a boundary no test can tell apart is a
	// mutant designed out rather than chased.
	behind := to - 1 + min(max(by, 0), 1)
	return map[string]string{"after": strconv.Itoa(here[behind].Number)}, true
}

// Beside is the state of the column one over from where a card stands, and
// whether there is one. Out of the band, left is the backlog — the band does
// not lie under it, but it is still what is to the band's left — and there is
// nothing to the band's right.
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

// Where finds a card on the board: the column it stands in, or the band, and
// its place there.
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

// InLanes lays the band out in its lanes: one per stopped state, in the order
// the band holds them, each in the order its cards were dealt.
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
