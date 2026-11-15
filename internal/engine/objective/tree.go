package objective

import (
	"maps"
	"slices"
)

// Tree is every objective by its number, and the sub-objectives of each.
type Tree struct {
	By       map[int]Objective
	Children map[int][]int
}

// TreeOf makes the tree of the objectives given.
func TreeOf(all []Objective) Tree {
	t := Tree{By: map[int]Objective{}, Children: map[int][]int{}}
	for _, o := range all {
		c := o.Carried()
		t.By[c.Number] = o
		if c.Parent != 0 {
			t.Children[c.Parent] = append(t.Children[c.Parent], c.Number)
		}
	}
	return t
}

// Reach is an objective and its sub-objectives.
func (t Tree) Reach(n int) []int { return append([]int{n}, t.Children[n]...) }

// Counted is the objectives served, with the parent of each.
func (t Tree) Counted(served []int) []int {
	out := slices.Clone(served)
	for _, n := range served {
		if parent := t.By[n].Carried().Parent; parent != 0 && !slices.Contains(out, parent) {
			out = append(out, parent)
		}
	}
	return out
}

// ChildrenOf is the sub-objectives of an objective.
func (t Tree) ChildrenOf(n int) []Objective {
	out := make([]Objective, 0, len(t.Children[n]))
	for _, c := range t.Children[n] {
		out = append(out, t.By[c])
	}
	return out
}

// Waiters is, for each objective, the objectives waiting on it.
func (t Tree) Waiters() map[int][]int {
	out := map[int][]int{}
	for _, n := range slices.Sorted(maps.Keys(t.By)) {
		for _, on := range t.By[n].Carried().WaitsOn {
			out[on] = append(out[on], n)
		}
	}
	return out
}
