package store

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// held opens a store, puts n tasks in it, and closes it — leaving a real file
// for the backups to be made of.
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

// A file that answers nothing is not a store. Counting it as an empty one is
// how somebody else's file gets copied and reported as yours.
func TestVault_refusesToCountWhatIsNotAStore(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cairn.db")
	if err := os.WriteFile(path, []byte("this is not a cairn store"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if _, err := NewVault(path).HoldsStore(); !errors.Is(err, ErrNotAStore) {
		t.Errorf("err = %v, want ErrNotAStore", err)
	}
}
