package gui

import (
	"slices"
	"strconv"

	"cairn/internal/reading"
)

// What the board is showing and what a card can be asked to do: a search
// narrowing it, the other page, a task taken up or off the board, how full a
// lane or a page is.

// Showing is the work narrowed to what a search reaches, which is what
// everything draws from. With no search it is the whole reading.
//
// The narrowing is done here and not asked of the store: a search is a way of
// looking, and a board that asked the store to search would be a board that
// waits to draw.
func (w Window) Showing() reading.Snapshot {
	if w.looking.Empty() {
		return w.snap
	}
	found := reading.Snapshot{Focus: w.snap.Focus, Version: w.snap.Version}
	// What a board shows of a search: a part that matched answers as the work
	// it is part of, since a sub-task has no card of its own here to appear on.
	found.Cards = reading.Found(w.snap.Cards, w.looking.Reads(), w.archived)
	found.Goals = reading.FoundGoals(w.snap.Goals, w.looking.Reads())
	return found
}

// Looking is what has been typed into the search so far.
func (w Window) Looking() string { return w.looking.Reads() }

// Searched is the search as a written text, caret and all.
func (w Window) Searched() Written { return w.looking }

// other is the page somebody is not on. Both tab and shift+tab come here.
//
// Written as "the other one" and not as a walk by a given number of pages.
// There are two, so which way round a walk goes is a difference nothing can
// see — and written as a walk it was arithmetic no test could ever judge, which
// the mutation gate said out loud. A third page would make it a walk again, and
// would need this rewritten rather than extended.
//
// Changing page lets go of what was chosen: a card chosen on the page you left
// is a card the arrows would walk from on the page you are on. It lets go of
// what went wrong too: said on one page, it is about that page, and read on
// the other it would be about nothing there.
func (w Window) other() Window {
	if w.tab == Tasks {
		w.tab = Objectives
	} else {
		w.tab = Tasks
	}
	w.chosen, w.chosenGoal = 0, false
	w.err, w.wrongFor = nil, 0
	return w
}

// working takes up the chosen task, or lets it go when it is already being
// worked on. Both name the task: several can be in focus at once, so letting
// go of nothing in particular would let go of every one of them.
func (w Window) working() Window {
	card, found := w.snap.Card(w.chosen)
	if !found {
		return w
	}
	return w.asking(reading.FocusOp(card), reading.Numbered(w.chosen))
}

// chosenIsArchived says whether the card chosen is one off the board.
func (w Window) chosenIsArchived() bool {
	card, _ := w.snap.Card(w.chosen)
	return card.Archived
}

// puttingAway takes the chosen card off the board, or brings it back.
func (w Window) puttingAway() Window {
	card, found := w.snap.Card(w.chosen)
	if !found {
		return w
	}
	return w.asking(reading.AwayOp(card), reading.Numbered(w.chosen))
}

// LaneSays is what a lane of the band writes above itself: what it is, and how
// many it holds. The four columns have counted since the beginning; the three
// lanes did not, so the one run of cards that always wants acting on was the
// one nobody could count without counting.
func (w Window) LaneSays(kind string) string {
	lane := slices.Index(Lanes, kind)
	if lane < 0 {
		return kind + "  0"
	}
	return kind + "  " + strconv.Itoa(len(w.inLane(lane)))
}

// Holds is how much a page holds, which is what its tab says beside its name.
// Read off what is showing, so a search that narrows the board narrows what the
// tabs claim as well.
func (w Window) Holds(tab Tab) int {
	if tab == Objectives {
		return len(w.Showing().Goals)
	}
	return len(w.Showing().Cards)
}
