package engine

import (
	"testing"
	"time"

	"cairn/internal/domain"
	"cairn/internal/objective"
	"cairn/internal/task"
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
