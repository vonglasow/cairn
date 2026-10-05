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

// An objective was searched on two fields where a task is searched on all of
// them. One word, "search", and two meanings a tab apart.
func TestAGoalIsSearchedTheWayACardIs(t *testing.T) {
	g := Goal{Number: 3, Title: "Ship cairn", State: "active", Priority: "high",
		Labels: []string{"q4"}}
	for _, word := range []string{"q4", "active", "high", "3"} {
		if !ReachesGoal(g, word) {
			t.Errorf("a search for %q does not reach the objective", word)
		}
	}
	if ReachesGoal(g, "ship q3") {
		t.Error("a word the objective does not carry still reached it")
	}
}

// How much of an objective is done, written one way on both boards, and its
// share worked out by the domain's rule — which keeps 0 and 100 for nothing
// done and everything done. The window divided on its own, and called 299 of
// 300 a hundred per cent finished.
func TestGoal_saysHowMuchIsDone(t *testing.T) {
	g := Goal{Done: 299, Total: 300}
	if g.Counted() != "299/300" {
		t.Errorf("Counted = %q", g.Counted())
	}
	if g.Percent() != 99 {
		t.Errorf("Percent = %d, want 99: one thing is still left", g.Percent())
	}
	if (Goal{}).Percent() != 0 {
		t.Error("nothing under it is not a share of anything")
	}
}
