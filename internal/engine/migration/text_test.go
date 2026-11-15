package migration

import (
	"cairn/internal/engine/backup"
	"strings"
	"testing"
	"time"
)

func builtChange(applied bool) Migration {
	c := Carried{
		Number: 2, Name: "a task can hang under another",
		Does:   "ALTER TABLE tasks ADD COLUMN parent INTEGER NOT NULL DEFAULT 0",
		Undoes: "ALTER TABLE tasks DROP COLUMN parent", Loses: "which task each sub-task hangs under",
	}
	if applied {
		c.Applied, c.Made = true, time.Date(2026, 1, 2, 15, 4, 0, 0, time.UTC)
	}
	return Made(c)
}

func TestList_countsWhatIsWaiting(t *testing.T) {
	got := List(
		[]Migration{builtChange(true), builtChange(true), builtChange(false)},
		backup.Held{Tasks: 2, Objectives: 1},
	)
	if !strings.Contains(got, "holds 2 task(s) and 1 objective(s), 1 change(s) waiting\n") {
		t.Errorf("List() = %q, want it to count exactly the one still waiting", got)
	}
	if !strings.Contains(got, "holds 2 task(s) and 1 objective(s)") {
		t.Errorf("List() = %q, want it to say what the store holds", got)
	}
}

func TestRolledBackText_namesTheChangeAndWhatItCost(t *testing.T) {
	got := RolledBackText(builtChange(true))
	if !strings.Contains(got, `rolled back "a task can hang under another"`) {
		t.Errorf("RolledBackText() = %q, want it to name the change", got)
	}
	if !strings.Contains(got, "lost: which task each sub-task hangs under") {
		t.Errorf("RolledBackText() = %q, want it to say what was lost", got)
	}
}
