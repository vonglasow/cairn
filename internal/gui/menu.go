package gui

import (
	"strconv"
	"strings"

	"cairn/internal/engine"
)

type menuState struct {
	menu     []string
	menuAt   Point
	menuOn   int
	menuSets string
}

const (
	menuWide = 180
	menuRow  = 28
)

// ContextAt is a right click at a point of the board: it chooses what is under the pointer, or nothing, and opens what can be done to it.
func (w Window) ContextAt(p Point) Window {
	if w.Blank() {
		return w
	}
	if w.screen != OnTheBoard {
		return w
	}
	w = w.aiming(p).choosingAt(w.at(p))
	w.menu, w.menuAt, w.menuOn, w.menuSets = w.menuFor(), p, 0, ""
	w.screen = OnMenu
	return w
}

func (w Window) menuFor() []string {
	if w.chosen == 0 {
		return []string{"New", "Search", "Archived", "Help", "New backup", "Backups", "Labels", "Export", "Import"}
	}
	if w.chosenGoal {
		return []string{"Edit", "Delete"}
	}
	card, _ := w.snap.Card(w.chosen)
	return []string{"Open", "Edit", "New linked task", engine.FocusOp(card).Offered(), "Move to…", "Priority…", "Labels…", engine.AwayOp(card).Offered(), "Delete"}
}

func (w Window) labellingChosen() Window {
	w = w.formOnto("label").choosingList()
	w.pickingLabels = w.screen == OnChoosing
	return w
}

func (w Window) datingChosen() Window {
	return w.formOnto("deadline").openingCalendar()
}

func (w Window) formOnto(arg string) Window {
	w.screen = OnTheBoard
	w = w.Edit()
	for at, field := range w.form.Fields() {
		if field.Arg == arg {
			w.form = w.form.onto(at)
		}
	}
	return w
}

// settingUnderTheCard opens the values a field of the chosen task takes in a menu under its card, the keyboard's way to what the right-click menu offers.
func (w Window) settingUnderTheCard(field string, values []string, held string) Window {
	for _, placed := range w.Placed() {
		if placed.Card.Number == w.chosen {
			w.menuAt = Point{X: placed.Slot.X, Y: placed.Slot.Y + placed.Slot.H}
		}
	}
	w.screen = OnMenu
	return w.settingFrom(field, values, held)
}

func (w Window) settingFrom(field string, values []string, held string) Window {
	w.menu, w.menuOn, w.menuSets = nil, 0, field
	for _, value := range values {
		if value != held {
			w.menu = append(w.menu, value)
		}
	}
	return w
}

// MenuKey is the key that does what a line of the menu does, which the line says at its right: the menu is where a key is learned.
func MenuKey(line string) string {
	switch line {
	case "Open":
		return "enter"
	case "Edit":
		return "e"
	case "New linked task":
		return "o"
	case "Focus", "Unfocus":
		return "f"
	case "Move to…":
		return "m"
	case "Priority…":
		return "p"
	case "Labels…":
		return "l"
	case "Archive", "Unarchive":
		return "x"
	}
	return Shortcut(line)
}

// Menu is what the open menu offers, line by line.
func (w Window) Menu() []string { return w.menu }

// MenuOn is the line of the menu the keyboard is on.
func (w Window) MenuOn() int { return w.menuOn }

// MenuBox is where the menu is drawn: from the point clicked, down and to the right, kept inside the window when the point is near an edge.
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

func (w Window) onMenu(name string) Window {
	if w.menuSets == "state" {
		for i, value := range w.menu {
			if len(name) == 1 && strings.HasPrefix(value, name) {
				return w.fromMenu(i)
			}
		}
	}
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
		return w.labellingChosen()
	case "Open":
		w.screen = OnTheBoard
		return w.opening()
	}
	w.screen = OnTheBoard
	if said == "Focus" || said == "Unfocus" {
		return w.working()
	}
	return w.Choose(said)
}

func (w Window) chosenCard() engine.Card {
	card, _ := w.snap.Card(w.chosen)
	return card
}
