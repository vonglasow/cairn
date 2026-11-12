package trace

import (
	"testing"
	"time"
)

func at(hour, minute int) time.Time { return atOn(3, hour, minute) }

func atOn(day, hour, minute int) time.Time {
	return time.Date(2026, 3, day, hour, minute, 0, 0, time.UTC)
}

func dayOf(day int) time.Time { return atOn(day, 0, 0) }

func took(task, hour, minute int) Act { return tookOn(task, 3, hour, minute) }

func tookOn(task, day, hour, minute int) Act {
	var a AsTold
	a.On(task)
	a.Did(FocusTaken)
	a.At(atOn(day, hour, minute))
	return a.Act()
}

func stopped(task, hour, minute int) Act { return stoppedOn(task, 3, hour, minute) }

func stoppedOn(task, day, hour, minute int) Act {
	var a AsTold
	a.On(task)
	a.Did(FocusReleased)
	a.At(atOn(day, hour, minute))
	return a.Act()
}

func somethingElse(task, hour, minute int) Act {
	var a AsTold
	a.On(task)
	a.Did(Changed)
	a.Changing("priority")
	a.Was("normal")
	a.Became("high")
	a.At(at(hour, minute))
	return a.Act()
}

func added(task int, day, span string, recorded time.Time) Act {
	var a AsTold
	a.On(task)
	a.Did(TimeAdded)
	a.Was(day)
	a.Became(span)
	a.At(recorded)
	return a.Act()
}

func worked(s Sheet, day time.Time, task int) (time.Duration, bool, bool) {
	for _, d := range s.Days {
		if !d.Day.Equal(day) {
			continue
		}
		for _, w := range d.Tasks {
			if w.Task == task {
				return w.Took, w.LeftOpen, true
			}
		}
	}
	return 0, false, false
}

func totalOf(s Sheet, day time.Time) time.Duration {
	for _, d := range s.Days {
		if d.Day.Equal(day) {
			return d.Total
		}
	}
	return 0
}

func TestTimed_countsTheTimeATaskWasInFocus(t *testing.T) {
	s := Timed([]Act{took(1, 9, 0), somethingElse(1, 9, 10), stopped(1, 10, 30)}, dayOf(3), dayOf(3), at(18, 0))
	if took, _, _ := worked(s, dayOf(3), 1); took != 90*time.Minute || totalOf(s, dayOf(3)) != 90*time.Minute {
		t.Errorf("task 1 took %s, the day %s; want 1h30 both", took, totalOf(s, dayOf(3)))
	}
}

func TestTimed_sharesTheMinutesTwoTasksOverlap(t *testing.T) {
	s := Timed([]Act{took(1, 9, 0), took(2, 10, 0), stopped(1, 11, 0), stopped(2, 11, 0)},
		dayOf(3), dayOf(3), at(18, 0))
	one, _, _ := worked(s, dayOf(3), 1)
	two, _, _ := worked(s, dayOf(3), 2)
	if one != 90*time.Minute || two != 30*time.Minute || totalOf(s, dayOf(3)) != 2*time.Hour {
		t.Errorf("task 1 %s, task 2 %s, the day %s; want 1h30, 30m and 2h", one, two, totalOf(s, dayOf(3)))
	}
	if tasks := s.Days[0].Tasks; tasks[0].Task != 1 || tasks[1].Task != 2 {
		t.Errorf("a day writes its tasks in the order %+v, want their numbers' order", tasks)
	}
}

