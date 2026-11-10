package engine

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"cairn/internal/domain"
)

func TestTaskExport_writesTheSeriesTheirTemplateAndTheirTrace(t *testing.T) {
	read := exportedFrom(t,
		asked(ObjectiveNew, "title", "Served by the series alone"),
		asked(TaskNew, "title", "Stand-up", "every", "every weekday 10:30", "objective", "1", "label", "work", "estimate", "15m"),
		asked(TaskEdit, "number", "1", "objective", "-1"),
		asked(TaskEdit, "number", "1", "state", "done"),
		asked(TaskEdit, "number", "2", "objective", "-1"),
		asked(TaskNew, "title", "Expense report", "every", "every 4 weeks on monday"),
		asked(SeriesStop, "number", "2"),
	)
	if read.Version != 6 {
		t.Errorf("version %d, want 6", read.Version)
	}
	if len(read.Series) != 2 {
		t.Fatalf("the file holds %d series, want 2", len(read.Series))
	}
	first, second := read.Series[0], read.Series[1]
	if first.Number != 1 || first.Rule != "every weekday 10:30" || first.Stopped != "" || first.From == "" {
		t.Errorf("the first series reads %+v", first)
	}
	if first.Template.Title != "Stand-up" || first.Template.Estimate != "15m" || !slices.Equal(first.Template.Labels, []string{"work"}) || !slices.Equal(first.Template.Objectives, []int{1}) {
		t.Errorf("its template reads %+v", first.Template)
	}
	if len(first.Acts) != 2 || first.Acts[0].Did != "made" || first.Acts[0].Task != 1 || first.Acts[1].Task != 2 || first.Acts[0].At == "" {
		t.Errorf("its acts read %+v", first.Acts)
	}
	if second.Stopped == "" || !slices.ContainsFunc(second.Acts, func(a SeriesAct) bool { return a.Did == "stopped" }) {
		t.Errorf("the second series reads %+v", second)
	}
	if len(read.Objectives) != 1 || read.Objectives[0].Number != 1 {
		t.Errorf("the objectives written are %+v, want the one the series serves", read.Objectives)
	}
}

func TestTaskImport_bringsTheSeriesBackLinkedToTheirTasks(t *testing.T) {
	e := imported(t, `{"cairn": "tasks", "version": 6,
		"objectives": [{"number": 4, "title": "Team", "state": "active", "priority": "normal"}],
		"labels": [{"label": "work"}],
		"tasks": [
		  {"number": 7, "title": "Stand-up", "state": "done", "priority": "normal"},
		  {"number": 8, "title": "Stand-up", "state": "todo", "priority": "normal", "deadline": "+0d"}],
		"series": [{"number": 3, "rule": "every day", "from": "-5d",
		  "template": {"title": "Stand-up", "priority": "high", "labels": ["daily"], "steps": ["Yesterday"], "estimate": "15m", "objectives": [4]},
		  "acts": [
		    {"did": "made", "task": 99, "day": "-3d", "through": "cli", "by": "me", "at": "2026-11-01T09:00:00Z"},
		    {"did": "made", "task": 7, "day": "-1d"},
		    {"did": "missed", "day": "-2d"},
		    {"did": "made", "task": 8, "day": "+0d"}]},
		  {"number": 5, "rule": "every monday", "from": "2026-01-05", "stopped": "2026-10-12",
		  "template": {"title": "Report", "priority": "normal"}}]}`)
	of, err := e.store.Occurrences()
	if err != nil {
		t.Fatal(err)
	}
	if of[1] != 1 || of[2] != 1 {
		t.Errorf("the occurrences are %v, want tasks 1 and 2 of ↻1", of)
	}
	s, err := e.seriesAt(1)
	if err != nil {
		t.Fatal(err)
	}
	today := time.Now()
	from, _ := domain.FromToday("-5d", today)
	if s.kept.Rule != "every day" || s.kept.Start != from || s.kept.Title != "Stand-up" || s.kept.Priority != "high" || s.kept.Estimate != 15*time.Minute ||
		!slices.Equal(s.kept.Steps, []string{"Yesterday"}) || !slices.Equal(s.kept.Labels, []string{"daily"}) || !slices.Equal(s.kept.Objectives, []int{1}) {
		t.Errorf("↻1 came back as %+v", s.kept)
	}
	if len(s.acts) != 4 || s.acts[0].Task != 0 || s.acts[0].By != "me" || !s.acts[0].At.Equal(time.Date(2026, 11, 1, 9, 0, 0, 0, time.UTC)) || s.acts[1].Task != 1 || s.acts[3].Task != 2 {
		t.Errorf("its acts came back as %+v", s.acts)
	}
	if day, _ := domain.FromToday("-2d", today); s.acts[2].Did != "missed" || s.acts[2].Day != day {
		t.Errorf("the day missed came back as %+v", s.acts[2])
	}
	stopped, err := e.seriesAt(2)
	if err != nil || stopped.kept.Stopped != "2026-10-12" || stopped.active() {
		t.Errorf("↻2 came back as %+v (%v)", stopped.kept, err)
	}
	labels, _ := e.everyLabel()
	if _, kept := labels["daily"]; !kept {
		t.Errorf("the label of the template was not kept: %v", labels)
	}
}

