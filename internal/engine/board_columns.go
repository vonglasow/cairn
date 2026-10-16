package engine

import "cairn/internal/domain"

// Column is one of the four, named as the board writes it.
type Column struct {
	State string
	Name  string
	Cards []Card
}

var columns = []struct{ state, name string }{
	{string(domain.Backlog), "BACKLOG"},
	{string(domain.Todo), "TODO"},
	{string(domain.Current), "CURRENT"},
	{string(domain.Done), "DONE"},
}

// Columns is how many columns stand side by side: backlog, todo, current, done.
const Columns = 4

// Stalled are the three states the band holds, in the order it lays them out: domain's three that mean a task is not moving.
var Stalled = []string{string(domain.Blocked), string(domain.Watching), string(domain.Paused)}

// Band is the name the three stopped states share.
const Band = "STALLED"

// Deal puts every card where it belongs: in the column of its state, or in the band if its state is one of the three that are not moving.
func Deal(cards []Card, withArchived bool) ([]Column, []Card) {
	out := make([]Column, len(columns))
	for i, c := range columns {
		out[i] = Column{State: c.state, Name: c.name}
	}
	var band []Card
	for _, card := range OnTheBoard(cards, withArchived) {
		if Stopped(card.State) {
			band = append(band, card)
			continue
		}
		for i := range out {
			if out[i].State == card.State {
				out[i].Cards = append(out[i].Cards, card)
				break
			}
		}
	}
	return out, band
}

// At answers the column of a given state, and whether there is one.
func At(all []Column, state string) (Column, bool) {
	for _, c := range all {
		if c.State == state {
			return c, true
		}
	}
	return Column{}, false
}

// IndexOf answers where a column stands, left to right.
func IndexOf(state string) int {
	for i, c := range columns {
		if c.state == state {
			return i
		}
	}
	return -1
}

// StateAt answers which column stands at a place, left to right.
func StateAt(i int) string {
	if i < 0 || i >= len(columns) {
		return ""
	}
	return columns[i].state
}

// OnTheBoard is what a board lays out: every task, the archived ones only when asked for.
func OnTheBoard(cards []Card, withArchived bool) []Card {
	var laid []Card
	for _, c := range cards {
		if c.Archived && !withArchived {
			continue
		}
		laid = append(laid, c)
	}
	return laid
}

// Found is what a board shows of a search: the tasks a query reaches, a word found only in a step reaching the task it belongs to.
func Found(cards []Card, query string, withArchived bool) []Card {
	if query == "" {
		return OnTheBoard(cards, withArchived)
	}
	reached := map[int]bool{}
	for _, c := range cards {
		if !ReachesCard(c, query) {
			continue
		}
		reached[c.Number] = true
	}
	var found []Card
	for _, c := range OnTheBoard(cards, withArchived) {
		if reached[c.Number] {
			found = append(found, c)
		}
	}
	return found
}
