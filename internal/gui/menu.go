package gui

// What a right click offers. What a card can have done to it was on the bar,
// at the foot of the window, once the card was chosen: a hand on the mouse
// went to the card, then down to the bar. A right click chooses what is under
// it and offers the same acts beside it — a task's, an objective's, or the
// board's where there is neither. Nothing here is new: every line is an offer
// the bar already makes, done the way the bar does it. See docs/gui.md.

import (
	"strconv"

	"cairn/internal/engine"
)

// menuState is what the menu keeps: its lines, the point it was opened at,
// the line the keyboard is on, and the field of the chosen task its lines
// are values for, when they are.
type menuState struct {
	menu     []string
	menuAt   Point
	menuOn   int
	menuSets string
}

const (
	// menuWide is how wide the menu is, and menuRow how high one line of it.
	menuWide = 180
	menuRow  = 28
)

// ContextAt is a right click at a point of the board: it chooses what is
// under the pointer, or nothing, and opens what can be done to it.
func (w Window) ContextAt(p Point) Window {
	if w.screen != OnTheBoard {
		return w
	}
	w = w.choosingAt(w.at(p))
	w.menu, w.menuAt, w.menuOn, w.menuSets = w.menuFor(), p, 0, ""
	w.screen = OnMenu
	return w
}

// menuFor is what the menu offers for what is chosen: what the bar would
// offer for it, the focus with it, and the board's own offers for nothing.
func (w Window) menuFor() []string {
	if w.chosen == 0 {
		return []string{"New", "Search", "Archived", "Help", "New backup", "Backups", "Export", "Import"}
	}
	if w.chosenGoal {
		return []string{"Edit", "Delete"}
	}
	card, _ := w.snap.Card(w.chosen)
	focus := "Focus"
	if card.InFocus {
		focus = "Unfocus"
	}
	away := "Archive"
	if card.Archived {
		away = "Unarchive"
	}
	return []string{"Edit", focus, "Move to…", "Priority…", "Labels…", away, "Delete"}
}

// settingFrom turns the menu into the values a field of the chosen task may
// take, but the one it holds: a column, or a priority, taken from the menu
// at once.
func (w Window) settingFrom(field string, values []string, held string) Window {
	w.menu, w.menuOn, w.menuSets = nil, 0, field
	for _, value := range values {
		if value != held {
			w.menu = append(w.menu, value)
		}
	}
	return w
}

// Menu is what the open menu offers, line by line.
func (w Window) Menu() []string { return w.menu }

// MenuOn is the line of the menu the keyboard is on.
func (w Window) MenuOn() int { return w.menuOn }

// MenuBox is where the menu is drawn: from the point clicked, down and to the
// right, kept inside the window when the point is near an edge.
func (w Window) MenuBox() Box {
	high := between + len(w.menu)*menuRow + between
	return Box{
		X: max(0, min(w.menuAt.X, w.width-menuWide)),
		Y: max(0, min(w.menuAt.Y, w.height-high)),
		W: menuWide,
		H: high,
	}
}

// MenuRowAt is where the i-th line of the menu is.
func MenuRowAt(box Box, i int) Box {
	return Box{X: box.X + between, Y: box.Y + between + i*menuRow, W: box.W - between - between, H: menuRow}
}

// onMenu is what a key does on the menu: the arrows and j k walk it, enter
// takes a line and esc closes it.
func (w Window) onMenu(name string) Window {
	switch name {
	case "down", "j":
		w.menuOn = min(len(w.menu)-1, w.menuOn+1)
	case "up", "k":
		w.menuOn = max(0, w.menuOn-1)
	case "enter":
		return w.fromMenu(w.menuOn)
	case "esc":
		w.screen = OnTheBoard
	}
	return w
}

// clickingMenu takes the line clicked, and a click off the menu closes it and
// does nothing else: it was a click to put the menu away.
func (w Window) clickingMenu(p Point) Window {
	box := w.MenuBox()
	for i := range w.menu {
		if MenuRowAt(box, i).In(p) {
			return w.fromMenu(i)
		}
	}
	w.screen = OnTheBoard
	return w
}

// fromMenu does what a line of the menu offers, the way the bar does it.
func (w Window) fromMenu(i int) Window {
	said, card := w.menu[i], w.chosenCard()
	if w.menuSets != "" {
		w.screen = OnTheBoard
		return w.asking(engine.TaskEdit, engine.Input{Given: map[string]string{
			"number": strconv.Itoa(w.chosen), w.menuSets: said}})
	}
	switch said {
	case "Move to…":
		return w.settingFrom("state", engine.States(), card.State)
	case "Priority…":
		return w.settingFrom("priority", engine.Priorities(), card.Priority)
	case "Labels…":
		w.screen = OnTheBoard
		w = w.Edit()
		for at, field := range w.form.Fields() {
			if field.Arg == "label" {
				w.form = w.form.onto(at)
			}
		}
		return w.choosingList()
	}
	w.screen = OnTheBoard
	if said == "Focus" || said == "Unfocus" {
		return w.working()
	}
	return w.Choose(said)
}

// chosenCard is the task chosen, as the board last read it.
func (w Window) chosenCard() engine.Card {
	card, _ := w.snap.Card(w.chosen)
	return card
}
