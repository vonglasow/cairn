package store

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func held(t *testing.T, path string, titles ...string) {
	t.Helper()
	s, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	for _, title := range titles {
		keep(t, s, title)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func vaultOn(t *testing.T, titles ...string) (*Vault, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "cairn.db")
	held(t, path, titles...)
	return NewVault(path), path
}

func TestVault_numbersTheFirstBackupOne(t *testing.T) {
	v, _ := vaultOn(t, "Write the spec")
	made, err := v.Make()
	if err != nil {
		t.Fatalf("Make: %v", err)
	}
	if made.Number != 1 {
		t.Errorf("Number = %d, want 1", made.Number)
	}
}

func TestVault_givesTheNextNumberToTheNextBackup(t *testing.T) {
	v, _ := vaultOn(t, "Write the spec")
	if _, err := v.Make(); err != nil {
		t.Fatalf("Make: %v", err)
	}
	second, err := v.Make()
	if err != nil {
		t.Fatalf("Make: %v", err)
	}
	if second.Number != 2 {
		t.Errorf("Number = %d, want 2", second.Number)
	}
}

func TestVault_neverHandsOutANumberFreedByADeletion(t *testing.T) {
	v, _ := vaultOn(t, "Write the spec")
	first, err := v.Make()
	if err != nil {
		t.Fatalf("Make: %v", err)
	}
	if err := v.Delete(first.Number); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	next, err := v.Make()
	if err != nil {
		t.Fatalf("Make: %v", err)
	}
	if next.Number != 2 {
		t.Errorf("Number = %d, want 2: a freed number is not handed out again", next.Number)
	}
}

func TestVault_listsEveryBackup(t *testing.T) {
	v, _ := vaultOn(t, "Write the spec")
	for range 2 {
		if _, err := v.Make(); err != nil {
			t.Fatalf("Make: %v", err)
		}
	}
	all, err := v.All()
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(all) != 2 || all[0].Number != 1 || all[1].Number != 2 {
		t.Errorf("All = %v, want backups 1 and 2 in order", all)
	}
}

func TestVault_listsNothingWhenNoBackupWasMade(t *testing.T) {
	v, _ := vaultOn(t, "Write the spec")
	all, err := v.All()
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(all) != 0 {
		t.Errorf("All = %v, want none", all)
	}
}

func TestVault_saysWhenABackupWasMade(t *testing.T) {
	v, _ := vaultOn(t, "Write the spec")
	made, err := v.Make()
	if err != nil {
		t.Fatalf("Make: %v", err)
	}
	if made.Made.IsZero() {
		t.Error("a backup should say when it was made")
	}
}

func TestVault_saysWhatABackupHolds(t *testing.T) {
	v, _ := vaultOn(t, "Write the spec", "Ship it")
	made, err := v.Make()
	if err != nil {
		t.Fatalf("Make: %v", err)
	}
	holds, err := v.Holds(made)
	if err != nil {
		t.Fatalf("Holds: %v", err)
	}
	if holds.Tasks != 2 {
		t.Errorf("Holds = %d, want 2", holds)
	}
}

func TestVault_findsABackupByItsNumber(t *testing.T) {
	v, _ := vaultOn(t, "Write the spec")
	if _, err := v.Make(); err != nil {
		t.Fatalf("Make: %v", err)
	}
	got, err := v.Find(1)
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	if got.Number != 1 {
		t.Errorf("Find = %v, want backup 1", got)
	}
}

func TestVault_saysThereIsNoSuchBackup(t *testing.T) {
	v, _ := vaultOn(t, "Write the spec")
	if _, err := v.Find(99); !errors.Is(err, ErrNoSuchBackup) {
		t.Errorf("err = %v, want ErrNoSuchBackup", err)
	}
	if err := v.Delete(99); !errors.Is(err, ErrNoSuchBackup) {
		t.Errorf("err = %v, want ErrNoSuchBackup", err)
	}
}

func TestVault_putsABackupBack(t *testing.T) {
	v, path := vaultOn(t, "Write the spec")
	if _, err := v.Make(); err != nil {
		t.Fatalf("Make: %v", err)
	}
	held(t, path, "Ship it")

	if err := v.Restore(1); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	s, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer func() { _ = s.Close() }()
	all, err := s.All()
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(all) != 1 || factsOf(all[0]).title != "Write the spec" {
		t.Errorf("All = %v, want what the backup held", all)
	}
}

func TestVault_backupsWhatItIsAboutToReplace(t *testing.T) {
	v, path := vaultOn(t, "Write the spec")
	if _, err := v.Make(); err != nil {
		t.Fatalf("Make: %v", err)
	}
	held(t, path, "Ship it")

	if err := v.Restore(1); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	all, err := v.All()
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("%d backups, want 2: what was replaced should have been copied first", len(all))
	}
}

