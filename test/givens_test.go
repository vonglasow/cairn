package test

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"cairn/internal/domain"
	"cairn/internal/engine"
	"cairn/internal/objective"
	"cairn/internal/task"
)

func givensOn(t *testing.T) Givens {
	t.Helper()
	e, err := engine.Open(filepath.Join(t.TempDir(), "cairn.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = e.Close() })
	return Givens{Other: func() Asker { return e }}
}

func shownText(t *testing.T, one task.Task) string {
	t.Helper()
	return task.Text(one, domain.Progress{})
}

func TestGivens_writeWhatTheyDescribe(t *testing.T) {
	g := givensOn(t)
	for _, step := range []func() error{
		func() error { return g.IAddedAnObjectiveTitled("Ship it") },
		func() error { return g.IAddedATaskTitled("one") },
		func() error { return g.IAddedATaskWithTheEstimate("two", "1h") },
		func() error { return g.IAddedATaskWithTheState("three", "current") },
		func() error { return g.IAddedATaskWithTheDescription("four", "about four") },
		func() error { return g.IAddedATaskWithTheParagraphs("five", "first", "second") },
		func() error { return g.IAddedATaskWithADescriptionOfLines("six", 3, "the end") },
		func() error { return g.IAddedATaskServingObjective("seven", 1) },
		func() error { return g.IAddedATaskWithTheStateServingObjective("eight", "watching", 1) },
		func() error { return g.IAddedATaskHangingUnderTask("nine", 4) },
		func() error { return g.IEditedTaskWithTheBlocker(2, "1") },
		func() error { return g.IEditedTaskWithTheState(1, "done") },
		func() error { return g.IEditedTaskWithTheLabels(3, "spec", "team") },
		func() error { return g.IEditedTaskWithTheDeadline(4, "2026-10-01") },
		func() error { return g.IPutTaskInFocus(3) },
		func() error { return g.IArchivedTask(1) },
		func() error { return g.IAddedTasksTo(2, "todo") },
		func() error { return g.IDeletedTask(9) },
	} {
		if err := step(); err != nil {
			t.Fatal(err)
		}
	}
	six, err := g.Task(6)
	if err != nil {
		t.Fatal(err)
	}
	if said := shownText(t, six); !strings.Contains(said, "line 1 of a description long enough to scroll\nline 2 of") ||
		strings.Contains(said, "line 3 of") || strings.Contains(said, "line 0") {
		t.Errorf("a description of three lines reads:\n%s", said)
	}
	all, err := g.Tasks()
	if err != nil {
		t.Fatal(err)
	}
	var titles []string
	for _, one := range all {
		f := one.Carried()
		titles = append(titles, f.Title)
	}
	if !slices.Contains(titles, "Task 1") || !slices.Contains(titles, "Task 2") || slices.Contains(titles, "Task 3") || slices.Contains(titles, "Task 0") {
		t.Errorf("two tasks added to todo are %v", titles)
	}
	for number, want := range map[int][]string{
		1: {"archived", "done"},
		2: {"1h", "#1"},
		3: {"current", "spec", "team"},
		4: {"about four", "2026-10-01"},
		5: {"first\nsecond"},
		6: {"the end"},
		8: {"watching"},
	} {
		shown, err := g.Task(number)
		if err != nil {
			t.Fatal(err)
		}
		for _, w := range want {
			if text := shownText(t, shown); !strings.Contains(text, w) {
				t.Errorf("task %d does not say %q:\n%s", number, w, text)
			}
		}
	}
	if len(all) == 0 {
		t.Error("nothing listed")
	}
	if _, err := g.Task(9); err == nil {
		t.Error("task 9 was deleted and is still shown")
	}
	goal, err := g.Objective(1)
	if err != nil {
		t.Fatal(err)
	}
	if text := objective.Text(goal, domain.Progress{}); !strings.Contains(text, "Ship it") {
		t.Errorf("the objective reads %q", text)
	}
	if _, err := g.Objective(9); err == nil {
		t.Error("an objective nobody added was shown")
	}
}
