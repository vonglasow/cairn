package gui

import (
	"slices"
	"time"

	"cairn/internal/engine"
)

// HoverOpens is how long an objective carried rests over a folded parent before the parent unfolds under it.
const HoverOpens = 600 * time.Millisecond

// GoalDragging says whether an objective is being carried by the pointer.
func (w Window) GoalDragging() bool { return w.grabbedGoal != 0 && w.held.Moved }

func (w Window) hoveringGoal(p Point) Window {
	over, zone := w.goalDropAt(p)
	g, found := w.snap.Goal(over)
	if !found || zone != engine.DropInto || !w.Folded(over) || len(g.Children) == 0 {
		w.hoverGoal = 0
		return w
	}
	if w.hoverGoal != over {
		w.hoverGoal, w.hoverSince = over, w.now
		return w
	}
	if w.now.Sub(w.hoverSince) < HoverOpens {
		return w
	}
	w.hoverGoal = 0
	w.dragOpened = append(slices.Clone(w.dragOpened), over)
	return w.unfolding(over)
}

func (w Window) lettingGoOfAGoal(p Point) Window {
	n, moved := w.grabbedGoal, w.held.Moved
	opened := w.dragOpened
	w.grabbedGoal, w.held, w.hoverGoal, w.dragOpened = 0, Held{}, 0, nil
	if !moved {
		return w.Click(p)
	}
	over, zone := w.goalDropAt(p)
	landsIn := over
	if g, found := w.snap.Goal(over); found && g.Parent != 0 {
		landsIn = g.Parent
	}
	for _, parent := range opened {
		if parent != landsIn {
			w = w.foldingAway(parent)
		}
	}
	in, moves := engine.Dropped(w.Showing().Goals, n, over, zone)
	if !moves {
		return w.settlingGoals()
	}
	return w.asking(engine.ObjectiveMove, in).choosingGoal(n)
}

func (w Window) goalDropAt(p Point) (int, engine.DropZone) {
	l := w.Layout()
	for i, b := range w.Tree() {
		box := ItemOf(GoalAt(l, i, w.goals), depthOf(b))
		if p.Y < box.Y || p.Y >= box.Y+box.H {
			continue
		}
		switch quarter := box.H / 4; {
		case p.Y < box.Y+quarter:
			return b.Goal.Number, engine.DropBefore
		case p.Y >= box.Y+box.H-quarter:
			if b.Under && b.Last && p.X < box.X {
				return b.Goal.Parent, engine.DropAfter
			}
			return b.Goal.Number, engine.DropAfter
		default:
			return b.Goal.Number, engine.DropInto
		}
	}
	return 0, engine.DropAfter
}

func (w Window) goalLanding() (engine.Input, Box, bool) {
	if !w.GoalDragging() {
		return engine.Input{}, Box{}, false
	}
	over, zone := w.goalDropAt(w.held.At)
	in, moves := engine.Dropped(w.Showing().Goals, w.grabbedGoal, over, zone)
	if !moves || over == 0 {
		return engine.Input{}, Box{}, false
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
		if zone == engine.DropBefore {
			return in, Box{X: box.X, Y: box.Y - 1, W: box.W, H: 2}, true
		}
		return in, Box{X: box.X, Y: box.Y + box.H - 1, W: box.W, H: 2}, true
	}
	return engine.Input{}, Box{}, false
}

// GoalLanding is the row an objective carried would land by, ringed: the row itself when it would go under it, a thin line above or below it otherwise; and whether there is one.
func (w Window) GoalLanding() (Box, bool) {
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
