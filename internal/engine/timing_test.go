package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"cairn/internal/trace"
)

func homeTimed(t *testing.T, path, said string, at time.Time) Home {
	t.Helper()
	home := t.TempDir()
	dir := filepath.Join(home, ".config", "cairn")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(said), 0o600); err != nil {
		t.Fatal(err)
	}
	now := at.Format(time.RFC3339)
	return Home{Dir: home, Lookup: func(k string) (string, bool) {
		switch k {
		case "CAIRN_STORE":
			return path, true
		case NowVariable:
			return now, true
		}
		return "", false
	}}
}

func timeOff(t *testing.T, path string, at time.Time) Home {
	t.Helper()
	return homeTimed(t, path, "[pages]\ntime = \"off\"\n", at)
}

func timeRecorded(t *testing.T, path string, at time.Time) Home {
	t.Helper()
	return homeTimed(t, path, "[pages]\ntime = \"hidden\"\n", at)
}

func actsOf(t *testing.T, path string, number int) []trace.Carried {
	t.Helper()
	e, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = e.Close() }()
	acts, err := e.history.ActsOn(number)
	if err != nil {
		t.Fatal(err)
	}
	out := make([]trace.Carried, 0, len(acts))
	for _, a := range acts {
		out = append(out, a.Carried())
	}
	return out
}

func timerActs(t *testing.T, path string) []string {
	t.Helper()
	var said []string
	for _, c := range actsOf(t, path, 1) {
		switch c.Did {
		case trace.FocusTaken, trace.FocusReleased, trace.TimerStarted, trace.TimerStopped:
			said = append(said, c.Did+" "+c.Became)
		}
	}
	return said
}

func ranAt(t *testing.T, op Operation, h Home, in Input) Answer {
	t.Helper()
	answer, err := DoFrom(op, h, in, Acting{})
	if err != nil {
		t.Fatal(err)
	}
	return answer
}

func TestFocus_takenWhileTheTimeIsOffHasNoTimer(t *testing.T) {
	path := storeAt(t, "one")
	nine := time.Date(2026, 9, 29, 9, 0, 0, 0, time.Local)
	ranAt(t, TaskFocus, timeOff(t, path, nine), given("number", "1"))
	if got := timerActs(t, path); len(got) != 1 || got[0] != trace.FocusTaken+" "+trace.WithoutATimer {
		t.Errorf("focus taken with the time off wrote %q, want one focus taken without a timer", got)
	}
}

func TestTiming_theTimerStopsWhenTheTimeIsOffAndStartsWhenItIsOn(t *testing.T) {
	path := storeAt(t, "one")
	nine := time.Date(2026, 9, 29, 9, 0, 0, 0, time.Local)
	ranAt(t, TaskFocus, timeRecorded(t, path, nine), given("number", "1"))
	ranAt(t, TaskList, timeRecorded(t, path, nine.Add(10*time.Minute)), given())
	ranAt(t, TaskList, timeOff(t, path, nine.Add(time.Hour)), given())
	ranAt(t, TaskList, timeOff(t, path, nine.Add(90*time.Minute)), given())
	ranAt(t, TaskList, timeRecorded(t, path, nine.Add(2*time.Hour)), given())
	want := []string{trace.FocusTaken + " ", trace.TimerStopped + " ", trace.TimerStarted + " "}
	got := timerActs(t, path)
	if len(got) != len(want) {
		t.Fatalf("the timer acts are %q, want %q", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("the timer acts are %q, want %q", got, want)
		}
	}
	acts := actsOf(t, path, 1)
	if stopped := acts[len(acts)-2]; !stopped.At.Equal(nine.Add(time.Hour)) {
		t.Errorf("the timer stopped at %s, want when cairn first acted with the time off", stopped.At)
	}
}

func TestTiming_aTaskNotInFocusHasNoTimerToStop(t *testing.T) {
	path := storeAt(t, "one")
	nine := time.Date(2026, 9, 29, 9, 0, 0, 0, time.Local)
	ranAt(t, TaskFocus, timeRecorded(t, path, nine), given("number", "1"))
	ranAt(t, TaskUnfocus, timeRecorded(t, path, nine.Add(time.Minute)), given("number", "1"))
	ranAt(t, TaskList, timeOff(t, path, nine.Add(time.Hour)), given())
	if got := timerActs(t, path); len(got) != 2 {
		t.Errorf("the timer acts are %q, want the focus taken and released alone", got)
	}
}

func TestAnswer_asksNoRenewalWhileTheTimeIsOff(t *testing.T) {
	path := storeAt(t, "one")
	nine := time.Date(2026, 9, 29, 9, 0, 0, 0, time.Local)
	said := "[pages]\ntime = \"off\"\n[focus]\nsession = \"1h\"\n"
	ranAt(t, TaskFocus, homeTimed(t, path, said, nine), given("number", "1"))
	if answer := ranAt(t, TaskList, homeTimed(t, path, said, nine.Add(2*time.Hour)), given()); len(answer.Asked) != 0 {
		t.Errorf("with the time off it asks %q", answer.Asked)
	}
	ranAt(t, TaskFocus, homeTimed(t, path, said, nine.Add(3*time.Hour)), given("number", "1"))
	for _, c := range actsOf(t, path, 1) {
		if c.Did == trace.FocusRenewed {
			t.Errorf("with the time off, focusing again renewed the session")
		}
	}
}

func TestAnswer_saysNoTimeSpentWhileTheTimeIsNotShown(t *testing.T) {
	for _, mode := range []string{"hidden", "off"} {
		t.Run(mode, func(t *testing.T) {
			path := storeAt(t, "one")
			nine := time.Date(2026, 9, 29, 9, 0, 0, 0, time.Local)
			on := homeTimed(t, path, "", nine)
			ranAt(t, ObjectiveNew, on, given("title", "ship"))
			ranAt(t, TaskEdit, on, Input{Given: map[string]string{"number": "1", "spent": "2h"}, Listed: map[string][]string{"objective": {"=1"}}})
			not := homeTimed(t, path, "[pages]\ntime = \""+mode+"\"\n", nine)
			shown := ranAt(t, TaskShow, not, given("number", "1"))
			if spent := factsOf(*shown.Task).Spent; spent != 0 {
				t.Errorf("task show says %s spent", spent)
			}
			if said := RenderText(TaskShow, shown); strings.Contains(said, "time spent") {
				t.Errorf("task show says the time spent:\n%s", said)
			}
			objective := ranAt(t, ObjectiveShow, not, given("number", "1"))
			if said := RenderText(ObjectiveShow, objective); strings.Contains(said, "time spent") {
				t.Errorf("objective show says the time spent:\n%s", said)
			}
			if listed := ranAt(t, ObjectiveList, not, given()); len(listed.Spent) != 0 {
				t.Errorf("objective list says %v spent", listed.Spent)
			}
			if spent := factsOf(*ranAt(t, TaskShow, on, given("number", "1")).Task).Spent; spent != 2*time.Hour {
				t.Errorf("shown again, the time spent is %s, want 2h", spent)
			}
		})
	}
}