func TestTimed_saysEachTaskToTheMinuteBegunAndAddsThoseUp(t *testing.T) {
	s := Timed([]Act{
		took(1, 9, 0), stopped(1, 9, 50),
		took(2, 10, 0), stopped(2, 10, 53),
		took(3, 11, 0), stopped(3, 11, 7),
	}, dayOf(3), dayOf(3), at(18, 0))
	one, _, _ := worked(s, dayOf(3), 1)
	two, _, _ := worked(s, dayOf(3), 2)
	three, _, _ := worked(s, dayOf(3), 3)
	if one != 50*time.Minute || two != 53*time.Minute || three != 7*time.Minute || totalOf(s, dayOf(3)) != 110*time.Minute {
		t.Errorf("task 1 %s, task 2 %s, task 3 %s, the day %s", one, two, three, totalOf(s, dayOf(3)))
	}
	edge := Timed([]Act{took(1, 9, 0)}, dayOf(3), dayOf(3), at(9, 7).Add(time.Second))
	if took, _, _ := worked(edge, dayOf(3), 1); took != 8*time.Minute {
		t.Errorf("seven minutes and a second read %s, want the eighth minute begun", took)
	}
	none := Timed([]Act{took(1, 9, 30), stopped(1, 9, 30)}, dayOf(3), dayOf(3), at(18, 0))
	if _, _, written := worked(none, dayOf(3), 1); written {
		t.Errorf("nothing at all was written: %+v", none)
	}
}

func TestSpent_isWhatTheDaysOfTheSheetComeTo(t *testing.T) {
	acts := []Act{
		took(1, 9, 0), stoppedAfter(1, at(9, 0).Add(10*time.Second)),
		tookOn(1, 4, 9, 0), stoppedAfter(1, atOn(4, 9, 0).Add(10*time.Second)),
	}
	if spent := Spent(acts, atOn(5, 0, 0)); spent[1] != 2*time.Minute {
		t.Errorf("ten seconds on each of two days spent %s, want the minute each day says", spent[1])
	}
}

func stoppedAfter(task int, when time.Time) Act {
	var a AsTold
	a.On(task)
	a.Did(FocusReleased)
	a.At(when)
	return a.Act()
}

func TestTimed_stopsAFocusLeftOpenAtMidnight(t *testing.T) {
	open := Timed([]Act{took(1, 17, 0)}, dayOf(3), dayOf(4), atOn(5, 12, 0))
	length, left, _ := worked(open, dayOf(3), 1)
	if length != 7*time.Hour || !left || totalOf(open, dayOf(4)) != 0 || len(open.Days) != 2 || len(open.Days[1].Tasks) != 0 {
		t.Errorf("a focus left open took %s, left open %v, the next day %s", length, left, totalOf(open, dayOf(4)))
	}
	released := Timed([]Act{took(1, 17, 0), stoppedOn(1, 4, 1, 0)}, dayOf(3), dayOf(4), atOn(5, 12, 0))
	if took, left, _ := worked(released, dayOf(3), 1); took != 7*time.Hour || !left || totalOf(released, dayOf(4)) != 0 {
		t.Errorf("a focus let go the next morning took %s, left open %v, the next day %s", took, left, totalOf(released, dayOf(4)))
	}
}

func TestTimed_runsAFocusOpenTodayToNow(t *testing.T) {
	s := Timed([]Act{took(1, 9, 0)}, dayOf(3), dayOf(3), at(10, 0))
	if took, left, _ := worked(s, dayOf(3), 1); took != time.Hour || left {
		t.Errorf("a focus open today took %s, left open %v; want 1h, not left", took, left)
	}
}

func TestTimed_countsTimeAddedByHandOnItsDay(t *testing.T) {
	later := atOn(6, 18, 0)
	s := Timed([]Act{
		added(1, "2026-03-03", "1h", later),
		added(2, "2026-03-03", "1h", later), added(2, "2026-03-03", "-30m", later),
		added(3, "2026-03-03", "-30m", later),
		added(4, "2026-03-04", "2h", later),
	}, dayOf(3), dayOf(3), atOn(6, 19, 0))
	one, _, _ := worked(s, dayOf(3), 1)
	two, _, _ := worked(s, dayOf(3), 2)
	_, _, three := worked(s, dayOf(3), 3)
	_, _, four := worked(s, dayOf(3), 4)
	if one != time.Hour || two != 30*time.Minute || three || four {
		t.Errorf("by hand: task 1 %s, task 2 %s, task 3 written %v, task 4 written %v", one, two, three, four)
	}
}

