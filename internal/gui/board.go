package gui

import (
	"cairn/internal/engine"
	"slices"
	"strconv"
)

// Showing is the work narrowed to what a search reaches, which is what everything draws from.
func (w Window) Showing() engine.Snapshot {
	if w.looking.Empty() {
		return w.snap
	}
	found := engine.Snapshot{Focus: w.snap.Focus, Version: w.snap.Version}
	found.Cards = engine.Found(w.snap.Cards, w.looking.Reads(), w.archived)
	found.Goals = engine.FoundGoals(w.snap.Goals, w.looking.Reads())
	return found
}

// Searched is the search as a written text, caret and all.
func (w Window) Searched() Written { return w.looking }

func (w Window) other(by int) Window {
	return w.onPage(Tab((int(w.tab) + by + len(Pages)) % len(Pages)))
}

func (w Window) onPage(tab Tab) Window {
	w.tab = tab
	w.chosen, w.chosenGoal, w.taken = 0, false, nil
	w.err, w.wrongFor = nil, 0
	if w.tab == Time {
		return w.readingWeek()
	}
	return w
}

func (w Window) working() Window { return w.toggling(engine.FocusOp) }

func (w Window) puttingAway() Window { return w.toggling(engine.AwayOp) }

func (w Window) toggling(op func(engine.Card) engine.Operation) Window {
	card, found := w.snap.Card(w.chosen)
	if !found {
		return w
	}
	return w.asking(op(card), engine.Numbered(w.chosen))
}

// LaneSays is what a lane of the band writes above itself: what it is, and how many it holds.
func (w Window) LaneSays(kind string) string {
	lane := slices.Index(engine.Stalled, kind)
	if lane < 0 {
		return kind + "  0"
	}
	return kind + "  " + strconv.Itoa(len(w.inLane(lane)))
}

// Holds is how much a page holds, which is what its tab says beside its name.
func (w Window) Holds(tab Tab) int {
	if tab == Objectives {
		return len(w.Showing().Goals)
	}
	if tab == Time {
		worked := map[int]bool{}
		for _, d := range w.sheet.Days {
			for _, one := range d.Tasks {
				worked[one.Number] = true
			}
		}
		return len(worked)
	}
	return len(engine.OnTheBoard(w.Showing().Cards, w.archived))
}
