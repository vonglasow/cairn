package gui

import (
	"cairn/internal/engine"
	"slices"
	"strconv"
	"strings"
)

// Showing is the work narrowed to what the filters keep and a search reaches, which is what everything draws from.
func (w Window) Showing() engine.Snapshot {
	if w.looking.Empty() && !w.filters.Narrows() && w.viewing == "" {
		return w.snap
	}
	cards := w.snap.Viewed(w.viewing, w.snap.Cards)
	if kept, err := w.filters.Kept(cards); err == nil {
		cards = kept
	}
	found := engine.Snapshot{Focus: w.snap.Focus, Version: w.snap.Version}
	found.Cards = engine.Found(cards, w.looking.Reads(), w.archived)
	found.Goals = engine.FoundGoals(w.filters.KeptGoals(w.snap.Goals), w.looking.Reads())
	return found
}

// Searched is the search as a written text, caret and all.
func (w Window) Searched() Written { return w.looking }

var pageOfTab = map[Tab]string{Tasks: engine.PageTasks, Objectives: engine.PageObjectives, Time: engine.PageTime, Series: engine.PageTasks}

// Paged answers the window showing only the pages the configuration leaves on, on the first of them when the one it was on is off.
func (w Window) Paged(p engine.Pages) Window {
	w.pages = p
	if shown := w.Shown(); len(shown) > 0 && !slices.Contains(shown, w.tab) {
		return w.onPage(shown[0])
	}
	return w
}

// Shown is the pages drawn, in order: every one the configuration did not turn off.
func (w Window) Shown() []Tab {
	var shown []Tab
	for tab := range Pages {
		if Tab(tab) == Series && !w.pages.Recurs() {
			continue
		}
		if w.pages.Shows(pageOfTab[Tab(tab)]) {
			shown = append(shown, Tab(tab))
		}
	}
	return shown
}

// Titled is a column's or a page's name as the window writes it: in words, not shouted.
func Titled(name string) string {
	if name == "" {
		return name
	}
	lower := strings.ToLower(name)
	return strings.ToUpper(lower[:1]) + lower[1:]
}

// ColumnHeads is what each column's head says: its name, then how many cards it holds.
func (w Window) ColumnHeads() []string {
	dealt, _ := engine.Deal(w.Showing().Cards, w.archived)
	heads := make([]string, 0, len(dealt))
	for _, column := range dealt {
		heads = append(heads, Titled(column.Name)+" "+strconv.Itoa(len(column.Cards)))
	}
	return heads
}

// EmptySays is what a column holding nothing says in place of its cards: how to make the first task on a board holding none, nothing at all on a column holding some.
func (w Window) EmptySays(column int) string {
	dealt, _ := engine.Deal(w.Showing().Cards, w.archived)
	if column < 0 || column >= len(dealt) || len(dealt[column].Cards) > 0 {
		return ""
	}
	if len(w.Showing().Cards) == 0 && column == engine.IndexOf("todo") {
		return "No task yet. " + w.keyboard.Spelled(Pretty(Shortcut("new"))) + " makes one."
	}
	return nothingHere
}

const nothingHere = "Nothing here"

// Dimmed says a card is drawn at half strength: its work is behind it, done or archived.
func Dimmed(c engine.Card) bool { return c.Archived || c.State == "done" }

// Blank says every page is turned off: the window is drawn, and nothing in it.
func (w Window) Blank() bool { return w.pages.None() }

func (w Window) other(by int) Window {
	shown := w.Shown()
	if len(shown) == 0 {
		return w
	}
	at := max(0, slices.Index(shown, w.tab))
	return w.onPage(shown[(at+by+len(shown))%len(shown)])
}

func (w Window) onPage(tab Tab) Window {
	if !slices.Contains(w.Shown(), tab) {
		return w
	}
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
	if tab == Series {
		return len(w.SeriesListed())
	}
	if tab == Time {
		return w.sheet.Worked()
	}
	return len(engine.OnTheBoard(w.Showing().Cards, w.archived))
}
