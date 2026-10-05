package migration

import (
	"strings"
	"testing"
	"time"

	"cairn/internal/domain"
)

func builtChange(applied bool) Migration {
	var a AsTold
	a.Numbered(2)
	a.Named("a task can hang under another")
	a.Does("ALTER TABLE tasks ADD COLUMN parent INTEGER NOT NULL DEFAULT 0")
	a.Undoes("ALTER TABLE tasks DROP COLUMN parent")
	a.Loses("which task each sub-task hangs under")
	if applied {
		a.Applied(time.Date(2026, 1, 2, 15, 4, 0, 0, time.UTC))
	} else {
		a.NotApplied()
	}
	return a.Migration()
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
	// The moment as this machine's zone writes it: 15:04 UTC on the 2nd is
	// already the 3rd east of UTC+9.
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
	var a AsTold
	a.Named("the shape a store starts with")
	a.Does("CREATE TABLE tasks (...)")
	a.NotUndoable()
	a.NothingLost()
	a.NotApplied()
	got := RolledBackText(a.Migration())
	if strings.Contains(got, "lost:") {
		t.Errorf("RolledBackText() = %q, want no mention of loss", got)
	}
}