func TestVault_findsTheMostRecentBackup(t *testing.T) {
	v, _ := vaultOn(t, "Write the spec")
	for range 2 {
		if _, err := v.Make(); err != nil {
			t.Fatalf("Make: %v", err)
		}
	}
	latest, err := v.Latest()
	if err != nil {
		t.Fatalf("Latest: %v", err)
	}
	if latest.Number != 2 {
		t.Errorf("Latest = %v, want backup 2", latest)
	}
}

func TestVault_saysThereIsNoBackupToPutBack(t *testing.T) {
	v, _ := vaultOn(t, "Write the spec")
	if _, err := v.Latest(); !errors.Is(err, ErrNoBackup) {
		t.Errorf("err = %v, want ErrNoBackup", err)
	}
}

func TestVault_reportsAPlaceItCannotUse(t *testing.T) {
	v := NewVault(filepath.Join(t.TempDir(), "nothing", "here", "cairn.db"))
	if _, err := v.Make(); err == nil {
		t.Error("Make should report a store it cannot backup")
	}
}

func TestVault_countsWhatTheStoreHolds(t *testing.T) {
	v, _ := vaultOn(t, "Write the spec", "Ship it")
	held, err := v.HoldsStore()
	if err != nil {
		t.Fatalf("HoldsStore: %v", err)
	}
	if held.Tasks != 2 {
		t.Errorf("HoldsStore = %d, want 2", held)
	}
}

