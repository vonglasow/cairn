package trace

// What a period adds up to, on acts built at chosen moments.
//
// This is the one part of the trace no scenario can hold: everything a scenario
// does happens in the same second, so a duration asserted there would be zero
// whatever the arithmetic did. The moments here are fabricated for that reason,
// and it is the only reason.

import (
	"testing"
	"time"
)

// at is a moment on the day these tests take place.
func at(hour, minute int) time.Time { return atOn(3, hour, minute) }

// atOn is a moment on another day of the same month.
func atOn(day, hour, minute int) time.Time {
	return time.Date(2026, 3, day, hour, minute, 0, 0, time.UTC)
}

// dayOf is the start of a day of that month, as a timesheet is asked for.
func dayOf(day int) time.Time { return atOn(day, 0, 0) }

// took builds an act of the focus being taken on a task, on the day these
// tests take place; tookOn on another.
func took(task, hour, minute int) Act { return tookOn(task, 3, hour, minute) }

func tookOn(task, day, hour, minute int) Act {
	var a AsTold
	a.On(task)
	a.Did(FocusTaken)
	a.NothingNamed()
	a.WasNothing()
	a.BecameNothing()
	a.ByHand()
	a.At(atOn(day, hour, minute))
	return a.Act()
}

// stopped builds an act of the focus being let go with nothing taking over;
// stoppedOn on another day.
func stopped(task, hour, minute int) Act { return stoppedOn(task, 3, hour, minute) }

func stoppedOn(task, day, hour, minute int) Act {
	var a AsTold
	a.On(task)
	a.Did(FocusReleased)
	a.NothingNamed()
	a.WasNothing()
	a.BecameNothing()
	a.ByHand()
	a.At(atOn(day, hour, minute))
	return a.Act()
}

// somethingElse builds an act that is not about the focus at all, so a test can
// check the arithmetic ignores it.
func somethingElse(task, hour, minute int) Act {
	var a AsTold
	a.On(task)
	a.Did(Changed)
	a.Changing("priority")
	a.Was("normal")
	a.Became("high")
	a.ByHand()
	a.At(at(hour, minute))
	return a.Act()
}

// added builds time added by hand for a day, recorded at a moment of its own.
func added(task int, day, span string, recorded time.Time) Act {
	var a AsTold
	a.On(task)
	a.Did(TimeAdded)
	a.NothingNamed()
	a.Was(day)
	a.Became(span)
	a.ByHand()
	a.At(recorded)
	return a.Act()
}

// worked reads what a sheet says of one task on one day, and whether it was
// left in focus that day; nothing when the day or the task is not written.
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

// totalOf is what a sheet says a day came to, nothing when it is not written.
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

// Two tasks in focus at once share the minutes they overlap: a day comes to
// what the clock says, never to more.
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

// Each task is rounded to the nearest quarter of an hour, half a quarter up,
// and the day is the sum of what was rounded. A task that rounds to nothing is
// not written; its day still is, coming to nothing.
func TestTimed_roundsEachTaskToTheQuarterAndAddsThoseUp(t *testing.T) {
	s := Timed([]Act{
		took(1, 9, 0), stopped(1, 9, 50),
		took(2, 10, 0), stopped(2, 10, 53),
		took(3, 11, 0), stopped(3, 11, 7),
	}, dayOf(3), dayOf(3), at(18, 0))
	one, _, _ := worked(s, dayOf(3), 1)
	two, _, _ := worked(s, dayOf(3), 2)
	_, _, three := worked(s, dayOf(3), 3)
	if one != 45*time.Minute || two != time.Hour || three || totalOf(s, dayOf(3)) != 105*time.Minute {
		t.Errorf("task 1 %s, task 2 %s, task 3 written %v, the day %s", one, two, three, totalOf(s, dayOf(3)))
	}
	half := Timed([]Act{took(1, 9, 0), stopped(1, 9, 7), took(1, 9, 30), stopped(1, 9, 30)}, dayOf(3), dayOf(3), at(18, 0))
	if _, _, written := worked(half, dayOf(3), 1); written || len(half.Days) != 1 || totalOf(half, dayOf(3)) != 0 {
		t.Errorf("seven minutes were written: %+v", half)
	}
	edge := Timed([]Act{took(1, 9, 0)}, dayOf(3), dayOf(3), at(9, 7).Add(30*time.Second))
	if took, _, _ := worked(edge, dayOf(3), 1); took != 15*time.Minute {
		t.Errorf("seven and a half minutes round to %s, want a quarter", took)
	}
}

// A focus left open stops at the first midnight after it was taken, and says
// so on the day it was left open; the days after it count nothing of it.
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

// A focus still open today runs to now: it is work being done, and nothing
// says it was forgotten.
func TestTimed_runsAFocusOpenTodayToNow(t *testing.T) {
	s := Timed([]Act{took(1, 9, 0)}, dayOf(3), dayOf(3), at(10, 0))
	if took, left, _ := worked(s, dayOf(3), 1); took != time.Hour || left {
		t.Errorf("a focus open today took %s, left open %v; want 1h, not left", took, left)
	}
}

// Time added by hand counts on the day it was added for, whenever it was
// recorded; less than nothing takes some back, and never below nothing.
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

// Acts are read in the order they happened, not the order they were written
// down: a focus put right after the fact is recorded after what followed it.
func TestTimed_readsActsInTheOrderTheyHappened(t *testing.T) {
	s := Timed([]Act{stopped(1, 10, 0), took(1, 9, 0)}, dayOf(3), dayOf(3), at(18, 0))
	if took, _, _ := worked(s, dayOf(3), 1); took != time.Hour {
		t.Errorf("a focus written down after its release took %s, want 1h", took)
	}
}

// Every day asked for is written and no other, in the order they ran, and
// taking up what is already in focus opens nothing.
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

// Written the way a sheet is filled in: hours and minutes, the minutes always
// two digits once there are hours, and nothing of a part that is nought.
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
