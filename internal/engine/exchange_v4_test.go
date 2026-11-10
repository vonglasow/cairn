package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"cairn/internal/domain"
)

func exportedFrom(t *testing.T, asks ...func() (Operation, Input)) exchanged {
	t.Helper()
	path := storeAt(t)
	ran(t, path, asks...)
	to := filepath.Join(t.TempDir(), "tasks.json")
	if _, err := Do(TaskExport, path, Input{Given: map[string]string{"to": to}}); err != nil {
		t.Fatal(err)
	}
	written, err := os.ReadFile(to)
	if err != nil {
		t.Fatal(err)
	}
	var read exchanged
	if err := json.Unmarshal(written, &read); err != nil {
		t.Fatal(err)
	}
	return read
}

func TestTaskExport_writesTheTimeDayByDayTheFocusAndTheLabels(t *testing.T) {
	read := exportedFrom(t,
		asked(TaskNew, "title", "Ship it"),
		asked(TaskNew, "title", "Review"),
		asked(TimesheetAdd, "number", "1", "time", "2h", "on", "2026-10-01"),
		asked(TimesheetAdd, "number", "1", "time", "45m", "on", "2026-10-02"),
		asked(TaskFocus, "number", "2"),
		func() (Operation, Input) { return LabelNew, given("label", "release") },
		func() (Operation, Input) { return LabelColour, given("label", "release", "colour", "teal") },
	)
	if read.Version != 6 {
		t.Errorf("version %d, want 6", read.Version)
	}
	if got := read.Tasks[0].Spent; got["2026-10-01"] != "2h" || got["2026-10-02"] != "45m" || len(got) != 2 {
		t.Errorf("task 1 spent %v, want 2h on the 1st and 45m on the 2nd", got)
	}
	if read.Tasks[0].Focused || !read.Tasks[1].Focused {
		t.Errorf("focused = %v, %v; want task 2 alone", read.Tasks[0].Focused, read.Tasks[1].Focused)
	}
	if !slices.Contains(read.Labels, labelFiled{Label: "release", Colour: "teal"}) {
		t.Errorf("labels = %v, want release in teal", read.Labels)
	}
}

func imported(t *testing.T, file string) *Engine {
	t.Helper()
	from := filepath.Join(t.TempDir(), "held.json")
	if err := os.WriteFile(from, []byte(file), 0o600); err != nil {
		t.Fatal(err)
	}
	e := openEngine(t)
	read, err := readExport(from)
	if err == nil {
		_, err = e.importing(read)
	}
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	return e
}

func TestTaskImport_readsWhatVersion4Adds(t *testing.T) {
	e := imported(t, `{"cairn": "tasks", "version": 4, "objectives": [{"number": 1, "title": "Release", "state": "active", "priority": "normal", "deadline": "+1w"}],
		"labels": [{"label": "release", "colour": "teal"}, {"label": "idle"}],
		"tasks": [{"number": 1, "title": "Ship it", "state": "todo", "priority": "normal", "deadline": "+3d", "objectives": [1],
		  "time_spent": "9h", "spent": {"-1d": "2h", "2026-10-01": "45m"}, "focused": true}]}`)
	today := time.Now()
	got, _ := e.show(1)
	f := factsOf(got)
	if domain.FormatDay(f.Deadline) != domain.FormatDay(today.AddDate(0, 0, 3)) {
		t.Errorf("deadline %v, want three days from today", f.Deadline)
	}
	if f.Spent < 2*time.Hour+45*time.Minute || f.Spent > 2*time.Hour+46*time.Minute {
		t.Errorf("spent %v, want 2h45m and the minute begun in focus: the days, not the total", f.Spent)
	}
	if f.State != domain.Current {
		t.Errorf("state %s: a task in focus is under way", f.State)
	}
	if held, _ := e.focus(); !slices.Equal(held, []int{1}) {
		t.Errorf("in focus %v, want task 1", held)
	}
	o, _ := e.showObjective(1)
	if domain.FormatDay(factsOfObjective(o).Deadline) != domain.FormatDay(today.AddDate(0, 0, 7)) {
		t.Errorf("objective deadline %v, want a week from today", factsOfObjective(o).Deadline)
	}
	labels, _ := e.labelsCarried()
	if !slices.Contains(labels, LabelCarried{Label: "release", Colour: "teal", Scope: "both"}) || !slices.ContainsFunc(labels, func(l LabelCarried) bool { return l.Label == "idle" }) {
		t.Errorf("labels %v, want release in teal and idle", labels)
	}
}

func TestTaskImport_bringsTheTotalOfAVersion3File(t *testing.T) {
	e := imported(t, `{"cairn": "tasks", "version": 3, "objectives": [], "tasks": [{"number": 1, "title": "Ship it", "state": "todo", "priority": "normal", "time_spent": "1h30m"}]}`)
	if got, _ := e.show(1); factsOf(got).Spent != 90*time.Minute {
		t.Errorf("spent %v, want 1h30m", factsOf(got).Spent)
	}
}

func TestTaskImport_refusesWhatVersion4CannotRead(t *testing.T) {
	for _, file := range []string{
		`{"cairn": "tasks", "version": 4, "objectives": [], "tasks": [{"number": 1, "title": "x", "state": "todo", "priority": "normal", "deadline": "+3x"}]}`,
		`{"cairn": "tasks", "version": 4, "objectives": [{"number": 1, "title": "x", "state": "active", "priority": "normal", "deadline": "soon"}], "tasks": []}`,
		`{"cairn": "tasks", "version": 4, "objectives": [], "tasks": [{"number": 1, "title": "x", "state": "todo", "priority": "normal", "spent": {"-1q": "1h"}}]}`,
		`{"cairn": "tasks", "version": 4, "objectives": [], "tasks": [{"number": 1, "title": "x", "state": "todo", "priority": "normal", "spent": {"-1d": "ages"}}]}`,
		`{"cairn": "tasks", "version": 4, "objectives": [], "labels": [{"label": "x", "colour": "plaid"}], "tasks": []}`,
		`{"cairn": "tasks", "version": 4, "objectives": [], "labels": [{"label": "a:b"}], "tasks": []}`,
		`{"cairn": "tasks", "version": 7, "objectives": [], "tasks": []}`,
	} {
		from := filepath.Join(t.TempDir(), "held.json")
		if err := os.WriteFile(from, []byte(file), 0o600); err != nil {
			t.Fatal(err)
		}
		e := openEngine(t)
		read, err := readExport(from)
		if err == nil {
			_, err = e.importing(read)
		}
		if err == nil {
			t.Errorf("%s was read in", file)
		}
		if tasks, _ := e.narrowed(&domain.Narrowing{}); len(tasks) != 0 {
			t.Errorf("%s added %d tasks before it was refused", file, len(tasks))
		}
		if labels, _ := e.labelsCarried(); len(labels) != 0 {
			t.Errorf("%s added labels %v before it was refused", file, labels)
		}
	}
}