func TestTimed_readsActsInTheOrderTheyHappened(t *testing.T) {
	s := Timed([]Act{stopped(1, 10, 0), took(1, 9, 0)}, dayOf(3), dayOf(3), at(18, 0))
	if took, _, _ := worked(s, dayOf(3), 1); took != time.Hour {
		t.Errorf("a focus written down after its release took %s, want 1h", took)
	}
}

func TestTimed_writesTheDaysAskedForInOrder(t *testing.T) {
	s := Timed([]Act{
		took(1, 9, 0), took(1, 9, 30), stopped(1, 10, 0),
		tookOn(1, 4, 9, 0), stoppedOn(1, 4, 11, 0),
		tookOn(1, 5, 9, 0), stoppedOn(1, 5, 12, 0),
	}, dayOf(3), dayOf(4), atOn(6, 0, 0))
	if len(s.Days) != 2 || !s.Days[0].Day.Equal(dayOf(3)) || !s.Days[1].Day.Equal(dayOf(4)) {
		t.Fatalf("the sheet writes %+v", s.Days)
	}
	if totalOf(s, dayOf(3)) != time.Hour || totalOf(s, dayOf(4)) != 2*time.Hour {
		t.Errorf("the days come to %s and %s", totalOf(s, dayOf(3)), totalOf(s, dayOf(4)))
	}
}

func TestHours_writesATimeTheWayASheetIsFilledIn(t *testing.T) {
	for d, want := range map[time.Duration]string{
		0: "0m", 45 * time.Minute: "45m", time.Hour: "1h", 90 * time.Minute: "1h30",
		11*time.Hour + 30*time.Minute: "11h30", 2*time.Hour + 5*time.Minute: "2h05",
	} {
		if got := Hours(d); got != want {
			t.Errorf("Hours(%s) = %q, want %q", d, got, want)
		}
	}
}

func TestTimed_readsOnPastTimeAddedThatCannotBeRead(t *testing.T) {
	later := atOn(6, 18, 0)
	s := Timed([]Act{
		added(1, "someday", "1h", later),
		added(2, "2026-03-03", "ages", later.Add(time.Minute)),
		added(3, "2026-03-03", "1h", later.Add(2*time.Minute)),
	}, dayOf(3), dayOf(3), atOn(6, 19, 0))
	_, _, one := worked(s, dayOf(3), 1)
	_, _, two := worked(s, dayOf(3), 2)
	three, _, _ := worked(s, dayOf(3), 3)
	if one || two || three != time.Hour {
		t.Errorf("task 1 written %v, task 2 written %v, task 3 %s; want only task 3, an hour", one, two, three)
	}
}

func TestTimed_saysLeftOpenOnlyOfTheDayItWasLeft(t *testing.T) {
	s := Timed([]Act{took(1, 17, 0), stoppedOn(1, 4, 1, 0), tookOn(1, 4, 9, 0), stoppedOn(1, 4, 10, 0)},
		dayOf(3), dayOf(4), atOn(5, 12, 0))
	if _, left, _ := worked(s, dayOf(4), 1); left {
		t.Error("the day after says the task was left open on it")
	}
}

func TestTimed_writesWhatComesAfterATaskRoundedToNothing(t *testing.T) {
	s := Timed([]Act{took(1, 9, 0), stopped(1, 9, 5), took(2, 10, 0), stopped(2, 11, 0)}, dayOf(3), dayOf(3), at(12, 0))
	if took, _, _ := worked(s, dayOf(3), 2); took != time.Hour {
		t.Errorf("task 2 took %s, want the hour after task 1's five minutes", took)
	}
}

func TestSpent_isTheTimeInFocusAndTheTimeAddedByHand(t *testing.T) {
	acts := []Act{
		took(1, 9, 0), stopped(1, 10, 30),
		added(1, "2026-03-02", "30m", at(11, 0)),
		added(1, "", "2h", at(11, 0)),
		took(2, 14, 0), stopped(2, 14, 20),
	}
	spent := Spent(acts, at(18, 0))
	if spent[1] != 4*time.Hour {
		t.Errorf("task 1 spent %s, want 4h", spent[1])
	}
	if spent[2] != 20*time.Minute {
		t.Errorf("task 2 spent %s, want 20m", spent[2])
	}
}

