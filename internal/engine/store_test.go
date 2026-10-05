package engine

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// storeAt gives a path to a real store, with n tasks in it, closed.
func storeAt(t *testing.T, titles ...string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "cairn.db")
	e, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	for _, title := range titles {
		if _, err := e.add(title, Changes{}); err != nil {
			t.Fatalf("Add: %v", err)
		}
	}
	if err := e.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	return path
}

func TestBackups_makesListsAndFindsABackup(t *testing.T) {
	vault := BackupsAt(storeAt(t, "Write the spec"))

	made, err := vault.Make()
	if err != nil {
		t.Fatalf("Make: %v", err)
	}
	if made.Number != 1 {
		t.Errorf("Number = %d, want 1", made.Number)
	}

	all, err := vault.All()
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("All = %v, want one backup", all)
	}

	found, err := vault.Find(made.Number)
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	if found.Number != made.Number {
		t.Errorf("Find = %v, want the backup just made", found)
	}

	latest, err := vault.Latest()
	if err != nil {
		t.Fatalf("Latest: %v", err)
	}
	if latest.Number != made.Number {
		t.Errorf("Latest = %v, want the backup just made", latest)
	}

	held, err := vault.Holds(found)
	if err != nil {
		t.Fatalf("Holds: %v", err)
	}
	if held.Tasks != 1 {
		t.Errorf("Holds = %d, want 1", held)
	}
}

func TestBackups_deletesABackup(t *testing.T) {
	vault := BackupsAt(storeAt(t, "Write the spec"))
	made, err := vault.Make()
	if err != nil {
		t.Fatalf("Make: %v", err)
	}
	if err := vault.Delete(made.Number); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := vault.Find(made.Number); !errors.Is(err, ErrNoSuchBackup) {
		t.Errorf("err = %v, want ErrNoSuchBackup", err)
	}
}

func TestBackups_putsABackupBack(t *testing.T) {
	path := storeAt(t, "Write the spec")
	vault := BackupsAt(path)
	made, err := vault.Make()
	if err != nil {
		t.Fatalf("Make: %v", err)
	}

	e, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if _, err := e.add("Ship it", Changes{}); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if err := e.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	if err := vault.Restore(made.Number); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	back, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer func() { _ = back.Close() }()
	all, err := back.tasks.All()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(all) != 1 {
		t.Errorf("%d tasks, want the one the backup held", len(all))
	}
}

func TestBackups_saysThereIsNoBackupToPutBack(t *testing.T) {
	if _, err := BackupsAt(storeAt(t)).Latest(); !errors.Is(err, ErrNoBackup) {
		t.Errorf("err = %v, want ErrNoBackup", err)
	}
}

func TestEngine_namesEveryChangeAndItsState(t *testing.T) {
	e := openEngine(t)
	all, err := e.migrations()
	if err != nil {
		t.Fatalf("Migrations: %v", err)
	}
	if len(all) == 0 {
		t.Fatal("Migrations should name every change this cairn knows")
	}
	for _, one := range all {
		change := factsOfChange(one)
		if change.name == "" || change.does == "" {
			t.Errorf("a change should carry its name and what it does, got %+v", change)
		}
		if !change.applied {
			t.Errorf("a change on a store made today should be applied, got %+v", change)
		}
	}
}

func TestEngine_saysNothingIsWaitingOnAStoreItJustMade(t *testing.T) {
	waiting, err := openEngine(t).store.Waiting()
	if err != nil {
		t.Fatalf("Waiting: %v", err)
	}
	if len(waiting) != 0 {
		t.Errorf("Waiting = %v, want none: making a store makes it current", waiting)
	}
}

func TestEngine_countsWhatTheStoreHolds(t *testing.T) {
	e := openEngine(t)
	add(t, e, "Write the spec")
	held, err := e.held()
	if err != nil {
		t.Fatalf("Held: %v", err)
	}
	if held.Tasks != 1 {
		t.Errorf("Held = %+v, want 1 task", held)
	}
}

