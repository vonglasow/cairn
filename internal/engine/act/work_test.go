package act

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
