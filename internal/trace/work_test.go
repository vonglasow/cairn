package trace

// What a period adds up to, on acts built at chosen moments.
//
// This is the one part of the trace no scenario can hold: everything a scenario
// does happens in the same second, so a duration asserted there would be zero
// whatever the arithmetic did. The moments here are fabricated for that reason,
// and it is the only reason.

import (
	"strings"
	"testing"
	"time"
)

// at is a moment on the day these tests take place.
func at(hour, minute int) time.Time {
	return time.Date(2026, 3, 3, hour, minute, 0, 0, time.UTC)
}

// took builds an act of the focus being taken on a task.
func took(task, hour, minute int) Act {
	var a AsTold
	a.On(task)
	a.Did(FocusTaken)
	a.NothingNamed()
	a.WasNothing()
	a.BecameNothing()
	a.ByHand()
	a.At(at(hour, minute))
	return a.Act()
}

// stopped builds an act of the focus being let go with nothing taking over.
func stopped(task, hour, minute int) Act {
	var a AsTold
	a.On(task)
	a.Did(FocusReleased)
	a.NothingNamed()
	a.WasNothing()
	a.BecameNothing()
	a.ByHand()
	a.At(at(hour, minute))
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

func TestAddedUp_measuresOneStretchFromItsEnds(t *testing.T) {
	w := AddedUp([]Act{took(1, 9, 0), stopped(1, 9, 45)}, at(18, 0))
	if got := len(w.stretches); got != 1 {
		t.Fatalf("%d stretches, want 1", got)
	}
	if got, want := w.on(1), 45*time.Minute; got != want {
		t.Errorf("time on task 1 = %s, want %s", got, want)
	}
}

// Two stretches on one task add up rather than replacing each other: coming back
// to a task is the ordinary shape of a day.
func TestAddedUp_addsUpEveryStretchOnATask(t *testing.T) {
	w := AddedUp([]Act{
		took(1, 9, 0), stopped(1, 9, 30),
		took(1, 14, 0), stopped(1, 14, 20),
	}, at(18, 0))
	if got := len(w.stretches); got != 2 {
		t.Fatalf("%d stretches, want 2", got)
	}
	if got, want := w.on(1), 50*time.Minute; got != want {
		t.Errorf("time on task 1 = %s, want %s", got, want)
	}
}

// The total of the stretches is not the length of the day, and this is the
// assertion that says so: nothing is in focus during a meeting, so a day has
// gaps, and cairn knows neither when it began nor when it ended.
func TestAddedUp_doesNotCountTheGapsBetweenStretches(t *testing.T) {
	w := AddedUp([]Act{
		took(1, 9, 0), stopped(1, 9, 30),
		took(1, 11, 0), stopped(1, 11, 30),
	}, at(18, 0))
	if got, want := w.on(1), time.Hour; got != want {
		t.Errorf("time on task 1 = %s, want %s — the 90 minutes between them are not work", got, want)
	}
}

// A switch is work taken up while other work was already under way. Taken up
// with nothing else open it is a start, not a switch.
//
// It used to be read off the release instead: a release naming what took over
// was a cut, one naming nothing was a stop. Nothing names what took over any
// more — several tasks are worked on at once and their stretches overlap on
// purpose — so what attention moving looks like is a take landing on top of
// something already open.
func TestAddedUp_countsTakingUpWorkOnTopOfWorkAlreadyOpen(t *testing.T) {
	for _, tc := range []struct {
		what string
		acts []Act
		want int
	}{
		{
			"the second task was taken up while the first was still held",
			[]Act{took(1, 9, 0), took(2, 9, 30), stopped(1, 10, 0), stopped(2, 10, 30)},
			1,
		},
		{
			"the work stopped and started again later",
			[]Act{took(1, 9, 0), stopped(1, 9, 30), took(2, 11, 0), stopped(2, 11, 30)},
			0,
		},
		{
			"a third was taken up on top of two already held",
			[]Act{
				took(1, 9, 0), took(2, 9, 10), took(3, 9, 20),
				stopped(1, 9, 30), stopped(2, 9, 40), stopped(3, 9, 50),
			},
			2,
		},
		{
			"the same task taken up again while it is already held",
			[]Act{took(1, 9, 0), took(1, 9, 10), stopped(1, 9, 30)},
			0,
		},
	} {
		if got := AddedUp(tc.acts, at(18, 0)).switches; got != tc.want {
			t.Errorf("%s: %d switches, want %d", tc.what, got, tc.want)
		}
	}
}

// Stretches on different tasks overlap, and each is measured from its own ends
// rather than being cut short by the next task being taken up.
func TestAddedUp_measuresOverlappingStretchesSeparately(t *testing.T) {
	w := AddedUp([]Act{
		took(1, 9, 0), took(2, 9, 30), stopped(2, 10, 0), stopped(1, 11, 0),
	}, at(18, 0))
	if got, want := w.on(1), 2*time.Hour; got != want {
		t.Errorf("time on task 1 = %s, want %s", got, want)
	}
	if got, want := w.on(2), 30*time.Minute; got != want {
		t.Errorf("time on task 2 = %s, want %s", got, want)
	}
}

// A stretch still open at the end runs to now: it is work being done, not work
// that never happened.
func TestAddedUp_runsAnOpenStretchToNow(t *testing.T) {
	w := AddedUp([]Act{took(1, 9, 0)}, at(9, 25))
	if got, want := w.on(1), 25*time.Minute; got != want {
		t.Errorf("time on task 1 = %s, want %s", got, want)
	}
}

// Acts that are not about the focus are not work, and do not open or close
// anything.
func TestAddedUp_ignoresWhatIsNotAboutTheFocus(t *testing.T) {
	w := AddedUp([]Act{
		somethingElse(1, 8, 0),
		took(1, 9, 0), somethingElse(1, 9, 15), stopped(1, 9, 30),
		somethingElse(2, 10, 0),
	}, at(18, 0))
	if got := len(w.stretches); got != 1 {
		t.Errorf("%d stretches, want 1", got)
	}
	if got, want := w.on(1), 30*time.Minute; got != want {
		t.Errorf("time on task 1 = %s, want %s", got, want)
	}
}

// The tasks come back in the order their numbers run, so a summary read twice
// says the same thing in the same order.
func TestAddedUp_namesTheTasksInOneSettledOrder(t *testing.T) {
	w := AddedUp([]Act{
		took(7, 9, 0), stopped(7, 9, 10),
		took(2, 10, 0), stopped(2, 10, 10),
		took(7, 11, 0), stopped(7, 11, 10),
	}, at(18, 0))
	if got := w.onTasks(); len(got) != 2 || got[0] != 2 || got[1] != 7 {
		t.Errorf("tasks = %v, want [2 7]", got)
	}
}

func TestSummary_saysTheCountsAndTheTimePerTask(t *testing.T) {
	got := Summary(AddedUp([]Act{
		took(1, 9, 0), took(2, 9, 30),
		stopped(1, 9, 30), stopped(2, 10, 45),
	}, at(18, 0)))
	for _, want := range []string{
		"2 stretches of work on 2 task(s)", "switched 1 time(s)", "#1  30m", "#2  1h 15m",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("the summary does not carry %q:\n%s", want, got)
		}
	}
}

// A period with acts but no work has nothing to add up, and says nothing rather
// than saying zero of everything.
func TestSummary_saysNothingWhenNoWorkWasDone(t *testing.T) {
	if got := Summary(AddedUp([]Act{somethingElse(1, 9, 0)}, at(18, 0))); got != "" {
		t.Errorf("Summary = %q, want nothing", got)
	}
}

// Written down to the minute, because a stretch measured to the second says
// more than anybody knows.
func TestSpent_writesALengthOfWorkAPersonCanRead(t *testing.T) {
	for _, tc := range []struct {
		of   time.Duration
		want string
	}{
		{30 * time.Second, "1m"},
		{45 * time.Minute, "45m"},
		{time.Hour, "1h 00m"},
		{90 * time.Minute, "1h 30m"},
		{9*time.Hour + 5*time.Minute, "9h 05m"},
	} {
		if got := spent(tc.of); got != tc.want {
			t.Errorf("spent(%s) = %q, want %q", tc.of, got, tc.want)
		}
	}
}