// behindStore gives a store one change short of what this cairn knows, by
// rolling one back through the engine itself.
func behindStore(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "cairn.db")
	e, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if _, err := e.rollBack(path); err != nil {
		t.Fatalf("RollBack: %v", err)
	}
	if err := e.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	return path
}

func TestEngine_rollsBackTheLastChangeAndBackupsFirst(t *testing.T) {
	path := behindStore(t)
	backups, err := BackupsAt(path).All()
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(backups) != 1 {
		t.Errorf("%d backups, want one set aside before rolling back", len(backups))
	}
}

func TestEngine_appliesWhatIsWaitingAndBackupsFirst(t *testing.T) {
	path := behindStore(t)
	e, err := OpenBehind(path)
	if err != nil {
		t.Fatalf("OpenBehind: %v", err)
	}
	defer func() { _ = e.Close() }()

	made, err := e.applyAll(path)
	if err != nil {
		t.Fatalf("ApplyAll: %v", err)
	}
	if len(made) != 1 {
		t.Errorf("applied %v, want the one change waiting", made)
	}
}

func TestEngine_appliesOneChangeByItsNumber(t *testing.T) {
	path := behindStore(t)
	e, err := OpenBehind(path)
	if err != nil {
		t.Fatalf("OpenBehind: %v", err)
	}
	defer func() { _ = e.Close() }()

	waiting, err := e.store.Waiting()
	if err != nil {
		t.Fatalf("Waiting: %v", err)
	}
	if len(waiting) != 1 {
		t.Fatalf("%d changes waiting, want 1", len(waiting))
	}
	name, err := e.apply(path, factsOfChange(waiting[0]).number)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if name == "" {
		t.Error("Apply should name the change it made")
	}
}

func TestEngine_refusesToOpenAStoreBehindIt(t *testing.T) {
	if _, err := Open(behindStore(t)); !errors.Is(err, ErrStoreBehind) {
		t.Errorf("err = %v, want ErrStoreBehind", err)
	}
}

func TestEngine_reportsAPlaceItCannotBackupBeforeChanging(t *testing.T) {
	e := openEngine(t)
	// A file where a folder must go: the backup cannot be written.
	blocked := filepath.Join(t.TempDir(), "cairn.db")
	if err := os.WriteFile(blocked, []byte("x"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	nowhere := filepath.Join(blocked, "nested", "cairn.db")

	if _, err := e.rollBack(nowhere); err == nil {
		t.Error("RollBack should report a backup it could not make")
	}
	if _, err := e.applyAll(nowhere); err == nil {
		t.Error("ApplyAll should report a backup it could not make")
	}
	if _, err := e.apply(nowhere, 2); err == nil {
		t.Error("Apply should report a backup it could not make")
	}
}

// A store that cannot be used as it stands is told what to type about it.
// Written once here and not at each door into the store: three of them open
// one, and the two boards used to leave somebody to go and find out.
func TestTold_saysWhatToRunAboutAStoreThatIsBehind(t *testing.T) {
	behind := fmt.Errorf("%w by 1 change(s)", ErrStoreBehind)
	got := Told(behind)
	if !errors.Is(got, ErrStoreBehind) {
		t.Errorf("Told lost what the error was about: %v", got)
	}
	if !strings.Contains(got.Error(), "cairn migration apply") {
		t.Errorf("Told = %q, and says nothing about what to run", got)
	}
}

// Anything else passes through untouched: an error nobody can act on is not
// improved by being handed a command to run.
func TestTold_leavesEveryOtherErrorAlone(t *testing.T) {
	// The message, and not errors.Is: what is being asked is that nothing was
	// added, and errors.Is would say yes to an error that had been wrapped —
	// which is the very thing this is watching for.
	for _, err := range []error{ErrNewerStore, ErrNotAStore, errors.New("something else")} {
		if got := Told(err); got.Error() != err.Error() {
			t.Errorf("Told(%v) = %v, want it left alone", err, got)
		}
	}
	if got := Told(nil); got != nil {
		t.Errorf("Told(nil) = %v, want nothing", got)
	}
}
