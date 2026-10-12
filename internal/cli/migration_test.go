package cli

import (
	"strings"
	"testing"
)

// Every command that touches the store has to answer the same way when there
// is no store to touch — including the backup commands, which are the ones
// somebody reaches for precisely when things have gone wrong.
func TestMigrationRollback_saysWhatItRolledBack(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("CAIRN_STORE", storeFor(t))
	mustRun(t, "task", "new", "Write the spec")

	out := mustRun(t, "migration", "rollback", "--yes").Stdout
	if !strings.Contains(out, "rolled back") {
		t.Errorf("stdout = %q, want it to name what it rolled back", out)
	}
}

// Putting a backup back replaces the store; saying which one was put back is how
// somebody knows it was the one they meant. A silent success looks exactly like
// a command that did nothing.
// Counted, not matched. A step reading this line with strings.Contains would
// accept "-1 change(s) waiting" as saying "1 change(s) waiting", so the count
// is asserted on its own here.
func TestMigrationList_countsWhatIsWaitingExactly(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("CAIRN_STORE", storeFor(t))
	mustRun(t, "task", "new", "Write the spec")

	if out := mustRun(t, "migration", "list").Stdout; !strings.Contains(out, "0 change(s) waiting") {
		t.Errorf("stdout = %q, want nothing waiting on a store made today", out)
	}

	mustRun(t, "migration", "rollback", "--yes")
	out := mustRun(t, "migration", "list").Stdout
	if !strings.Contains(out, "1 change(s) waiting") {
		t.Errorf("stdout = %q, want one change waiting", out)
	}
	if strings.Contains(out, "-1 change(s) waiting") {
		t.Errorf("stdout = %q, the count went backwards", out)
	}
}

// The first change is a change like any other: `show 1` must work, or the
// boundary that rejects a number below one is off.
func TestMigrationShow_showsTheFirstChange(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("CAIRN_STORE", storeFor(t))
	mustRun(t, "task", "new", "Write the spec")

	out := mustRun(t, "migration", "show", "1").Stdout
	if !strings.Contains(out, "the shape a store starts with") {
		t.Errorf("stdout = %q, want the first change", out)
	}
	if got := run(t, "migration", "show", "0").Code; got == exitOK {
		t.Error("`migration show 0` should fail: there is no change zero")
	}
}