func TestSpent_sharesTheMinutesTwoTasksOverlap(t *testing.T) {
	spent := Spent([]Act{took(1, 9, 0), took(2, 9, 0), stopped(1, 10, 0), stopped(2, 10, 0)}, at(18, 0))
	if spent[1] != 30*time.Minute || spent[2] != 30*time.Minute {
		t.Errorf("spent %v, want 30m each", spent)
	}
}

func TestSpent_countsAMinuteBegun(t *testing.T) {
	var a AsTold
	a.On(1)
	a.Did(FocusReleased)
	a.At(at(9, 0).Add(10 * time.Second))
	spent := Spent([]Act{took(1, 9, 0), a.Act()}, at(18, 0))
	if spent[1] != time.Minute {
		t.Errorf("ten seconds in focus spent %s, want 1m", spent[1])
	}
}

func TestSpent_neverFallsBelowNothing(t *testing.T) {
	spent := Spent([]Act{added(1, "2026-03-03", "30m", at(9, 0)), added(1, "2026-03-03", "-1h", at(9, 0))}, at(18, 0))
	if spent[1] != 0 {
		t.Errorf("taking back more than was added spent %s, want nothing", spent[1])
	}
}

func TestSpent_runsAFocusStillOpenToNow(t *testing.T) {
	spent := Spent([]Act{took(1, 9, 0)}, at(9, 45))
	if spent[1] != 45*time.Minute {
		t.Errorf("a focus still open spent %s, want 45m", spent[1])
	}
}

func TestSpent_runsOnThroughARenewedFocus(t *testing.T) {
	renewed := Made(Carried{Task: 1, Did: FocusRenewed, At: at(10, 0)})
	acts := []Act{took(1, 9, 0), renewed, stopped(1, 11, 0)}
	if spent := Spent(acts, at(18, 0)); spent[1] != 2*time.Hour {
		t.Errorf("task 1 spent %s across a renewal, want 2h", spent[1])
	}
}

func TestTimed_leavesOutATaskDeletedWithItsTime(t *testing.T) {
	gone := Made(Carried{Task: 1, Did: Deleted, Was: "title: Write the spec; column: todo; " + WentWithItsTime, At: at(12, 0)})
	kept := Made(Carried{Task: 2, Did: Deleted, Was: "title: Ship it; column: todo", At: at(12, 0)})
	acts := WithoutTimeTaken([]Act{took(1, 9, 0), stopped(1, 10, 0), took(2, 10, 0), stopped(2, 11, 0), gone, kept})
	spent := Spent(acts, at(18, 0))
	if spent[1] != 0 || spent[2] != time.Hour {
		t.Errorf("spent = %v, want nothing on task 1, deleted with its time, and 1h on task 2", spent)
	}
}

func TestTitleWent_isTheTitleADeletionTookWithIt(t *testing.T) {
	gone := Made(Carried{Task: 1, Did: Deleted, Was: "title: Write the spec; column: todo"})
	if got := TitleWent([]Act{gone}, 1); got != "Write the spec" {
		t.Errorf("TitleWent = %q", got)
	}
	if got := TitleWent([]Act{gone}, 2); got != "" {
		t.Errorf("TitleWent of a task never deleted = %q", got)
	}
}

func TestTitleWent_readsNoTitleOffAnActThatWasNoDeletion(t *testing.T) {
	retitled := Made(Carried{Task: 1, Did: Changed, Field: "title", Was: "title: Draft"})
	if got := TitleWent([]Act{retitled}, 1); got != "" {
		t.Errorf("TitleWent of a task only renamed = %q, want nothing", got)
	}
}

func TestTitleWent_readsPastTheActsOfOtherTasks(t *testing.T) {
	gone := Made(Carried{Task: 1, Did: Deleted, Was: "title: Write the spec; column: todo"})
	if got := TitleWent([]Act{took(2, 9, 0), gone}, 1); got != "Write the spec" {
		t.Errorf("TitleWent behind an act of task 2 = %q", got)
	}
}

