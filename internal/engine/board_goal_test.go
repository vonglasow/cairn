package engine

import (
	"maps"
	"testing"
	"time"

	"cairn/internal/domain"
	"cairn/internal/objective"
	"cairn/internal/task"
)

func numbersOf(goals []Goal) []int {
	out := make([]int, 0, len(goals))
	for _, g := range goals {
		out = append(out, g.Number)
	}
	return out
}

func TestSorted_keepsTheOrderAPersonGave(t *testing.T) {
	got := Sorted([]Goal{
		{Number: 4, State: "dropped"},
		{Number: 3, State: "reached"},
		{Number: 2, State: "on-hold"},
		{Number: 1, State: "active"},
		{Number: 5, State: "active"},
		{Number: 6, State: "invented"},
	})
	want := []int{4, 3, 2, 1, 5, 6}
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

func TestAGoalIsReachedByWhatTheEngineSaysItCarries(t *testing.T) {
	g := Goal{Number: 3, Title: "Ship cairn", Said: []string{"q4", "active"}}
	if !ReachesGoal(g, "q4 active") {
		t.Error("what the engine said does not reach the objective")
	}
	if ReachesGoal(g, "ship") {
		t.Error("a word the engine did not say reached the objective")
	}
}

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

func TestAGoalSaysTheTimeSpentOnTheTasksServingIt(t *testing.T) {
	f := newFake()
	spent := func(number int, d time.Duration, serves ...int) task.Task {
		var a task.AsTold
		a.Numbered(number)
		a.Titled("Task")
		a.InColumn("todo")
		a.Spent(d)
		for _, n := range serves {
			a.Serves(n)
		}
		return a.Task()
	}
	f.tasks = []task.Task{spent(1, time.Hour, 1), spent(2, 30*time.Minute, 1, 2), spent(3, 2*time.Hour)}
	var one, two objective.AsTold
	one.Numbered(1)
	one.Titled("Ship it")
	two.Numbered(2)
	two.Titled("Docs")
	f.goals = []objective.Objective{one.Objective(), two.Objective()}
	found, err := Look(f)
	if err != nil {
		t.Fatal(err)
	}
	if got := found.Goals[0].Spent; got != 90*time.Minute {
		t.Errorf("objective 1 says %v spent, want 1h30m", got)
	}
	if got := found.Goals[1].Spent; got != 30*time.Minute {
		t.Errorf("objective 2 says %v spent, want 30m", got)
	}
}

func TestAGoalCountsTheStepsOfItsTasksApart(t *testing.T) {
	g := Goal{Done: 1, Total: 3, Steps: domain.Tally{Total: 2, Done: 1}}
	if got := g.Counted(); got != "0/1 · 1/2 steps" {
		t.Errorf("Counted = %q, want the task and its steps apart", got)
	}
	if got := (Goal{Done: 1, Total: 2}).Counted(); got != "1/2" {
		t.Errorf("Counted with no step = %q, want 1/2", got)
	}
}

func TestStepped_movesAnObjectiveAmongItsSiblingsAndStopsAtTheEnds(t *testing.T) {
	goals := []Goal{{Number: 1}, {Number: 4, Parent: 1}, {Number: 5, Parent: 1}, {Number: 6, Parent: 1}, {Number: 2}, {Number: 3}}
	for _, tc := range []struct {
		n    int
		up   bool
		want map[string]string
	}{
		{5, true, map[string]string{"number": "5", "first": "true"}},
		{6, true, map[string]string{"number": "6", "after": "4"}},
		{4, false, map[string]string{"number": "4", "after": "5"}},
		{2, true, map[string]string{"number": "2", "first": "true"}},
		{1, false, map[string]string{"number": "1", "after": "2"}},
	} {
		in, ok := Stepped(goals, tc.n, tc.up)
		if !ok || !maps.Equal(in.Given, tc.want) {
			t.Errorf("Stepped(%d, up %v) = %v %v, want %v", tc.n, tc.up, in.Given, ok, tc.want)
		}
	}
	for _, tc := range []struct {
		n  int
		up bool
	}{{4, true}, {6, false}, {1, true}, {3, false}} {
		if _, ok := Stepped(goals, tc.n, tc.up); ok {
			t.Errorf("Stepped(%d, up %v) moved, want it to stop at the end", tc.n, tc.up)
		}
	}
}

func TestDropped_saysWhereAnObjectiveLetGoOverARowGoes(t *testing.T) {
	goals := []Goal{{Number: 1, Children: []int{4, 5}}, {Number: 4, Parent: 1}, {Number: 5, Parent: 1}, {Number: 2}, {Number: 3}}
	at := func(n int) int { return n }
	for _, tc := range []struct {
		name    string
		dragged int
		over    int
		zone    DropZone
		want    map[string]string
	}{
		{"before the first level one", 3, at(1), DropBefore, map[string]string{"number": "3", "parent": "none", "first": "true"}},
		{"after a level one", 1, at(2), DropAfter, map[string]string{"number": "1", "after": "2"}},
		{"before a sibling", 3, at(2), DropBefore, map[string]string{"number": "3", "after": "1"}},
		{"into a parent", 2, at(1), DropInto, map[string]string{"number": "2", "parent": "1", "last": "true"}},
		{"before a sub-objective", 2, at(5), DropBefore, map[string]string{"number": "2", "after": "4"}},
		{"a parent over a sub-objective goes by its parent", 1, at(5), DropAfter, map[string]string{"number": "1", "after": "1"}},
		{"into a sub-objective is after it", 2, at(4), DropInto, map[string]string{"number": "2", "after": "4"}},
		{"below every row is last at level one", 4, 0, DropAfter, map[string]string{"number": "4", "parent": "none", "last": "true"}},
	} {
		if _, ok := Dropped(goals, 1, 3, DropInto); ok {
			t.Errorf("a parent dropped into another moved, want nothing to move: two levels at most")
		}
		in, ok := Dropped(goals, tc.dragged, tc.over, tc.zone)
		if tc.name == "a parent over a sub-objective goes by its parent" {
			if ok {
				t.Errorf("%s: moved %v, want nothing since it lands where it is", tc.name, in.Given)
			}
			continue
		}
		if !ok || !maps.Equal(in.Given, tc.want) {
			t.Errorf("%s: Dropped = %v %v, want %v", tc.name, in.Given, ok, tc.want)
		}
	}
}

func TestDropped_belowEveryRowIsNothingForTheLastTopObjective(t *testing.T) {
	goals := []Goal{{Number: 1, Children: []int{3}}, {Number: 3, Parent: 1}, {Number: 2}}
	if _, ok := Dropped(goals, 2, 0, DropAfter); ok {
		t.Error("the last top objective let go below every row moves")
	}
	in, ok := Dropped(goals, 1, 0, DropAfter)
	if want := map[string]string{"number": "1", "parent": "none", "last": "true"}; !ok || !maps.Equal(in.Given, want) {
		t.Errorf("the first top objective let go below every row: %v %v", in.Given, ok)
	}
}

func TestStepped_movesTheFirstDownAndNothingNotListed(t *testing.T) {
	goals := []Goal{{Number: 1}, {Number: 2}}
	if in, ok := Stepped(goals, 1, false); !ok || in.Given["after"] != "2" {
		t.Errorf("the first moved down: %v, %v", in.Given, ok)
	}
	if _, ok := Stepped(goals, 9, false); ok {
		t.Error("an objective not listed moved")
	}
}
