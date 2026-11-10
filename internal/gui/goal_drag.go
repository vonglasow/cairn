package gui

import "cairn/internal/engine"

// GoalDragging says whether an objective is being carried by the pointer.
func (w Window) GoalDragging() bool { return w.grabbedGoal != 0 && w.held.Moved }

func (w Window) lettingGoOfAGoal(p Point) Window {
	n, moved := w.grabbedGoal, w.held.Moved
	w.grabbedGoal, w.held = 0, Held{}
	if !moved {
		return w.Click(p)
	}
	over, zone := w.goalDropAt(p)
	in, moves := engine.Dropped(w.Showing().Goals, n, over, zone)
	if !moves {
		return w
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
			return b.Goal.Number, engine.DropAfter
		default:
			return b.Goal.Number, engine.DropInto
		}
	}
	return 0, engine.DropAfter
}

// GoalLanding is the row an objective carried would land by, ringed: the row itself when it would go under it, a thin line above or below it otherwise; and whether there is one.
func (w Window) GoalLanding() (Box, bool) {
	if !w.GoalDragging() {
		return Box{}, false
	}
	over, zone := w.goalDropAt(w.held.At)
	if _, moves := engine.Dropped(w.Showing().Goals, w.grabbedGoal, over, zone); !moves || over == 0 {
		return Box{}, false
	}
	l := w.Layout()
	for i, b := range w.Tree() {
		if b.Goal.Number != over {
			continue
		}
		box := ItemOf(GoalAt(l, i, w.goals), depthOf(b))
		switch zone {
		case engine.DropBefore:
			return Box{X: box.X, Y: box.Y - 1, W: box.W, H: 2}, true
		case engine.DropAfter:
			return Box{X: box.X, Y: box.Y + box.H - 1, W: box.W, H: 2}, true
		}
		return box, true
	}
	return Box{}, false
}
