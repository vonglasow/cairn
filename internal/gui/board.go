package gui

import (
	"cairn/internal/board"
	"cairn/internal/engine"
	"cairn/internal/gui/geom"
	"slices"
	"strconv"
	"strings"
)

var pageOfTab = map[board.Tab]string{board.Tasks: engine.PageTasks, board.Objectives: engine.PageObjectives, board.Time: engine.PageTime, board.Series: engine.PageTasks, board.Agents: engine.PageAgents}

// Paged answers the window showing only the pages the configuration leaves on, on the first of them when the one it was on is off.
func (w Window) Paged(p engine.Pages) Window {
	w.Pages = p
	if shown := w.Shown(); len(shown) > 0 && !slices.Contains(shown, w.Tab) {
		return w.onPage(shown[0])
	}
	return w
}

// Shown is the pages drawn, in order: every one the configuration did not turn off.
func (w Window) Shown() []board.Tab {
	var shown []board.Tab
	for tab := range Pages {
		if board.Tab(tab) == board.Series && !w.Pages.Recurs() {
			continue
		}
		if w.Pages.Shows(pageOfTab[board.Tab(tab)]) {
			shown = append(shown, board.Tab(tab))
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
	dealt, _ := board.Deal(w.Showing().Cards, w.Archived)
	heads := make([]string, 0, len(dealt))
	for _, column := range dealt {
		heads = append(heads, Titled(column.Name)+" "+strconv.Itoa(len(column.Cards)))
	}
	return heads
}

// EmptySays is what a column holding nothing says in place of its cards: how to make the first task on a board holding none, nothing at all on a column holding some.
func (w Window) EmptySays(column int) string {
	dealt, _ := board.Deal(w.Showing().Cards, w.Archived)
	if column < 0 || column >= len(dealt) || len(dealt[column].Cards) > 0 {
		return ""
	}
	if len(w.Showing().Cards) == 0 && column == board.IndexOf("todo") {
		return "No task yet. `" + w.keyboard.Spelled(Pretty(Shortcut("new"))) + "` makes one."
	}
	return nothingHere
}

const nothingHere = "Nothing here"

// Dimmed says a card is drawn at half strength: its work is behind it, done or archived.
func Dimmed(c board.Card) bool { return c.Archived || c.State == "done" }

// BlankSays is what a window with every page turned off says, over the offer to open the settings.
const BlankSays = "Every tab is turned off"

// BlankSays is what the window says when it has no page to show, empty otherwise.
func (w Window) BlankSays() string {
	if w.Blank() && w.Screen != board.OnConfiguration {
		return BlankSays
	}
	return ""
}

// BlankOffer is where a window with no page offers to open the settings: under what it says, in its middle.
func BlankOffer(width, height int) geom.Box {
	return geom.Box{X: width/2 - 90, Y: height/2 + lineHigh, W: 180, H: 30}
}

func (w Window) other(by int) Window {
	shown := w.Shown()
	if len(shown) == 0 {
		return w
	}
	at := max(0, slices.Index(shown, w.Tab))
	return w.onPage(shown[(at+by+len(shown))%len(shown)])
}

func (w Window) onPage(tab board.Tab) Window {
	if !slices.Contains(w.Shown(), tab) {
		return w
	}
	w.Tab = tab
	w.Chosen, w.ChosenGoal, w.Gathered = 0, false, nil
	w.Err, w.wrongFor = nil, 0
	if w.Tab == board.Time {
		return w.readingWeek()
	}
	return w
}

func (w Window) working() Window { return w.toggling(board.FocusOp) }

func (w Window) puttingAway() Window { return w.toggling(board.AwayOp) }

func (w Window) toggling(op func(board.Card) engine.Operation) Window {
	card, found := w.Snap.Card(w.Chosen)
	if !found {
		return w
	}
	return w.asking(op(card), board.Numbered(w.Chosen))
}
