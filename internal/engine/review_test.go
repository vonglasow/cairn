package engine

import (
	"testing"
	"time"

	"cairn/internal/domain"
	"cairn/internal/objective"
	"cairn/internal/task"
	"cairn/internal/trace"
)

func served(number int, state domain.State, spent time.Duration, due time.Time, archived bool) task.Task {
	var a task.AsTold
	a.Numbered(number)
	a.Titled("Task")
	a.InColumn(state)
	a.Spent(spent)
	a.Due(due)
	if archived {
		a.Archived()
	}
	return a.Task()
}

func TestAReviewAddsTheTimeOfEveryTaskAndCountsWhatIsLeft(t *testing.T) {
	soon := time.Date(2026, 11, 5, 0, 0, 0, 0, time.UTC)
	later := time.Date(2026, 11, 20, 0, 0, 0, 0, time.UTC)
	under := []task.Task{
		served(1, domain.Done, time.Hour, time.Time{}, false),
		served(2, "todo", 30*time.Minute, later, false),
		served(3, "backlog", 0, soon, false),
		served(4, "todo", 15*time.Minute, soon.AddDate(0, 0, -1), true),
		served(5, "todo", 0, time.Time{}, false),
	}
	got := reviewOf(under)
	want := objective.Review{Spent: time.Hour + 45*time.Minute, Left: objective.Left{Tasks: 3, Due: soon, On: 3}}
	if got != want {
		t.Errorf("the review is %+v, want %+v", got, want)
	}
	if left := leftOf(under); len(left) != 3 || left[0] != 2 || left[1] != 3 || left[2] != 5 {
		t.Errorf("what is left is %v", left)
	}
	if got := reviewOf(nil); got != (objective.Review{}) {
		t.Errorf("nothing under it reviews as %+v", got)
	}
}

func TestThePeriodFoldsTheTimeOfItsTasksIntoMonths(t *testing.T) {
	sheet := trace.Sheet{Days: []trace.Day{
		{Day: time.Date(2026, 8, 4, 0, 0, 0, 0, time.UTC), Tasks: []trace.Worked{{Task: 1, Took: time.Hour}, {Task: 9, Took: 3 * time.Hour}}},
		{Day: time.Date(2026, 8, 20, 0, 0, 0, 0, time.UTC), Tasks: []trace.Worked{{Task: 2, Took: 30 * time.Minute}}},
		{Day: time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC), Tasks: []trace.Worked{{Task: 2, Took: 45 * time.Minute}}},
		{Day: time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC), Tasks: []trace.Worked{{Task: 9, Took: time.Hour}}},
	}}
	got := monthsOf(sheet, map[int]bool{1: true, 2: true})
	want := []objective.Month{
		{Month: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC), Spent: 90 * time.Minute},
		{Month: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC), Spent: 45 * time.Minute},
	}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("the months are %+v, want %+v", got, want)
	}
}

func TestThePeriodNamesWhatWasPutInDoneWithinItAndIsStillThere(t *testing.T) {
	first := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	last := time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)
	act := func(n int, became string, at time.Time) trace.Act {
		return trace.Made(trace.Carried{Task: n, Did: trace.Changed, Field: "column", Was: "todo", Became: became, At: at})
	}
	acts := []trace.Act{
		act(1, "done", first.Add(-time.Hour)),
		act(2, "done", first.Add(2*time.Hour)),
		act(3, "done", first.Add(3*time.Hour)),
		act(4, "done", last.Add(30*time.Hour)),
		act(5, "done", first.Add(5*time.Hour)),
		act(5, "todo", first.Add(6*time.Hour)),
		act(5, "done", first.AddDate(0, 0, 3)),
		act(6, "done", first.Add(time.Hour)),
	}
	under := []task.Task{
		served(1, domain.Done, 0, time.Time{}, false),
		served(2, domain.Done, 0, time.Time{}, false),
		served(3, "todo", 0, time.Time{}, false),
		served(4, domain.Done, 0, time.Time{}, false),
		served(5, domain.Done, 0, time.Time{}, true),
	}
	got := finishedIn(acts, under, first, last)
	if len(got) != 2 || got[0].Number != 2 || got[1].Number != 5 || !got[1].On.Equal(first.AddDate(0, 0, 3)) {
		t.Errorf("done in the period: %+v", got)
	}
}
