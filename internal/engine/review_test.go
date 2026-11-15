package engine

import (
	"cairn/internal/engine/calendar"
	"cairn/internal/engine/item"
	"testing"
	"time"

	"cairn/internal/engine/act"
	"cairn/internal/engine/task"
)

func served(number int, state item.State, spent time.Duration, due time.Time, archived bool) task.Task {
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

func TestThePeriodNamesWhatWasPutInDoneWithinItAndIsStillThere(t *testing.T) {
	first := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	last := time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)
	movedTo := func(n int, became string, at time.Time) act.Act {
		return act.Made(act.Carried{Task: n, Did: act.Changed, Field: "column", Was: "todo", Became: became, At: at})
	}
	acts := []act.Act{
		movedTo(1, "done", first.Add(-time.Hour)),
		movedTo(2, "done", first.Add(2*time.Hour)),
		movedTo(3, "done", first.Add(3*time.Hour)),
		movedTo(4, "done", last.Add(30*time.Hour)),
		movedTo(5, "done", first.Add(5*time.Hour)),
		movedTo(5, "todo", first.Add(6*time.Hour)),
		movedTo(5, "done", first.AddDate(0, 0, 3)),
		movedTo(6, "done", first.Add(time.Hour)),
	}
	under := []task.Task{
		served(1, item.Done, 0, time.Time{}, false),
		served(2, item.Done, 0, time.Time{}, false),
		served(3, "todo", 0, time.Time{}, false),
		served(4, item.Done, 0, time.Time{}, false),
		served(5, item.Done, 0, time.Time{}, true),
	}
	got := finishedIn(acts, under, first, last)
	if len(got) != 2 || got[0].Number != 2 || got[1].Number != 5 || !got[1].On.Equal(first.AddDate(0, 0, 3)) {
		t.Errorf("done in the period: %+v", got)
	}
}

func TestObjectiveReport_reviewsTheTimeAndTheTasksDoneInThePeriod(t *testing.T) {
	e := openEngine(t)
	addObjective(t, e, "Release")
	serving(t, e, "Ship", 1)
	today := calendar.FormatDay(time.Now())
	if _, err := e.Ask(TimesheetAdd, given("number", "1", "time", "1h", "on", today)); err != nil {
		t.Fatal(err)
	}
	if _, err := e.Ask(TaskEdit, given("number", "1", "state", "done")); err != nil {
		t.Fatal(err)
	}
	a, err := e.Ask(objectiveReport, given("number", "1", "since", calendar.FormatDay(time.Now().AddDate(0, 0, -7)), "until", today))
	if err != nil {
		t.Fatal(err)
	}
	if a.Period == nil || a.Period.Spent() != time.Hour || len(a.Period.Done) != 1 || a.Period.Done[0].Number != 1 {
		t.Errorf("period %+v, want 1h and task 1 done", a.Period)
	}
}
