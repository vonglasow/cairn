package gui

import (
	"cairn/internal/board"
	"cairn/internal/gui/geom"
	"slices"
	"time"

	"cairn/internal/engine"
)

// HoverOpens is how long an objective carried rests over a folded parent before the parent unfolds under it.
const HoverOpens = 600 * time.Millisecond

// GoalDragging says whether an objective is being carried by the pointer.
func (w Window) GoalDragging() bool { return w.grabbedGoal != 0 && w.held.Moved }

func (w Window) hoveringGoal(p geom.Point) Window {
	over, zone := w.goalDropAt(p)
	g, found := w.Snap.Goal(over)
	if !found || zone != board.DropInto || !w.Folded(over) || len(g.Children) == 0 {
		w.hoverGoal = 0
		return w
	}
	if w.hoverGoal != over {
		w.hoverGoal, w.hoverSince = over, w.Now
		return w
	}
	if w.Now.Sub(w.hoverSince) < HoverOpens {
		return w
	}
	w.hoverGoal = 0
	w.dragOpened = append(slices.Clone(w.dragOpened), over)
	return w.unfolding(over)
}

func (w Window) lettingGoOfAGoal(p geom.Point) Window {
	n, moved := w.grabbedGoal, w.held.Moved
	opened := w.dragOpened
	w.grabbedGoal, w.held, w.hoverGoal, w.dragOpened = 0, Held{}, 0, nil
	if !moved {
		return w.Click(p)
	}
	over, zone := w.goalDropAt(p)
	landsIn := over
	if g, found := w.Snap.Goal(over); found && g.Parent != 0 {
		landsIn = g.Parent
	}
	for _, parent := range opened {
		if parent != landsIn {
			w = w.foldingAway(parent)
		}
	}
	in, moves := board.Dropped(w.Showing().Goals, n, over, zone)
	if !moves {
		return w.settlingGoals()
	}
	return w.asking(engine.ObjectiveMove, in).choosingGoal(n)
}

func (w Window) goalDropAt(p geom.Point) (int, board.DropZone) {
	l := w.Layout()
	for i, b := range w.Tree() {
		box := ItemOf(GoalAt(l, i, w.goals), depthOf(b))
		if p.Y < box.Y || p.Y >= box.Y+box.H {
			continue
		}
		switch quarter := box.H / 4; {
		case p.Y < box.Y+quarter:
			return b.Goal.Number, board.DropBefore
		case p.Y >= box.Y+box.H-quarter:
			if b.Under && b.Last && p.X < box.X {
				return b.Goal.Parent, board.DropAfter
			}
			return b.Goal.Number, board.DropAfter
		default:
			return b.Goal.Number, board.DropInto
		}
	}
	return 0, board.DropAfter
}

func (w Window) goalLanding() (engine.Input, geom.Box, bool) {
	if !w.GoalDragging() {
		return engine.Input{}, geom.Box{}, false
	}
	over, zone := w.goalDropAt(w.held.At)
	in, moves := board.Dropped(w.Showing().Goals, w.grabbedGoal, over, zone)
	if !moves || over == 0 {
		return engine.Input{}, geom.Box{}, false
	}
	l := w.Layout()
	for i, b := range w.Tree() {
		if b.Goal.Number != over {
			continue
		}
		box := ItemOf(GoalAt(l, i, w.goals), depthOf(b))
		if in.Given["last"] == "true" && in.Given["parent"] != "none" {
			return in, box, true
		}
		if zone == board.DropBefore {
			return in, geom.Box{X: box.X, Y: box.Y - landingHigh/2 - 2, W: box.W, H: landingHigh}, true
		}
		return in, geom.Box{X: box.X, Y: box.Y + box.H + 2 - landingHigh/2, W: box.W, H: landingHigh}, true
	}
	return engine.Input{}, geom.Box{}, false
}

func (w Window) goalRow(n int) (geom.Box, bool) {
	l := w.Layout()
	for i, b := range w.Tree() {
		if b.Goal.Number == n {
			return ItemOf(GoalAt(l, i, w.goals), depthOf(b)), true
		}
	}
	return geom.Box{}, false
}

// GoalGhost is the objective in mid-air: its own row's size, held where it was taken; and whether there is one.
func (w Window) GoalGhost() (geom.Box, bool) {
	if !w.GoalDragging() {
		return geom.Box{}, false
	}
	row, found := w.goalRow(w.grabbedGoal)
	if !found {
		return geom.Box{}, false
	}
	return geom.Box{X: row.X + w.held.At.X - w.held.From.X, Y: row.Y + w.held.At.Y - w.held.From.Y, W: row.W, H: row.H}, true
}

// GoalLanding is the row an objective carried would land by, ringed: the row itself when it would go under it, a thin line above or below it otherwise; and whether there is one.
func (w Window) GoalLanding() (geom.Box, bool) {
	_, box, ok := w.goalLanding()
	return box, ok
}

// GoalLandingSays is the tag beside the landing: where the objective carried would go.
func (w Window) GoalLandingSays() string {
	in, _, ok := w.goalLanding()
	if !ok {
		return ""
	}
	given := in.Given
	switch {
	case given["last"] == "true" && given["parent"] != "none":
		return "under #" + given["parent"]
	case given["first"] == "true" && given["parent"] == "none":
		return "first"
	case given["first"] == "true":
		return "first under #" + given["parent"]
	case given["after"] != "":
		return "after #" + given["after"]
	}
	return ""
}
