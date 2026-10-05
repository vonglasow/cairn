package reading

import "cairn/internal/engine"

// FocusOp is what one key does to a card's focus: takes it up, or lets it go
// when it is already being worked on. Both name the task, since several can be
// in focus at once and letting go of nothing in particular would let go of
// every one of them.
func FocusOp(c Card) engine.Operation {
	if c.InFocus {
		return engine.TaskUnfocus
	}
	return engine.TaskFocus
}

// AwayOp is what one key does to a card's place on the board: takes it off, or
// brings it back when it is already off.
func AwayOp(c Card) engine.Operation {
	if c.Archived {
		return engine.TaskUnarchive
	}
	return engine.TaskArchive
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
