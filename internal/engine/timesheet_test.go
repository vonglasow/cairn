package engine

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func given(pairs ...string) Input {
	in := Input{Given: map[string]string{}}
	for i := 0; i+1 < len(pairs); i += 2 {
		in.Given[pairs[i]] = pairs[i+1]
	}
	return in
}

// Time added by hand on a day past comes to that day of a timesheet: in the
// answer, in what a person reads and in what a machine reads. How focus
// becomes time is pinned in internal/trace, where moments can be chosen.
func TestTimesheet_measuresWhatWasDoneOnADayPast(t *testing.T) {
	path := storeAt(t)
	if _, err := Do(TaskAdd, path, given("title", "one")); err != nil {
		t.Fatal(err)
	}
	for _, length := range []string{"90m", "30m"} {
		if _, err := Do(TimesheetAdd, path, given("number", "1", "time", length, "on", "2026-09-29")); err != nil {
			t.Fatal(err)
		}
	}
	answer, err := Do(TimesheetShow, path, given("since", "2026-09-29", "until", "2026-09-29"))
	if err != nil {
		t.Fatalf("timesheet show: %v", err)
	}
	if answer.Sheet == nil || len(answer.Sheet.Days) != 1 || answer.Sheet.Days[0].Total != 2*time.Hour {
		t.Fatalf("the sheet is %+v", answer.Sheet)
	}
	said, ok := Of(TimesheetShow, answer).(Timesheet)
	if !ok || said.Days[0].Day != "2026-09-29" || said.Days[0].Minutes != 120 ||
		said.Days[0].Tasks[0].Number != 1 || said.Days[0].Tasks[0].Title != "one" || said.Days[0].Tasks[0].Minutes != 120 {
		t.Errorf("a machine reads %+v", Of(TimesheetShow, answer))
	}
	if text := RenderText(TimesheetShow, answer); !strings.Contains(text, "#1  2h  one") || !strings.Contains(text, "total  2h") {
		t.Errorf("a person reads:\n%s", text)
	}
	history, err := Do(TraceShow, path, given("number", "1"))
	if err != nil {
		t.Fatalf("trace show: %v", err)
	}
	if text := RenderText(TraceShow, history); !strings.Contains(text, "30m for 2026-09-29") {
		t.Errorf("the history does not say what was added by hand:\n%s", text)
	}
}

// With nothing named, a timesheet is today's, and says today even when
// nothing was done on it.
func TestTimesheet_isTodaysWithNothingNamed(t *testing.T) {
	path := storeAt(t)
	if _, err := Do(TaskAdd, path, given("title", "one")); err != nil {
		t.Fatal(err)
	}
	// Today read on both sides of the call: midnight may fall between them,
	// and either day is the one the call meant.
	before := startOfToday()
	answer, err := Do(TimesheetShow, path, Input{})
	if err != nil {
		t.Fatal(err)
	}
	after := startOfToday()
	if answer.Sheet == nil || len(answer.Sheet.Days) != 1 ||
		!answer.Sheet.Days[0].Day.Equal(before) && !answer.Sheet.Days[0].Day.Equal(after) {
		t.Errorf("today's sheet is %+v", answer.Sheet)
	}
}

// Time added by hand is refused for a day to come, for a length nobody can
// read or that is no time at all, and for a task that is not there.
func TestTimesheet_refusesTimeThatCannotHaveBeenSpent(t *testing.T) {
	path := storeAt(t)
	if _, err := Do(TaskAdd, path, given("title", "one")); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		what string
		in   Input
		want error
	}{
		{"a day to come", given("number", "1", "time", "1h", "on", "2099-01-01"), ErrDayToCome},
		{"a length nobody can read", given("number", "1", "time", "long"), ErrNotATime},
		{"no time at all", given("number", "1", "time", "0m"), ErrNotATime},
		{"a task that is not there", given("number", "9", "time", "1h"), ErrNoSuchTask},
	} {
		if _, err := Do(TimesheetAdd, path, c.in); !errors.Is(err, c.want) {
			t.Errorf("%s: %v, want %v", c.what, err, c.want)
		}
	}
}

// Focus is taken and let go when it is done, and at no other moment: a focus
// that could be placed after the fact would be a timesheet anybody could
// write, and the time it measures would say what was typed rather than what
// was done.
func TestFocus_isTakenAndLetGoNowAndAtNoOtherMoment(t *testing.T) {
	for _, op := range []Operation{TaskFocus, TaskUnfocus} {
		if err := op.Check(given("number", "1", "at", "09:00")); err == nil {
			t.Errorf("%s takes a moment", op.Line())
		}
	}
}

// Time added answers the day it was added to, and too much is taken back the
// same way: the history keeps the length as it was given, minus and all.
func TestTimesheetAdd_answersItsDayAndTakesBack(t *testing.T) {
	path := storeAt(t)
	if _, err := Do(TaskAdd, path, given("title", "one")); err != nil {
		t.Fatal(err)
	}
	if _, err := Do(TimesheetAdd, path, given("number", "1", "time", "1h", "on", "2026-09-29")); err != nil {
		t.Fatal(err)
	}
	answer, err := Do(TimesheetAdd, path, given("number", "1", "time", "-30m", "on", "2026-09-29"))
	if err != nil {
		t.Fatal(err)
	}
	if answer.Sheet == nil || len(answer.Sheet.Days) != 1 || answer.Sheet.Days[0].Total != 30*time.Minute {
		t.Errorf("adding answers %+v, want the day at 30m", answer.Sheet)
	}
	history, err := Do(TraceShow, path, given("number", "1"))
	if err != nil {
		t.Fatal(err)
	}
	if text := RenderText(TraceShow, history); !strings.Contains(text, "-30m for 2026-09-29") {
		t.Errorf("the history does not keep what was taken back:\n%s", text)
	}
}
