package gui

import (
	"cairn/internal/board"
	"cairn/internal/gui/geom"
	"slices"
	"strconv"
	"strings"

	"cairn/internal/engine"
)

type menuState struct {
	menu     []string
	menuAt   geom.Point
	menuOn   int
	menuSets string
	rules    []int
}

const (
	menuWide = 180
	menuRow  = 28
)

// ContextAt is a right click at a point of the board: it chooses what is under the pointer, or nothing, and opens what can be done to it.
func (w Window) ContextAt(p geom.Point) Window {
	if w.Blank() {
		return w
	}
	if w.Screen != board.OnTheBoard {
		return w
	}
	if w.Tab == board.Agents {
		return w.sessionMenuAt(p)
	}
	w = w.aiming(p)
	if n := w.at(p); w.IsTaken(n) && w.Tab == board.Tasks {
		w.Chosen = n
	} else {
		w = w.choosingAt(n)
	}
	w.menu, w.menuAt, w.menuOn, w.menuSets, w.rules = w.menuFor(), p, 0, "", nil
	if len(w.Taken()) > 0 && !w.ChosenGoal {
		w.rules = []int{slices.Index(w.menu, "Archive")}
	}
	w.Screen = board.OnMenu
	return w
}

func (w Window) menuFor() []string {
	if w.Chosen == 0 {
		return []string{"New", "Search", "Archived", "Help", "New backup", "Backups", "Labels", "Export", "Import"}
	}
	if taken := w.Taken(); len(taken) > 0 && !w.ChosenGoal {
		focus := "Unfocus"
		for _, n := range taken {
			if !slices.Contains(w.Snap.Focus, n) {
				focus = "Focus"
			}
		}
		return []string{"Move to…", "Priority…", "Labels…", focus, "Archive", "Delete"}
	}
	thing := "task"
	if w.ChosenGoal {
		thing = "objective"
	}
	var lines []string
	for _, op := range board.MenuOf(thing, w.Pages) {
		lines = append(lines, w.linesOf(op)...)
	}
	if !w.ChosenGoal && w.Pages.Offers(engine.AgentStart) {
		at := len(lines)
		if deleting := slices.Index(lines, "Delete"); deleting >= 0 {
			at = deleting
		}
		lines = slices.Insert(lines, at, StartSessionSays)
	}
	return lines
}

func (w Window) linesOf(op engine.Operation) []string {
	card := w.ChosenCard()
	switch {
	case w.ChosenGoal && (op.Verb == "edit" || op.Verb == "delete"):
		return []string{board.Offered(op)}
	case w.ChosenGoal:
		return nil
	case op.Verb == "show":
		return []string{"Open"}
	case op.Verb == "edit":
		return []string{"Edit", "New linked task", "Move to…", "Priority…", "Labels…"}
	case op.Verb == "delete", op.Verb == board.FocusOp(card).Verb, op.Verb == board.AwayOp(card).Verb:
		return []string{board.Offered(op)}
	case op.Verb == "skip" && card.Series != nil:
		return []string{SkipThisOneSays}
	}
	return nil
}

func (w Window) actedOn() engine.Input {
	if taken := w.Taken(); len(taken) > 0 {
		return board.Several(taken)
	}
	return board.Numbered(w.Chosen)
}

func (w Window) labellingChosen() Window {
	w = w.formOnto("label").choosingList()
	w.PickingLabels = w.Screen == board.OnChoosing
	if taken := w.Taken(); len(taken) > 0 && w.PickingLabels {
		return w.labelsOfSeveral(taken)
	}
	return w
}

func (w Window) datingChosen() Window {
	return w.formOnto("deadline").openingCalendar()
}

func (w Window) formOnto(arg string) Window {
	w.Screen = board.OnTheBoard
	w = w.Edit()
	for at, field := range w.Form.Fields {
		if field.Arg == arg {
			w.Form = w.Form.Onto(at)
		}
	}
	return w
}

