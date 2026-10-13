package migration

import (
	"fmt"
	"testing"
	"time"
)

type told struct {
	said []string
}

func (r *told) Numbered(n int)      { r.said = append(r.said, fmt.Sprintf("Numbered(%d)", n)) }
func (r *told) Named(s string)      { r.said = append(r.said, fmt.Sprintf("Named(%q)", s)) }
func (r *told) Does(s string)       { r.said = append(r.said, fmt.Sprintf("Does(%q)", s)) }
func (r *told) Undoes(s string)     { r.said = append(r.said, fmt.Sprintf("Undoes(%q)", s)) }
func (r *told) NotUndoable()        { r.said = append(r.said, "NotUndoable()") }
func (r *told) Loses(s string)      { r.said = append(r.said, fmt.Sprintf("Loses(%q)", s)) }
func (r *told) NothingLost()        { r.said = append(r.said, "NothingLost()") }
func (r *told) Applied(t time.Time) { r.said = append(r.said, fmt.Sprintf("Applied(%s)", t)) }
func (r *told) NotApplied()         { r.said = append(r.said, "NotApplied()") }

func TestTell_saysEveryFactAChangeCarries(t *testing.T) {
	made := time.Date(2026, 1, 2, 15, 4, 0, 0, time.UTC)
	full := Migration{
		number: 2, name: "a task can hang under another",
		does: "ALTER TABLE tasks ADD COLUMN parent", undoes: "ALTER TABLE tasks DROP COLUMN parent",
		loses: "which task each sub-task hangs under", applied: true, made: made,
	}
	got := &told{}
	full.Tell(got)

	want := []string{
		`Numbered(2)`, `Named("a task can hang under another")`,
		`Does("ALTER TABLE tasks ADD COLUMN parent")`,
		`Undoes("ALTER TABLE tasks DROP COLUMN parent")`,
		`Loses("which task each sub-task hangs under")`,
		fmt.Sprintf("Applied(%s)", made),
	}
	if len(got.said) != len(want) {
		t.Fatalf("said = %v, want %v", got.said, want)
	}
	for i := range want {
		if got.said[i] != want[i] {
			t.Errorf("said[%d] = %q, want %q", i, got.said[i], want[i])
		}
	}

	floor := Migration{number: 1, name: "the shape a store starts with", does: "CREATE TABLE tasks (...)"}
	got = &told{}
	floor.Tell(got)

	wantFloor := []string{
		`Numbered(1)`, `Named("the shape a store starts with")`,
		`Does("CREATE TABLE tasks (...)")`,
		`NotUndoable()`, `NothingLost()`, `NotApplied()`,
	}
	if len(got.said) != len(wantFloor) {
		t.Fatalf("said = %v, want %v", got.said, wantFloor)
	}
	for i := range wantFloor {
		if got.said[i] != wantFloor[i] {
			t.Errorf("said[%d] = %q, want %q", i, got.said[i], wantFloor[i])
		}
	}
}
