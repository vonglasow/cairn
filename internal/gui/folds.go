package gui

import (
	"cairn/internal/board"
	"cairn/internal/gui/geom"
	"slices"
	"strconv"
	"strings"

	"cairn/internal/engine"
)

func (w Window) unfolding(n int) Window {
	if !slices.Contains(w.Unfolded, n) {
		w.Unfolded = append(slices.Clone(w.Unfolded), n)
	}
	return w
}

func (w Window) foldingAway(n int) Window {
	if c, ok := w.Snap.Goal(w.Chosen); w.ChosenGoal && ok && c.Parent == n {
		w.Chosen = n
	}
	w.Unfolded = slices.DeleteFunc(slices.Clone(w.Unfolded), func(o int) bool { return o == n })
	return w
}

func (w Window) togglingFold(n int) Window {
	if w.Folded(n) {
		return w.unfolding(n).settlingGoals()
	}
	return w.foldingAway(n).settlingGoals()
}

func (w Window) settlingGoals() Window {
	w.goals = w.settledGoals(w.goals)
	return w
}

func (w Window) openingParentOf(n int) Window {
	if g, ok := w.Snap.Goal(n); ok && g.Parent != 0 {
		return w.unfolding(g.Parent)
	}
	return w
}

func (w Window) choosingGoal(n int) Window {
	w.Chosen, w.ChosenGoal, w.DetailGoal = n, true, n
	w = w.openingParentOf(n)
	for at, b := range w.Tree() {
		if b.Goal.Number == n {
			return w.followingGoal(at)
		}
	}
	return w
}

func (w Window) foldingKey(name string) (Window, bool) {
	g, ok := w.Detailed()
	if !ok {
		return w, slices.Contains([]string{"left", "right", "z", "alt+left", "alt+right"}, name)
	}
	switch name {
	case "right":
		w = w.choosingGoal(g.Number)
		if w.Folded(g.Number) {
			return w.unfolding(g.Number), true
		}
		if len(g.Children) > 0 {
			for _, b := range w.Tree() {
				if b.Under && b.Goal.Parent == g.Number {
					return w.choosingGoal(b.Goal.Number), true
				}
			}
		}
		return w, true
	case "left":
		if g.Parent != 0 {
			return w.choosingGoal(g.Parent), true
		}
		w = w.choosingGoal(g.Number)
		if len(g.Children) > 0 {
			return w.foldingAway(g.Number).settlingGoals(), true
		}
		return w, true
	case "z":
		if g.Parent != 0 {
			w = w.choosingGoal(g.Number)
			return w.foldingAway(g.Parent).choosingGoal(g.Parent).settlingGoals(), true
		}
		if len(g.Children) > 0 {
			return w.choosingGoal(g.Number).togglingFold(g.Number), true
		}
		return w, true
	case "shift+up", "shift+down":
		in, moves := board.Stepped(w.Showing().Goals, g.Number, name == "shift+up")
		if !moves {
			return w, true
		}
		return w.asking(engine.ObjectiveMove, in).choosingGoal(g.Number), true
	case "alt+right":
		for _, b := range board.Branches(w.Showing().Goals) {
			if len(b.Goal.Children) > 0 {
				w = w.unfolding(b.Goal.Number)
			}
		}
		return w.settlingGoals(), true
	case "alt+left":
		if c, chosen := w.Snap.Goal(w.Chosen); w.ChosenGoal && chosen && c.Parent != 0 {
			w.Chosen = c.Parent
		}
		w.Unfolded = nil
		w = w.settlingGoals()
		if w.ChosenGoal {
			return w.choosingGoal(w.Chosen), true
		}
		return w, true
	}
	return w, false
}

func (w Window) narrowList() bool { return GoalAt(w.Layout(), 0, 0).W < narrowItem }

const narrowItem = 300

// SignalOf is what an objective's second line ends with in the late ink: its own deadline past or what it waits on, or, folded, the first of what it hides that wants seeing, the rest counted — and whether a dot stands before it, saying it is hidden.
func (w Window) SignalOf(b board.Branch) (string, bool) {
	g := b.Goal
	if b.Holds {
		return "", false
	}
	narrow := w.narrowList()
	if g.LateOn(w.Today()) {
		return g.Deadline.Format("2 Jan") + ", late", false
	}
	if on, waits := w.Snap.Unreached(g); waits {
		if narrow {
			return "waits #" + strconv.Itoa(on.Number), false
		}
		return "waits on " + strings.Join(w.UnreachedTitles(g), ", "), false
	}
	if b.Under || !w.Folded(g.Number) {
		return "", false
	}
	hid := w.Snap.Hidden(g.Number, w.Today())
	if len(hid) == 0 {
		return "", false
	}
	first := hid[0]
	said := "#" + strconv.Itoa(first.Goal.Number)
	switch {
	case first.Late && narrow:
		said += " late"
	case first.Late:
		said += " " + first.Goal.Title + " late"
	case narrow:
		said += " waits"
	default:
		said += " waits on #" + strconv.Itoa(first.On.Number)
	}
	if more := len(hid) - 1; more > 0 {
		said += " +" + strconv.Itoa(more)
	}
	return said, true
}

func (w Window) hiddenAt(p geom.Point) ([]string, geom.Box, int) {
	if w.Screen != board.OnTheBoard || w.Tab != board.Objectives || !GoalsRun(w.Layout()).In(p) {
		return nil, geom.Box{}, 0
	}
	for at, b := range w.Tree() {
		said, dot := w.SignalOf(b)
		if !dot {
			continue
		}
		slot := GoalAt(w.Layout(), at, w.goals)
		_, _, second := ItemLine(ItemOf(slot, 0))
		mark, words := SignalAt(second, said, dot, w.measured())
		reach := geom.Box{X: mark.X, Y: second.Y, W: words.X + words.W - mark.X, H: second.H}
		if reach.In(p) {
			return w.HiddenSays(b.Goal.Number), reach, slot.W
		}
	}
	return nil, geom.Box{}, 0
}
