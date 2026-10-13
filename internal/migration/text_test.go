package migration

import (
	"strings"
	"testing"
	"time"

	"cairn/internal/domain"
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

func TestText_showsWhatAChangeDoesAndWhatUndoingItDoes(t *testing.T) {
	got := Text(builtChange(false))
	for _, want := range []string{
		"#2", "waiting", "a task can hang under another",
		"what it does", "ALTER TABLE tasks ADD COLUMN parent",
		"what undoing it does", "ALTER TABLE tasks DROP COLUMN parent",
		"what undoing it loses", "which task each sub-task hangs under",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("Text() = %q, want it to contain %q", got, want)
		}
	}
}

func TestText_saysWhenAChangeWasApplied(t *testing.T) {
	got := Text(builtChange(true))
	if want := "applied " + domain.FormatMoment(time.Date(2026, 1, 2, 15, 4, 0, 0, time.UTC)); !strings.Contains(got, want) {
		t.Errorf("Text() = %q, want it to say when the change was applied", got)
	}
}

func TestList_countsWhatIsWaiting(t *testing.T) {
	got := List(
		[]Migration{builtChange(true), builtChange(true), builtChange(false)},
		domain.Held{Tasks: 2, Objectives: 1},
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

func TestRolledBackText_saysNothingLostWhenNothingWas(t *testing.T) {
	got := RolledBackText(Made(Carried{Name: "the shape a store starts with", Does: "CREATE TABLE tasks (...)"}))
	if strings.Contains(got, "lost:") {
		t.Errorf("RolledBackText() = %q, want no mention of loss", got)
	}
}
