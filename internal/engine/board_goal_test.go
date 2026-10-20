package engine

import (
	"testing"
	"time"

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
