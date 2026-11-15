package engine

import (
	"path/filepath"
	"strings"
	"testing"
)

func choosingStore(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "cairn.db")
	for _, title := range []string{"Write the spec", "Ship it", "Announce it", "Thank the reviewers"} {
		did(t, path, TaskNew, ask("title", title))
	}
	return path
}

func listedOn(t *testing.T, path, day string, a request) []map[string]any {
	t.Helper()
	return rowsAsJSON(t, TaskList, doneOn(t, path, day, TaskList, a))
}

func TestTaskList_keepsTheTasksOneIsInTheWayOf(t *testing.T) {
	path := choosingStore(t)
	did(t, path, TaskEdit, ask("number", "2").with("blocker", "1"))
	did(t, path, TaskEdit, ask("number", "4").with("blocker", "3", "1"))
	did(t, path, TaskEdit, ask("number", "3").with("blocker", "2"))
	rows := listedOn(t, path, "2026-11-13", ask("blocker", "1"))
	if len(rows) != 2 || rows[0]["number"] != float64(2) || rows[1]["number"] != float64(4) {
		t.Errorf("listed %v, want tasks 2 and 4", rows)
	}
}

func TestTaskList_aBlockerThatIsNoTaskIsRefused(t *testing.T) {
	path := choosingStore(t)
	for _, said := range []string{"x", "0", "-2"} {
		if _, err := Do(TaskList, path, ask("blocker", said).input()); err == nil || !strings.Contains(err.Error(), said) {
			t.Errorf("--blocker %s: %v, want it refused naming it", said, err)
		}
	}
}

func TestTaskShow_namesTheTasksItIsInTheWayOf(t *testing.T) {
	path := choosingStore(t)
	did(t, path, TaskEdit, ask("number", "3").with("blocker", "1"))
	did(t, path, TaskEdit, ask("number", "2").with("blocker", "1"))
	answer := did(t, path, TaskShow, ask("number", "1"))
	if said := RenderText(TaskShow, answer); !strings.Contains(said, "blocks       ⊳2  #2, #3\n") {
		t.Errorf("task show said:\n%s\nwant it to name tasks 2 and 3", said)
	}
	if said := RenderText(TaskShow, did(t, path, TaskShow, ask("number", "2"))); strings.Contains(said, "⊳") {
		t.Errorf("task show said:\n%s\nwant nothing of what it blocks", said)
	}
}