// settingUnderTheCard opens the values a field of the chosen task takes in a menu under its card, the keyboard's way to what the right-click menu offers.
func (w Window) settingUnderTheCard(field string, values []string, held string) Window {
	for _, placed := range w.Placed() {
		if placed.Card.Number == w.Chosen {
			w.menuAt = geom.Point{X: placed.Slot.X, Y: placed.Slot.Y + placed.Slot.H}
		}
	}
	w.Screen = board.OnMenu
	return w.settingFrom(field, values, held)
}

func (w Window) settingFrom(field string, values []string, held string) Window {
	w.menu, w.menuOn, w.menuSets, w.rules = nil, 0, field, nil
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

// MenuRules are the lines of the menu a rule is drawn above, each starting a group of its own.
func (w Window) MenuRules() []int { return w.rules }

// MenuOn is the line of the menu the keyboard is on.
func (w Window) MenuOn() int { return w.menuOn }

// MenuHead is what the menu of gathered cards says above its lines: how many, and which.
func (w Window) MenuHead() string {
	taken := w.Taken()
	if w.Screen != board.OnMenu || len(taken) == 0 || w.ChosenGoal {
		return ""
	}
	said := strconv.Itoa(len(taken)) + " tasks ·"
	for _, n := range taken {
		said += " #" + strconv.Itoa(n)
	}
	return said
}

func (w Window) menuHeadHigh() int {
	if w.MenuHead() == "" {
		return 0
	}
	return menuRow
}

// MenuRows is the part of the menu its lines are in, under its head.
func (w Window) MenuRows() geom.Box {
	box := w.MenuBox()
	box.Y, box.H = box.Y+w.menuHeadHigh(), box.H-w.menuHeadHigh()
	return box
}

// MenuBox is where the menu is drawn: from the point clicked, down and to the right, kept inside the window when the point is near an edge.
func (w Window) MenuBox() geom.Box {
	high := w.menuHeadHigh() + between + len(w.menu)*menuRow + between
	return geom.Box{
		X: max(0, min(w.menuAt.X, w.width-menuWide)),
		Y: max(0, min(w.menuAt.Y, w.height-high)),
		W: menuWide,
		H: high,
	}
}

// MenuRowAt is where the i-th line of the menu is.
func MenuRowAt(box geom.Box, i int) geom.Box {
	return geom.Box{X: box.X + between, Y: box.Y + between + i*menuRow, W: box.W - between - between, H: menuRow}
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
		w.Screen = board.OnTheBoard
	}
	return w
}

func (w Window) clickingMenu(p geom.Point) Window {
	box := w.MenuRows()
	for i := range w.menu {
		if MenuRowAt(box, i).In(p) {
			return w.fromMenu(i)
		}
	}
	w.Screen = board.OnTheBoard
	return w
}

func (w Window) fromMenu(i int) Window {
	said := w.menu[i]
	if w.menuSets != "" {
		w.Screen = board.OnTheBoard
		in := w.actedOn()
		in.Given[w.menuSets] = said
		return w.asking(engine.TaskEdit, in)
	}
	return w.doingLine(said)
}

func (w Window) doingLine(said string) Window {
	card := w.ChosenCard()
	switch said {
	case "Move to…":
		return w.settingFrom("state", engine.States(), card.State)
	case "Priority…":
		return w.settingFrom("priority", board.Priorities(), card.Priority)
	case "Labels…":
		return w.labellingChosen()
	case "Open":
		w.Screen = board.OnTheBoard
		return w.opening()
	case SkipThisOneSays:
		w.Screen = board.OnTheBoard
		return w.asking(engine.TaskSkip, board.Numbered(w.Chosen))
	case OpenTaskSays, GoToSessionSays, CopyResumeSays, NewTaskForSays, TieSays, AllowSays, DenySays, AllowAlwaysSays:
		w.Screen = board.OnTheBoard
		return w.onSession(said)
	case StartSessionSays:
		w.Screen = board.OnTheBoard
		return w.startingSession()
	}
	w.Screen = board.OnTheBoard
	if len(w.Taken()) > 0 && (said == "Focus" || said == "Unfocus") {
		op := engine.TaskFocus
		if said == "Unfocus" {
			op = engine.TaskUnfocus
		}
		return w.asking(op, board.Several(w.Taken()))
	}
	if said == "Focus" || said == "Unfocus" {
		return w.working()
	}
	return w.Choose(said)
}
