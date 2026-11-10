package engine

import (
	"slices"
	"time"
)

// Hid is a sub-objective a folded parent hides that wants seeing: late, or waiting on the one named.
type Hid struct {
	Goal Goal
	Late bool
	On   Goal
}

// LateOn says whether an objective is past its deadline on the day given, and neither reached nor dropped.
func (g Goal) LateOn(today time.Time) bool {
	if g.Deadline.IsZero() || g.State == "reached" || g.State == "dropped" {
		return false
	}
	return g.Deadline.Before(startOfDay(today))
}

// Unreached is the first objective one waits on that is not reached yet, and whether there is one.
func (s Snapshot) Unreached(g Goal) (Goal, bool) {
	for _, n := range g.WaitsOn {
		if on, ok := s.Goal(n); ok && on.State != "reached" {
			return on, true
		}
	}
	return Goal{}, false
}

// Hidden is what a parent hides once folded that wants seeing: each sub-objective late first, then each waiting on one not reached; nothing of one reached or dropped.
func (s Snapshot) Hidden(n int, today time.Time) []Hid {
	parent, _ := s.Goal(n)
	var late, waiting []Hid
	for _, c := range parent.Children {
		sub, ok := s.Goal(c)
		if !ok || sub.State == "reached" || sub.State == "dropped" {
			continue
		}
		if sub.LateOn(today) {
			late = append(late, Hid{Goal: sub, Late: true})
			continue
		}
		if on, waits := s.Unreached(sub); waits {
			waiting = append(waiting, Hid{Goal: sub, On: on})
		}
	}
	return append(late, waiting...)
}

// FoundBranches is the tree a search draws: what it reaches, each sub-objective under its parent, a parent that does not match itself standing as what holds the match.
func FoundBranches(goals []Goal, query string) []Branch {
	found := FoundGoals(goals, query)
	if query == "" {
		return Branches(found)
	}
	matched := map[int]bool{}
	for _, g := range found {
		matched[g.Number] = true
	}
	holding := map[int]bool{}
	for _, g := range found {
		if g.Parent == 0 || matched[g.Parent] || holding[g.Parent] {
			continue
		}
		if parent, ok := goalAmong(goals, g.Parent); ok {
			holding[parent.Number] = true
			found = append(found, parent)
		}
	}
	tree := Branches(found)
	for i := range tree {
		tree[i].Holds = holding[tree[i].Goal.Number]
	}
	return tree
}

func goalAmong(goals []Goal, n int) (Goal, bool) {
	at := slices.IndexFunc(goals, func(g Goal) bool { return g.Number == n })
	if at < 0 {
		return Goal{}, false
	}
	return goals[at], true
}
