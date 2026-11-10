package engine

import (
	"slices"
	"strconv"

	"cairn/internal/domain"
	"cairn/internal/objective"
)

// Branch is one objective as a tree draws it: the objective, whether it sits under the one before it, and whether it is the last under its parent.
type Branch struct {
	Goal  Goal
	Under bool
	Last  bool
}

// Branches lays the objectives out as a tree, in the order a list keeps them: each of level one followed by its sub-objectives, and a sub-objective whose parent is left out standing at level one.
func Branches(goals []Goal) []Branch {
	sorted := Sorted(goals)
	listed := map[int]bool{}
	for _, g := range sorted {
		listed[g.Number] = true
	}
	var out []Branch
	for _, g := range sorted {
		if g.Parent != 0 && listed[g.Parent] {
			continue
		}
		out = append(out, Branch{Goal: g})
		var below []Goal
		for _, c := range sorted {
			if c.Parent == g.Number {
				below = append(below, c)
			}
		}
		for i, c := range below {
			out = append(out, Branch{Goal: c, Under: true, Last: i == len(below)-1})
		}
	}
	return out
}

// Reach is an objective and its sub-objectives: what is counted towards it.
func (s Snapshot) Reach(n int) []int {
	g, _ := s.Goal(n)
	return append([]int{n}, g.Children...)
}

// Segment is one way of narrowing an objective's board: all of it, one sub-objective, or the objective itself; the objective filter's value standing for it, what it reaches, and how many of the cards given it keeps.
type Segment struct {
	Name  string
	Value string
	Reach []int
	Count int
}

// TreeOf is the level-one objective the one numbered heads or sits under: its parent, or itself.
func (s Snapshot) TreeOf(n int) int {
	if g, found := s.Goal(n); found && g.Parent != 0 {
		return g.Parent
	}
	return n
}

// Segments are the ways an objective holding sub-objectives narrows its board, counted over the cards given; nothing for an objective holding none.
func (s Snapshot) Segments(n int, cards []Card) []Segment {
	g, found := s.Goal(n)
	if !found || len(g.Children) == 0 {
		return nil
	}
	value := strconv.Itoa(n)
	out := []Segment{{Name: "all", Value: value, Reach: s.Reach(n)}}
	for _, c := range g.Children {
		sub, _ := s.Goal(c)
		out = append(out, Segment{Name: sub.Title, Value: strconv.Itoa(c), Reach: []int{c}})
	}
	out = append(out, Segment{Name: "itself", Value: value + " itself", Reach: []int{n}})
	for i := range out {
		out[i].Count = len(Serving(cards, out[i].Reach...))
	}
	return out
}

// LabelsFor is every label that may be put on what is named, in alphabetical order.
func (s Snapshot) LabelsFor(on domain.Bearer) []string {
	var out []string
	for label, scope := range s.Scopes {
		if parsed, err := domain.ParseScope(scope); err == nil && parsed.Admits(on) {
			out = append(out, label)
		}
	}
	slices.Sort(out)
	return out
}

// Pick is one objective offered to a task, as a tree: whether it is ticked, and how many ticked below it count it already.
type Pick struct {
	Branch
	Ticked  bool
	Through int
}

// Picks lays every objective out as a tree for a task serving the ones ticked.
func (s Snapshot) Picks(ticked []int) []Pick {
	var out []Pick
	for _, b := range Branches(s.Goals) {
		p := Pick{Branch: b, Ticked: slices.Contains(ticked, b.Goal.Number)}
		for _, c := range b.Goal.Children {
			if slices.Contains(ticked, c) {
				p.Through++
			}
		}
		out = append(out, p)
	}
	return out
}

// CountedOnce refuses the objectives ticked for a task when one of them sits under another, as saving them would be refused.
func (s Snapshot) CountedOnce(task int, ticked []int) error {
	served := make([]objective.Objective, 0, len(ticked))
	for _, n := range ticked {
		if g, found := s.Goal(n); found {
			served = append(served, objective.Made(objective.Carried{Number: g.Number, Title: g.Title, Parent: g.Parent}))
		}
	}
	return objective.CountedOnce(task, served)
}

