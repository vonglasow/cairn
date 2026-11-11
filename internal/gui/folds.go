package gui

import (
	"slices"
	"strconv"
	"strings"

	"cairn/internal/domain"
	"cairn/internal/engine"
)

// Tree is the objectives drawn, in the order the tab lists them: each of level one followed by its sub-objectives when it is unfolded; a search draws what it reaches, under the parent holding it.
func (w Window) Tree() []engine.Branch {
	if !w.looking.Empty() {
		return engine.FoundBranches(w.filters.KeptGoals(w.snap.Goals), w.looking.Reads())
	}
	var out []engine.Branch
	for _, b := range engine.Branches(w.Showing().Goals) {
		if b.Under && w.Folded(b.Goal.Parent) {
			continue
		}
		out = append(out, b)
	}
	return out
}

// Folded says whether a parent's sub-objectives are put away: every parent starts so, until it is unfolded in this window or holds what is chosen; nothing is folded while a search is typed.
func (w Window) Folded(n int) bool {
	g, ok := w.snap.Goal(n)
	if !ok || len(g.Children) == 0 || !w.looking.Empty() || slices.Contains(w.opened, n) {
		return false
	}
	if c, chosen := w.snap.Goal(w.chosen); w.chosenGoal && chosen && c.Parent == n {
		return false
	}
	return true
}

func (w Window) unfolding(n int) Window {
	if !slices.Contains(w.opened, n) {
		w.opened = append(slices.Clone(w.opened), n)
	}
	return w
}

func (w Window) foldingAway(n int) Window {
	if c, ok := w.snap.Goal(w.chosen); w.chosenGoal && ok && c.Parent == n {
		w.chosen = n
	}
	w.opened = slices.DeleteFunc(slices.Clone(w.opened), func(o int) bool { return o == n })
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
	if g, ok := w.snap.Goal(n); ok && g.Parent != 0 {
		return w.unfolding(g.Parent)
	}
	return w
}

func (w Window) choosingGoal(n int) Window {
	w.chosen, w.chosenGoal, w.detailGoal = n, true, n
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
		in, moves := engine.Stepped(w.Showing().Goals, g.Number, name == "shift+up")
		if !moves {
			return w, true
		}
		return w.asking(engine.ObjectiveMove, in).choosingGoal(g.Number), true
	case "alt+right":
		for _, b := range engine.Branches(w.Showing().Goals) {
			if len(b.Goal.Children) > 0 {
				w = w.unfolding(b.Goal.Number)
			}
		}
		return w.settlingGoals(), true
	case "alt+left":
		if c, chosen := w.snap.Goal(w.chosen); w.chosenGoal && chosen && c.Parent != 0 {
			w.chosen = c.Parent
		}
		w.opened = nil
		w = w.settlingGoals()
		if w.chosenGoal {
			return w.choosingGoal(w.chosen), true
		}
		return w, true
	}
	return w, false
}

// GoalsCount is what the head of the list counts: the objectives shown, or, while a search is typed, those it reaches of all there are.
func (w Window) GoalsCount() string {
	if w.looking.Empty() {
		return strconv.Itoa(len(w.Showing().Goals))
	}
	reached := 0
	for _, b := range w.Tree() {
		if !b.Holds {
			reached++
		}
	}
	return strconv.Itoa(reached) + " of " + strconv.Itoa(len(w.snap.Goals))
}

func (w Window) narrowList() bool { return GoalAt(w.Layout(), 0, 0).W < narrowItem }

const narrowItem = 300

// SignalOf is what an objective's second line ends with in the late ink: its own deadline past or what it waits on, or, folded, the first of what it hides that wants seeing, the rest counted — and whether a dot stands before it, saying it is hidden.
func (w Window) SignalOf(b engine.Branch) (string, bool) {
	g := b.Goal
	if b.Holds {
		return "", false
	}
	narrow := w.narrowList()
	if g.LateOn(w.today()) {
		return g.Deadline.Format("2 Jan") + ", late", false
	}
	if on, waits := w.snap.Unreached(g); waits {
		if narrow {
			return "waits #" + strconv.Itoa(on.Number), false
		}
		return "waits on " + strings.Join(w.unreachedTitles(g), ", "), false
	}
	if b.Under || !w.Folded(g.Number) {
		return "", false
	}
	hid := w.snap.Hidden(g.Number, w.today())
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

func (w Window) unreachedTitles(g engine.Goal) []string {
	var titles []string
	for _, n := range g.WaitsOn {
		if on, ok := w.snap.Goal(n); ok && on.State != "reached" {
			titles = append(titles, on.Title)
		}
	}
	return titles
}

// HiddenSays is what the bubble over a folded parent's signal says: what it hides, each whole, and how to see it.
func (w Window) HiddenSays(n int) []string {
	g, _ := w.snap.Goal(n)
	lines := []string{"Hidden under " + g.Title + ":"}
	for _, h := range w.snap.Hidden(n, w.today()) {
		named := "#" + strconv.Itoa(h.Goal.Number) + " " + h.Goal.Title
		if h.Late {
			lines = append(lines, named+", its deadline "+domain.ShortDay(h.Goal.Deadline)+" past, not reached.")
			continue
		}
		lines = append(lines, named+" waits on #"+strconv.Itoa(h.On.Number)+" "+h.On.Title+", not reached.")
	}
	return append(lines, "→ or a click on its chevron shows it.")
}

func (w Window) hiddenAt(p Point) ([]string, Box, int) {
	if w.screen != OnTheBoard || w.tab != Objectives || !GoalsRun(w.Layout()).In(p) {
		return nil, Box{}, 0
	}
	for at, b := range w.Tree() {
		said, dot := w.SignalOf(b)
		if !dot {
			continue
		}
		slot := GoalAt(w.Layout(), at, w.goals)
		_, _, second := ItemLine(ItemOf(slot, 0))
		mark, words := SignalAt(second, said, dot, w.measured())
		reach := Box{X: mark.X, Y: second.Y, W: words.X + words.W - mark.X, H: second.H}
		if reach.In(p) {
			return w.HiddenSays(b.Goal.Number), reach, slot.W
		}
	}
	return nil, Box{}, 0
}
