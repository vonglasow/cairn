package gui

import (
	"cairn/internal/board"
	"cairn/internal/engine/act"
	"cairn/internal/gui/geom"
	"cairn/internal/theme"
	"time"
)

// PanelOffers is where the offers of the panel that has the window are drawn, and so where they are clicked.
func (w Window) PanelOffers() []geom.Box { return w.offersAlong(w.panel()) }

// ServingRows is how many tiles that box shows at once.
func (w Window) ServingRows() int {
	return max(1, w.workFrame().H/geom.SubHeight)
}

// Contrast is how far two colours stand apart, as a scheme is held to.
var Contrast = theme.Contrast

func valueOf(f board.Form, name string) string {
	for _, field := range f.Fields {
		if field.Name == name {
			return field.Value
		}
	}
	return ""
}

// OnSystem is the window written for another system than the one it runs on: what a test asks for, so that what it reads does not depend on the machine.
func (w Window) OnSystem(system string) Window {
	w.keyboard = KeyboardOf(system)
	return w
}

// WhereChosen is where what is chosen is drawn and the run of the window it is drawn in, so that a test can ask whether one is inside the other.
func (w Window) WhereChosen() (box, run geom.Box, ok bool) {
	l := w.Layout()
	if w.Tab == board.Objectives {
		for i, b := range w.Tree() {
			if b.Goal.Number == w.Chosen {
				return GoalAt(l, i, w.goals), GoalsRun(l), true
			}
		}
		return geom.Box{}, geom.Box{}, false
	}
	dealt, _ := board.Deal(w.Showing().Cards, w.Archived)
	for i, held := range dealt {
		for j, c := range held.Cards {
			if c.Number == w.Chosen {
				return l.CardAt(i, j, w.scroll[i]), l.Column[i], true
			}
		}
	}
	for lane := range board.Stalled {
		for j, c := range w.InLane(lane) {
			if c.Number == w.Chosen {
				return l.LaneCardAt(lane, j, w.lanes[lane]), l.Lane[lane], true
			}
		}
	}
	return geom.Box{}, geom.Box{}, false
}

// PartOn is the title of the part the keyboard is on, and whether it is on one.
func (w Window) PartOn() (string, bool) {
	parts := w.Form.Parts()
	if w.Screen != board.OnForm || !w.Form.OnParts() || w.Form.PartAt >= len(parts) {
		return "", false
	}
	return parts[w.Form.PartAt].Title, true
}

// Hovered is the offer the pointer is resting on.
func (w Window) Hovered() (string, bool) { return w.over, w.hovered }

// Serving is the work under the objective that is open, by title and by where each task stands.
func (w Window) Serving() []board.Card {
	if !w.ChosenGoal {
		return nil
	}
	return board.Serving(w.Snap.Cards, w.Chosen)
}

// On puts the keyboard on a field by name, which is what a click does and what a test says.
func (w Window) On(name string) Window {
	for i, field := range w.Form.Fields {
		if field.Name == name {
			w.Form = w.Form.Onto(i)
		}
	}
	return w
}

func anAct(did, field, became, through string) act.Act {
	var told act.AsTold
	told.On(1)
	told.Did(did)
	told.At(time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC))
	if field != "" {
		told.Changing(field)
		told.Became(became)
	}
	if through != "" {
		told.Through(through)
	}
	return told.Act()
}