func TestSpent_leavesOutATaskThatComesToNothing(t *testing.T) {
	spent := Spent([]Act{added(1, "2026-03-03", "30m", at(9, 0)), added(1, "2026-03-03", "-30m", at(9, 0))}, at(18, 0))
	if length, there := spent[1]; there {
		t.Errorf("a task whose time was all taken back spent %s, want it left out", length)
	}
}

func TestTimed_writesWhatComesAfterATaskWithTimeTakenOff(t *testing.T) {
	for _, off := range []string{"-30m", "-1h"} {
		acts := []Act{took(1, 9, 0), stopped(1, 9, 30), added(1, "2026-03-03", off, at(12, 0)), took(2, 10, 0), stopped(2, 11, 0)}
		s := Timed(acts, dayOf(3), dayOf(3), at(12, 0))
		if took, _, there := worked(s, dayOf(3), 2); !there || took != time.Hour {
			t.Errorf("with %s off task 1, task 2 took %s, want the hour", off, took)
		}
	}
}

func timer(task int, did string, hour, minute int) Act {
	var a AsTold
	a.On(task)
	a.Did(did)
	a.At(at(hour, minute))
	return a.Act()
}

func tookWithoutATimer(task, hour, minute int) Act {
	var a AsTold
	a.On(task)
	a.Did(FocusTaken)
	a.Became(WithoutATimer)
	a.At(at(hour, minute))
	return a.Act()
}

func TestTimed_countsNothingOfAFocusTakenWithoutATimer(t *testing.T) {
	s := Timed([]Act{tookWithoutATimer(1, 9, 0), stopped(1, 11, 0)}, dayOf(3), dayOf(3), at(18, 0))
	if _, _, written := worked(s, dayOf(3), 1); written {
		t.Errorf("a focus without a timer counted: %+v", s)
	}
}

func TestTimed_countsFromTheTimerStartedAndToTheTimerStopped(t *testing.T) {
	s := Timed([]Act{
		tookWithoutATimer(1, 9, 0), timer(1, TimerStarted, 10, 0), stopped(1, 10, 30),
		took(2, 9, 0), timer(2, TimerStopped, 9, 45), stopped(2, 12, 0),
	}, dayOf(3), dayOf(3), at(18, 0))
	one, _, _ := worked(s, dayOf(3), 1)
	two, _, _ := worked(s, dayOf(3), 2)
	if one != 30*time.Minute || two != 45*time.Minute {
		t.Errorf("task 1 %s, task 2 %s; want 30m from the timer started, 45m to the timer stopped", one, two)
	}
	open := Timed([]Act{took(1, 9, 0), timer(1, TimerStarted, 9, 30)}, dayOf(3), dayOf(3), at(10, 0))
	if took, _, _ := worked(open, dayOf(3), 1); took != time.Hour {
		t.Errorf("a timer started while it runs counted %s, want the hour since the focus", took)
	}
}

func TestTimerRuns_saysWhetherTheTimerOfATaskRuns(t *testing.T) {
	for _, c := range []struct {
		name        string
		acts        []Act
		runs, known bool
	}{
		{"nothing said", []Act{somethingElse(1, 9, 0)}, false, false},
		{"focus taken", []Act{took(1, 9, 0)}, true, true},
		{"focus taken without a timer", []Act{tookWithoutATimer(1, 9, 0)}, false, true},
		{"timer stopped", []Act{took(1, 9, 0), timer(1, TimerStopped, 9, 30)}, false, true},
		{"timer started", []Act{tookWithoutATimer(1, 9, 0), timer(1, TimerStarted, 9, 30)}, true, true},
		{"focus released", []Act{took(1, 9, 0), stopped(1, 9, 30)}, false, true},
		{"another task's", []Act{took(2, 9, 0)}, false, false},
		{"another task's first", []Act{took(2, 9, 0), took(1, 9, 30)}, true, true},
	} {
		if runs, known := TimerRuns(c.acts, 1); runs != c.runs || known != c.known {
			t.Errorf("%s: runs %v, known %v; want %v, %v", c.name, runs, known, c.runs, c.known)
		}
	}
}
