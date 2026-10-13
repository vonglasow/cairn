package migration

import (
	"testing"
	"time"
)

func TestAsTold_buildsWhatItWasTold(t *testing.T) {
	made := time.Date(2026, 1, 2, 15, 4, 0, 0, time.UTC)
	full := Migration{
		number: 2, name: "a task can hang under another",
		does: "ALTER TABLE tasks ADD COLUMN parent", undoes: "ALTER TABLE tasks DROP COLUMN parent",
		loses: "which task each sub-task hangs under", applied: true, made: made,
	}
	var a AsTold
	full.Tell(&a)
	got := a.Migration()
	if got != full {
		t.Errorf("Migration() = %+v, want %+v", got, full)
	}

	floor := Migration{number: 1, name: "the shape a store starts with", does: "CREATE TABLE tasks (...)"}
	var b AsTold
	floor.Tell(&b)
	if got := b.Migration(); got != floor {
		t.Errorf("Migration() = %+v, want %+v", got, floor)
	}
}

func TestAsTold_clearsWhatItIsToldIsAbsent(t *testing.T) {
	var a AsTold
	a.Undoes("DROP COLUMN parent")
	a.NotUndoable()
	a.Loses("which task each sub-task hangs under")
	a.NothingLost()
	a.Applied(time.Now())
	a.NotApplied()

	got := a.Migration()
	if got.undoes != "" {
		t.Errorf("undoes = %q, want cleared", got.undoes)
	}
	if got.loses != "" {
		t.Errorf("loses = %q, want cleared", got.loses)
	}
	if got.applied {
		t.Errorf("applied = %v, want cleared", got.applied)
	}
}