// Left is what is left of an objective over its subtree: its tasks and its sub-objectives', each once.
func (s Snapshot) Left(n int) objective.Left {
	var l objective.Left
	for _, c := range Serving(s.Cards, s.Reach(n)...) {
		l = l.Counting(c.Number, domain.State(c.State), c.Archived, c.Deadline)
	}
	return l
}

// Parents are the objectives one may sit under: those of level one, itself left out.
func (s Snapshot) Parents(n int) []Goal {
	var out []Goal
	for _, b := range Branches(s.Goals) {
		if !b.Under && b.Goal.Parent == 0 && b.Goal.Number != n {
			out = append(out, b.Goal)
		}
	}
	return out
}

// Fate is what becomes of a sub-objective, or of the tasks, when an objective is deleted.
type Fate string

// The fates a deletion chooses between; none is chosen until the person chooses it.
const (
	Undecided  Fate = ""
	KeepIt     Fate = "keep"
	DeleteIt   Fate = "delete"
	DetachThem Fate = "detach"
)

// Deleting is an objective about to be deleted, and what has been chosen for what goes with it.
type Deleting struct {
	Number int
	Subs   []int
	own    []int
	holds  map[int][]int
	fates  map[int]Fate
	tasks  Fate
}

// Deleting is what deleting an objective asks about: each of its sub-objectives, and the tasks it touches.
func (s Snapshot) Deleting(n int) Deleting {
	g, _ := s.Goal(n)
	d := Deleting{Number: n, Subs: g.Children, own: cardNumbers(Serving(s.Cards, n)), holds: map[int][]int{}, fates: map[int]Fate{}}
	for _, c := range g.Children {
		d.holds[c] = cardNumbers(Serving(s.Cards, c))
	}
	return d
}

func cardNumbers(cards []Card) []int {
	out := make([]int, 0, len(cards))
	for _, c := range cards {
		out = append(out, c.Number)
	}
	return out
}

// Touched is the tasks the deletion reaches as chosen so far: those on the objective itself, and those of each sub-objective deleted with it.
func (d Deleting) Touched() []int {
	out := slices.Clone(d.own)
	for _, c := range d.Subs {
		if d.fates[c] != DeleteIt {
			continue
		}
		for _, n := range d.holds[c] {
			if !slices.Contains(out, n) {
				out = append(out, n)
			}
		}
	}
	slices.Sort(out)
	return out
}

// Choose is the deletion with a fate chosen for one sub-objective.
func (d Deleting) Choose(sub int, f Fate) Deleting {
	fates := make(map[int]Fate, len(d.fates)+1)
	for k, v := range d.fates {
		fates[k] = v
	}
	fates[sub] = f
	d.fates = fates
	return d
}

// Tasks is the deletion with a fate chosen for the tasks it touches.
func (d Deleting) Tasks(f Fate) Deleting {
	d.tasks = f
	return d
}

// FateOf is what was chosen for one sub-objective.
func (d Deleting) FateOf(sub int) Fate { return d.fates[sub] }

// TasksFate is what was chosen for the tasks.
func (d Deleting) TasksFate() Fate { return d.tasks }

// AsksTasks says whether tasks are touched: those on the objective itself, and those of a sub-objective deleted with it.
func (d Deleting) AsksTasks() bool { return len(d.Touched()) > 0 }

// Asks says whether anything has to be chosen before the objective goes.
func (d Deleting) Asks() bool { return len(d.Subs) > 0 || d.AsksTasks() }

// Ready says whether everything asked was chosen.
func (d Deleting) Ready() bool {
	for _, c := range d.Subs {
		if d.fates[c] == Undecided {
			return false
		}
	}
	return !d.AsksTasks() || d.tasks != Undecided
}

// Input is what deleting asks the engine, with every choice made.
func (d Deleting) Input() Input {
	in := Agreed(Numbered(d.Number))
	in.Listed = map[string][]string{}
	for _, c := range d.Subs {
		in.Listed[string(d.fates[c])] = append(in.Listed[string(d.fates[c])], strconv.Itoa(c))
	}
	if d.AsksTasks() {
		in.Given["tasks"] = string(d.tasks)
	}
	return in
}
