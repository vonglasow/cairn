package reading

import (
	"testing"
)

// numbersOf is what a failed sort is reported with: the order it came out in,
// which is the only thing worth reading when the order is wrong.
func numbersOf(goals []Goal) []int {
	out := make([]int, 0, len(goals))
	for _, g := range goals {
		out = append(out, g.Number)
	}
	return out
}

func TestSorted(t *testing.T) {
	got := Sorted([]Goal{
		{Number: 4, State: "dropped"},
		{Number: 3, State: "reached"},
		{Number: 2, State: "on-hold"},
		{Number: 1, State: "active"},
		{Number: 5, State: "active"},
		{Number: 6, State: "invented"},
	})
	want := []int{1, 5, 2, 3, 4, 6}
	for i, n := range want {
		if got[i].Number != n {
			t.Fatalf("sorted to %v, want %v", numbersOf(got), want)
		}
	}
}

func TestSortedLeavesWhatItWasGivenAlone(t *testing.T) {
	given := []Goal{{Number: 2, State: "reached"}, {Number: 1, State: "active"}}
	Sorted(given)
	if given[0].Number != 2 {
		t.Fatal("sorting changed the list it was handed")
	}
}

func TestReaches(t *testing.T) {
	g := Goal{Title: "Ship the desktop board", Description: "before the winter"}
	for _, query := range []string{"desktop", "DESKTOP", "winter"} {
		if !ReachesGoal(g, query) {
			t.Fatalf("%q did not reach it", query)
		}
	}
	if ReachesGoal(g, "spec") {
		t.Fatal("a word that is nowhere in it reached it")
	}
}
