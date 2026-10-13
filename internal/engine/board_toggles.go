package engine

import "strings"

// FocusOp is what one key does to a card's focus: takes it up, or lets it go when it is already being worked on.
func FocusOp(c Card) Operation {
	if c.InFocus {
		return TaskUnfocus
	}
	return TaskFocus
}

// AwayOp is what one key does to a card's place on the board: takes it off, or brings it back when it is already off.
func AwayOp(c Card) Operation {
	if c.Archived {
		return TaskUnarchive
	}
	return TaskArchive
}

// Offered is the word a board offers an act by: its own verb, so what is offered and what it does are the one decision.
func (o Operation) Offered() string {
	if o.Verb == "" {
		return ""
	}
	return strings.ToUpper(o.Verb[:1]) + o.Verb[1:]
}

// Card is the card numbered n in what was read, and whether there is one.
func (s Snapshot) Card(n int) (Card, bool) {
	for _, c := range s.Cards {
		if c.Number == n {
			return c, true
		}
	}
	return Card{}, false
}

// Goal is the objective numbered n in what was read, and whether there is one.
func (s Snapshot) Goal(n int) (Goal, bool) {
	for _, g := range s.Goals {
		if g.Number == n {
			return g, true
		}
	}
	return Goal{}, false
}
