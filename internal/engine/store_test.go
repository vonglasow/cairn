package engine

import (
	"errors"
	"os"
	"path/filepath"
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
		if _, err := e.Add(title, Changes{}); err != nil {
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
	if held != 1 {
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
	if _, err := e.Add("Ship it", Changes{}); err != nil {
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
	all, err := back.List()
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
	all, err := e.Migrations()
	if err != nil {
		t.Fatalf("Migrations: %v", err)
	}
	if len(all) == 0 {
		t.Fatal("Migrations should name every change this cairn knows")
	}
	for _, change := range all {
		if change.Name == "" || change.Does == "" {
			t.Errorf("a change should carry its name and what it does, got %+v", change)
		}
		if !change.Applied || change.Made.IsZero() {
			t.Errorf("a change on a store made today should be applied and dated, got %+v", change)
		}
	}
}

func TestEngine_saysNothingIsWaitingOnAStoreItJustMade(t *testing.T) {
	waiting, err := openEngine(t).Waiting()
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
	held, err := e.TasksHeld()
	if err != nil {
		t.Fatalf("TasksHeld: %v", err)
	}
	if held != 1 {
		t.Errorf("TasksHeld = %d, want 1", held)
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
	if _, err := e.RollBack(path); err != nil {
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

	made, err := e.ApplyAll(path)
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

	waiting, err := e.Waiting()
	if err != nil {
		t.Fatalf("Waiting: %v", err)
	}
	if len(waiting) != 1 {
		t.Fatalf("%d changes waiting, want 1", len(waiting))
	}
	name, err := e.Apply(path, waiting[0].Number)
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

	if _, err := e.RollBack(nowhere); err == nil {
		t.Error("RollBack should report a backup it could not make")
	}
	if _, err := e.ApplyAll(nowhere); err == nil {
		t.Error("ApplyAll should report a backup it could not make")
	}
	if _, err := e.Apply(nowhere, 2); err == nil {
		t.Error("Apply should report a backup it could not make")
	}
}