func TestTaskImport_refusesASeriesItCannotRead(t *testing.T) {
	for _, series := range []string{
		`{"number": 1, "rule": "every day", "from": "2026-11-09", "template": {"title": "x", "priority": "normal", "objectives": [9]}}`,
		`{"number": 1, "rule": "every blue moon", "from": "2026-11-09", "template": {"title": "x", "priority": "normal"}}`,
		`{"number": 1, "rule": "every day", "from": "someday", "template": {"title": "x", "priority": "normal"}}`,
		`{"number": 1, "rule": "every day", "from": "2026-11-09", "stopped": "someday", "template": {"title": "x", "priority": "normal"}}`,
		`{"number": 1, "rule": "every day", "from": "2026-11-09", "template": {"title": "x", "priority": "urgent"}}`,
		`{"number": 1, "rule": "every day", "from": "2026-11-09", "template": {"title": "", "priority": "normal"}}`,
		`{"number": 1, "rule": "every day", "from": "2026-11-09", "template": {"title": "x", "priority": "normal", "estimate": "long"}}`,
		`{"number": 1, "rule": "every day", "from": "2026-11-09", "template": {"title": "x", "priority": "normal", "labels": ["a\nb"]}}`,
		`{"number": 1, "rule": "every day", "from": "2026-11-09", "template": {"title": "x", "priority": "normal"}, "acts": [{"did": "dreamt"}]}`,
		`{"number": 1, "rule": "every day", "from": "2026-11-09", "template": {"title": "x", "priority": "normal"}, "acts": [{"did": "made", "task": 1, "day": "someday"}]}`,
		`{"number": 1, "rule": "every day", "from": "2026-11-09", "template": {"title": "x", "priority": "normal"}, "acts": [{"did": "made", "task": 1, "at": "noon"}]}`,
	} {
		from := filepath.Join(t.TempDir(), "held.json")
		file := `{"cairn": "tasks", "version": 6, "objectives": [], "tasks": [{"number": 1, "title": "Ship it", "state": "todo", "priority": "normal"}], "series": [` + series + `]}`
		if err := os.WriteFile(from, []byte(file), 0o600); err != nil {
			t.Fatal(err)
		}
		e := openEngine(t)
		read, err := readExport(from)
		if err == nil {
			_, err = e.importing(read)
		}
		if err == nil {
			t.Errorf("%s was read", series)
			continue
		}
		if tasks, _ := e.everyTask(); len(tasks) != 0 {
			t.Errorf("%s added %d tasks before it was refused", series, len(tasks))
		}
	}
	from := filepath.Join(t.TempDir(), "twice.json")
	if err := os.WriteFile(from, []byte(`{"cairn": "tasks", "version": 6, "objectives": [], "tasks": [], "series": [
		{"number": 1, "rule": "every day", "from": "2026-11-09", "template": {"title": "x", "priority": "normal"}},
		{"number": 1, "rule": "every day", "from": "2026-11-09", "template": {"title": "y", "priority": "normal"}}]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	read, err := readExport(from)
	if err == nil {
		_, err = openEngine(t).importing(read)
	}
	if !errors.Is(err, ErrNotAnExport) {
		t.Errorf("a series in the file twice reads %v", err)
	}
}

func TestTaskImport_aFileOfVersion6IsRead(t *testing.T) {
	from := filepath.Join(t.TempDir(), "held.json")
	if err := os.WriteFile(from, []byte(`{"cairn": "tasks", "version": 7, "objectives": [], "tasks": []}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := readExport(from); !errors.Is(err, ErrNotAnExport) {
		t.Errorf("a file of version 7 reads %v", err)
	}
}