func TestVault_refusesToCountWhatIsNotAStore(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cairn.db")
	if err := os.WriteFile(path, []byte("this is not a cairn store"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if _, err := NewVault(path).HoldsStore(); !errors.Is(err, ErrNotAStore) {
		t.Errorf("err = %v, want ErrNotAStore", err)
	}
}

func TestHeldOn_readsTheOlderShapeThatHasNoObjectives(t *testing.T) {
	older := filepath.Join(t.TempDir(), "older.db")
	db, err := sql.Open("sqlite", older)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer func() { _ = db.Close() }()
	if _, err := db.Exec(
		`CREATE TABLE tasks (number INTEGER PRIMARY KEY); INSERT INTO tasks VALUES (1), (2)`); err != nil {
		t.Fatalf("laying down the older shape: %v", err)
	}

	held, err := heldOn(db)
	if err != nil {
		t.Fatalf("heldOn: %v", err)
	}
	if held.Tasks != 2 || held.Objectives != 0 {
		t.Errorf("it holds %+v, want 2 tasks and no objectives", held)
	}
}

func TestNumberOf_takesNoNameButOurs(t *testing.T) {
	for _, tc := range []struct {
		name   string
		number int
		ours   bool
	}{
		{"cairn-0004.db", 4, true},
		{"notes.txt", 0, false},
		{"cairn-0004.sqlite", 0, false},
		{"cairn-four.db", 0, false},
		{"cairn-.db", 0, false},
	} {
		number, ours := numberOf(tc.name)
		if ours != tc.ours || number != tc.number {
			t.Errorf("numberOf(%q) = %d, %v — want %d, %v",
				tc.name, number, ours, tc.number, tc.ours)
		}
	}
}

func TestVault_backsUpWhatAStoreHeldOpenHasNotWrittenBack(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cairn.db")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()
	keep(t, s, "Write the spec")

	made, err := NewVault(path).Make()
	if err != nil {
		t.Fatalf("Make: %v", err)
	}
	holds, err := NewVault(path).Holds(made)
	if err != nil {
		t.Fatal(err)
	}
	if holds.Tasks != 1 {
		t.Errorf("the backup holds %d tasks, want the one written while the store was held", holds.Tasks)
	}
}

func TestVault_putsABackupBackWithNothingOfTheOldStoreBesideIt(t *testing.T) {
	v, path := vaultOn(t, "Write the spec")
	made, err := v.Make()
	if err != nil {
		t.Fatal(err)
	}
	for _, beside := range []string{path + "-wal", path + "-shm"} {
		if err := os.WriteFile(beside, []byte("left over"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := v.Restore(made.Number); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	for _, beside := range []string{path + "-wal", path + "-shm"} {
		if _, err := os.Stat(beside); !errors.Is(err, os.ErrNotExist) {
			t.Errorf("%s is still there after a backup was put back", filepath.Base(beside))
		}
	}
	held, err := v.HoldsStore()
	if err != nil || held.Tasks != 1 {
		t.Errorf("the store put back holds %+v, %v, want the one task", held, err)
	}
}

func TestVault_backsUpADamagedStoreAsItLies(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cairn.db")
	if err := os.WriteFile(path, []byte("not a database at all"), 0o600); err != nil {
		t.Fatal(err)
	}
	made, err := NewVault(path).Make()
	if err != nil {
		t.Fatalf("Make: %v", err)
	}
	kept, err := os.ReadFile(made.Path)
	if err != nil || string(kept) != "not a database at all" {
		t.Errorf("the backup holds %q, %v, want the file as it lay", kept, err)
	}
}

func ownerOnly(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if mode := info.Mode().Perm(); mode&0o077 != 0 {
		t.Errorf("%s is %v, want readable by its owner alone: the folder it sits in is", filepath.Base(path), mode)
	}
}

func TestVault_keepsABackupForItsOwnerAlone(t *testing.T) {
	v, _ := vaultOn(t, "Write the spec")
	made, err := v.Make()
	if err != nil {
		t.Fatalf("Make: %v", err)
	}
	ownerOnly(t, made.Path)
}

func TestVault_keepsABackupOfADamagedStoreForItsOwnerAlone(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cairn.db")
	if err := os.WriteFile(path, []byte("not a database at all"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	made, err := NewVault(path).Make()
	if err != nil {
		t.Fatalf("Make: %v", err)
	}
	ownerOnly(t, made.Path)
}

func TestVault_keepsARestoredStoreForItsOwnerAlone(t *testing.T) {
	v, path := vaultOn(t, "Write the spec")
	made, err := v.Make()
	if err != nil {
		t.Fatalf("Make: %v", err)
	}
	if err := v.Restore(made.Number); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	ownerOnly(t, path)
}

func TestVault_skipsANumberAnotherCairnTookMeanwhile(t *testing.T) {
	v, _ := vaultOn(t, "Write the spec")
	if _, err := v.Make(); err != nil {
		t.Fatalf("Make: %v", err)
	}
	theirs := v.pathOf(2)
	if err := os.WriteFile(theirs, []byte("theirs"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	made, err := v.Make()
	if err != nil {
		t.Fatalf("Make: %v", err)
	}
	if made.Number != 3 {
		t.Errorf("Number = %d, want 3: a number another cairn took is not taken again", made.Number)
	}
	if got, _ := os.ReadFile(theirs); string(got) != "theirs" {
		t.Errorf("the other cairn's backup was overwritten")
	}
}

func TestVault_refusesToPutBackADamagedBackup(t *testing.T) {
	v, path := vaultOn(t, "Write the spec")
	made, err := v.Make()
	if err != nil {
		t.Fatalf("Make: %v", err)
	}
	if err := os.WriteFile(made.Path, []byte("not a database at all"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := v.Restore(made.Number); !errors.Is(err, ErrDamagedBackup) {
		t.Fatalf("Restore = %v, want ErrDamagedBackup", err)
	}
	s, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v — the store was replaced by what could not be read", err)
	}
	defer func() { _ = s.Close() }()
	if held, _ := s.Held(); held.Tasks != 1 {
		t.Errorf("%d tasks, want the store left as it was", held.Tasks)
	}
}

func TestOpen_keepsAStoreWhosePathReadsLikeAnAddress(t *testing.T) {
	path := filepath.Join(t.TempDir(), "odd?mode=ro#and 100%.db")
	held(t, path, "Write the spec")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("the store is not where it was asked to be: %v", err)
	}
	s, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer func() { _ = s.Close() }()
	if held, _ := s.Held(); held.Tasks != 1 {
		t.Errorf("%d tasks, want the one kept", held.Tasks)
	}
	made, err := NewVault(path).Make()
	if err != nil {
		t.Fatalf("Make: %v", err)
	}
	if got, _ := NewVault(path).Holds(made); got.Tasks != 1 {
		t.Errorf("the backup holds %d tasks, want 1", got.Tasks)
	}
}

func TestOpen_makesAStoreForItsOwnerAlone(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cairn.db")
	held(t, path, "Write the spec")
	ownerOnly(t, path)
}

func TestVault_leavesNoBackupBehindWhenMakingOneFails(t *testing.T) {
	v := NewVault(filepath.Join(t.TempDir(), "nowhere", "cairn.db"))
	if _, err := v.Make(); err == nil {
		t.Fatal("Make succeeded with no store to back up")
	}
	if all, _ := v.All(); len(all) != 0 {
		t.Errorf("%d backups listed, want none: a backup that failed is not a backup", len(all))
	}
}

func TestVault_refusesToPutBackABackupThatHoldsNoTasks(t *testing.T) {
	v, path := vaultOn(t, "Write the spec")
	made, err := v.Make()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(made.Path, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := v.Restore(made.Number); !errors.Is(err, ErrDamagedBackup) {
		t.Fatalf("Restore of an empty file = %v, want ErrDamagedBackup", err)
	}
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()
	if held, _ := s.Held(); held.Tasks != 1 {
		t.Errorf("%d tasks, want the store left as it was", held.Tasks)
	}
}

func TestSnapshot_saysWhyAReadableStoreWasNotBackedUp(t *testing.T) {
	_, path := vaultOn(t, "Write the spec")
	taken := filepath.Join(t.TempDir(), "taken.db")
	if err := os.WriteFile(taken, []byte("somebody else's"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := snapshot(path, taken); err == nil {
		t.Error("snapshot copied a readable store by hand rather than saying why it could not read it")
	}
	if got, _ := os.ReadFile(taken); string(got) != "somebody else's" {
		t.Error("what was in the way was overwritten")
	}
}

func TestVault_putsBackThroughNoLinkLeftInItsWay(t *testing.T) {
	v, path := vaultOn(t, "Write the spec")
	made, err := v.Make()
	if err != nil {
		t.Fatal(err)
	}
	elsewhere := filepath.Join(t.TempDir(), "somebody's file")
	if err := os.WriteFile(elsewhere, []byte("keep me"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(elsewhere, path+".arriving"); err != nil {
		t.Fatal(err)
	}
	if err := v.Restore(made.Number); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if got, _ := os.ReadFile(elsewhere); string(got) != "keep me" {
		t.Error("the restore wrote through a link somebody left beside the store")
	}
}

func TestStore_saysWhenItsFileWasReplaced(t *testing.T) {
	v, path := vaultOn(t, "Write the spec")
	made, err := v.Make()
	if err != nil {
		t.Fatal(err)
	}
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()
	if replaced, err := s.Replaced(); err != nil || replaced {
		t.Fatalf("Replaced = %v, %v before anything happened", replaced, err)
	}
	if err := v.Restore(made.Number); err != nil {
		t.Fatal(err)
	}
	if replaced, err := s.Replaced(); err != nil || !replaced {
		t.Errorf("Replaced = %v, %v after a backup was put back under it", replaced, err)
	}
}

func TestVault_looksAtABackupWithoutTouchingIt(t *testing.T) {
	v, _ := vaultOn(t, "Write the spec")
	made, err := v.Make()
	if err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(made.Path)
	if err != nil {
		t.Fatal(err)
	}
	stamp, _ := os.Stat(made.Path)
	if _, err := v.Holds(made); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(made.Path)
	again, _ := os.Stat(made.Path)
	if string(after) != string(before) || !again.ModTime().Equal(stamp.ModTime()) {
		t.Error("looking at a backup changed it")
	}
}

func TestVault_makesNoStoreByLookingForOne(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cairn.db")
	if _, err := NewVault(path).HoldsStore(); err == nil {
		t.Error("HoldsStore answered about a store that is not there")
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("a store was made by looking for one: %v", err)
	}
}
