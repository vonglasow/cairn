package engine

import (
	"cairn/internal/engine/calendar"
	"cairn/internal/engine/task"
	"encoding/json"
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

func TestTimesheet_measuresWhatWasDoneOnADayPast(t *testing.T) {
	path := storeAt(t)
	if _, err := Do(TaskNew, path, given("title", "one")); err != nil {
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
	history, err := Do(TraceShow, path, given("number", "1"))
	if err != nil {
		t.Fatalf("trace show: %v", err)
	}
	if said, _ := json.Marshal(Of(TraceShow, history)); !strings.Contains(string(said), "2026-09-29") || !strings.Contains(string(said), "30m") {
		t.Errorf("the history does not say what was added by hand: %s", said)
	}
}

func TestTimesheet_refusesTimeThatCannotHaveBeenSpent(t *testing.T) {
	path := storeAt(t)
	if _, err := Do(TaskNew, path, given("title", "one")); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		what string
		in   Input
		want error
	}{
		{"a day to come", given("number", "1", "time", "1h", "on", "2099-01-01"), calendar.ErrDayToCome},
		{"a length nobody can read", given("number", "1", "time", "long"), calendar.ErrNotATime},
		{"no time at all", given("number", "1", "time", "0m"), calendar.ErrNotATime},
		{"a task that is not there", given("number", "9", "time", "1h"), task.ErrNoSuchTask},
	} {
		if _, err := Do(TimesheetAdd, path, c.in); !errors.Is(err, c.want) {
			t.Errorf("%s: %v, want %v", c.what, err, c.want)
		}
	}
}

func confirmed(in Input) Input {
	in.Confirmed = true
	return in
}
