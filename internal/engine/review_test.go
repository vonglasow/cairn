package engine

import (
	"cairn/internal/engine/calendar"
	"cairn/internal/engine/item"
	"reflect"
	"testing"
	"time"

	"cairn/internal/engine/act"
	"cairn/internal/engine/objective"
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

func TestAReviewAddsTheTimeOfEveryTaskAndCountsWhatIsLeft(t *testing.T) {
	soon := time.Date(2026, 11, 5, 0, 0, 0, 0, time.UTC)
	later := time.Date(2026, 11, 20, 0, 0, 0, 0, time.UTC)
	under := []task.Task{
		served(1, item.Done, time.Hour, time.Time{}, false),
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
	sheet := act.Sheet{Days: []act.Day{
		{Day: time.Date(2026, 8, 4, 0, 0, 0, 0, time.UTC), Tasks: []act.Worked{{Task: 1, Took: time.Hour}, {Task: 9, Took: 3 * time.Hour}}},
		{Day: time.Date(2026, 8, 20, 0, 0, 0, 0, time.UTC), Tasks: []act.Worked{{Task: 2, Took: 30 * time.Minute}}},
		{Day: time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC), Tasks: []act.Worked{{Task: 2, Took: 45 * time.Minute}}},
		{Day: time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC), Tasks: []act.Worked{{Task: 9, Took: time.Hour}}},
	}}
	got := objective.MonthsOf(sheet, map[int]bool{1: true, 2: true})
	want := []objective.Month{
		{Month: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC), Spent: 90 * time.Minute},
		{Month: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC), Spent: 45 * time.Minute},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("the months are %+v, want %+v", got, want)
	}
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

func TestMonthsOf_countsADayPastOneWhereNothingServedWasDone(t *testing.T) {
	first := time.Date(2026, 10, 1, 0, 0, 0, 0, time.Local)
	sheet := act.Sheet{Days: []act.Day{
		{Day: first, Tasks: []act.Worked{{Task: 2, Took: time.Hour}}},
		{Day: first.AddDate(0, 0, 1), Tasks: []act.Worked{{Task: 1, Took: 30 * time.Minute}}},
	}}
	got := objective.MonthsOf(sheet, map[int]bool{1: true})
	if len(got) != 1 || !got[0].Month.Equal(first) || got[0].Spent != 30*time.Minute {
		t.Errorf("months %+v, want 30m in October", got)
	}
}

func TestFinishedIn_readsDoneOnlyOffAColumnMoved(t *testing.T) {
	e := openEngine(t)
	done, err := e.add("Ship", Changes{State: of("done")})
	if err != nil {
		t.Fatal(err)
	}
	first := time.Date(2026, 10, 1, 0, 0, 0, 0, time.Local)
	renamed := []act.Act{anAct("changed", "title", "Ship", "done", "", first.AddDate(0, 0, 2))}
	if got := finishedIn(renamed, []task.Task{done}, first, first.AddDate(0, 0, 9)); len(got) != 0 {
		t.Errorf("a title that became done was read as the task finished: %+v", got)
	}
}

func TestFinishedIn_leavesOutATaskPutInDoneBeforeThePeriod(t *testing.T) {
	e := openEngine(t)
	done, err := e.add("Ship", Changes{State: of("done")})
	if err != nil {
		t.Fatal(err)
	}
	first := time.Date(2026, 10, 1, 0, 0, 0, 0, time.Local)
	last := first.AddDate(0, 0, 9)
	before := []act.Act{anAct("changed", "column", "todo", "done", "", first.AddDate(0, 0, -1))}
	if got := finishedIn(before, []task.Task{done}, first, last); len(got) != 0 {
		t.Errorf("done before the period, found %+v", got)
	}
	within := []act.Act{anAct("changed", "column", "todo", "done", "", first.AddDate(0, 0, 2))}
	if got := finishedIn(within, []task.Task{done}, first, last); len(got) != 1 || got[0].Number != 1 {
		t.Errorf("done in the period, found %+v", got)
	}
}

func TestOfReview_saysTheParentItselfAndEverySubObjective(t *testing.T) {
	month := objective.Month{Month: time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC), Spent: 2 * time.Hour,
		Split: []objective.Part{{Number: 0, Spent: time.Hour}, {Number: 2, Spent: time.Hour}}}
	r := ofReview(builtObjective(t, 1, "Ship"), nil, item.Progress{}, objective.Period{Months: []objective.Month{month}})
	if len(r.Months) != 1 || r.Months[0].Itself != "1h" || len(r.Months[0].SubObjectives) != 1 || r.Months[0].SubObjectives[0].Number != 2 {
		t.Errorf("the months read %+v", r.Months)
	}
}

func anAct(did, field, was, became, through string, at time.Time) act.Act {
	return act.Made(act.Carried{Task: 1, Did: did, Field: field, Was: was, Became: became, Through: through, At: at})
}
