package engine

import (
	"errors"
	"strconv"
	"testing"
)

func TestBackupList_listsEveryBackup(t *testing.T) {
	path := storeAt(t)
	if _, err := Do(BackupAdd, path, Input{}); err != nil {
		t.Fatalf("Do(BackupAdd): %v", err)
	}
	if _, err := Do(BackupAdd, path, Input{}); err != nil {
		t.Fatalf("Do(BackupAdd): %v", err)
	}

	answer, err := Do(BackupList, path, Input{})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if len(answer.Backups) != 2 {
		t.Fatalf("len(answer.Backups) = %d, want 2", len(answer.Backups))
	}
}

func TestBackupList_answersAnEmptySetNotNilWhenNoBackupWasEverMade(t *testing.T) {
	path := storeAt(t)

	answer, err := Do(BackupList, path, Input{})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if answer.Backups == nil {
		t.Fatalf("answer.Backups = nil, want an empty, non-nil set: no backup is a set with none in it, not an absent answer")
	}
}

func TestBackupShow_saysWhenABackupWasMadeAndWhatItHolds(t *testing.T) {
	path := storeAt(t)
	if _, err := Do(TaskAdd, path, Input{Given: map[string]string{"title": "one"}}); err != nil {
		t.Fatalf("Do(TaskAdd): %v", err)
	}
	made, err := Do(BackupAdd, path, Input{})
	if err != nil {
		t.Fatalf("Do(BackupAdd): %v", err)
	}

	answer, err := Do(BackupShow, path, Input{Given: map[string]string{"number": strconv.Itoa(made.Backup.Number)}})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if answer.Backup == nil || answer.Backup.Number != made.Backup.Number {
		t.Fatalf("answer.Backup = %v, want backup %d", answer.Backup, made.Backup.Number)
	}
	if answer.Holds == nil || answer.Holds.Tasks != 1 {
		t.Fatalf("answer.Holds = %v, want a count of 1 task", answer.Holds)
	}
}

func TestBackupDelete_deletesABackup(t *testing.T) {
	path := storeAt(t)
	made, err := Do(BackupAdd, path, Input{})
	if err != nil {
		t.Fatalf("Do(BackupAdd): %v", err)
	}

	answer, err := Do(BackupDelete, path, Input{Given: map[string]string{"number": strconv.Itoa(made.Backup.Number)}})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if answer.Deleted == nil || *answer.Deleted != made.Backup.Number {
		t.Fatalf("answer.Deleted = %v, want %d", answer.Deleted, made.Backup.Number)
	}

	list, err := Do(BackupList, path, Input{})
	if err != nil {
		t.Fatalf("Do(BackupList): %v", err)
	}
	if len(list.Backups) != 0 {
		t.Fatalf("len(list.Backups) = %d, want 0: the backup was not deleted", len(list.Backups))
	}
}

func TestBackupRestore_resolvesBeforeRestoring(t *testing.T) {
	path := storeAt(t)
	if _, err := Do(BackupAdd, path, Input{}); err != nil {
		t.Fatalf("Do(BackupAdd): %v", err)
	}
	latest, err := Do(BackupAdd, path, Input{})
	if err != nil {
		t.Fatalf("Do(BackupAdd): %v", err)
	}

	// A number nothing carries is refused before anything else is touched.
	if _, err := Do(BackupRestore, path, Input{Given: map[string]string{"number": "99"}}); !errors.Is(err, ErrNoSuchBackup) {
		t.Fatalf("err = %v, want %v", err, ErrNoSuchBackup)
	}

	// No number given restores the most recent one. Checked before the
	// explicit-number restore below, which backs up first and would
	// otherwise make a newer backup "the most recent one".
	answer, err := Do(BackupRestore, path, Input{})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if answer.PutBack == nil || *answer.PutBack != latest.Backup.Number {
		t.Fatalf("answer.PutBack = %v, want %d", answer.PutBack, latest.Backup.Number)
	}

	// A number a backup does carry restores that one, not whichever the
	// number happens to resolve to by accident.
	first, err := Do(BackupRestore, path, Input{Given: map[string]string{"number": "1"}})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if first.PutBack == nil || *first.PutBack != 1 {
		t.Fatalf("first.PutBack = %v, want 1", first.PutBack)
	}
}

func TestBackupAdd_countsBeforeMaking(t *testing.T) {
	path := storeAt(t)
	if _, err := Do(TaskAdd, path, Input{Given: map[string]string{"title": "one"}}); err != nil {
		t.Fatalf("Do(TaskAdd): %v", err)
	}

	answer, err := Do(BackupAdd, path, Input{})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if answer.Backup == nil {
		t.Fatalf("answer.Backup = nil, want a backup")
	}
	if answer.Holds == nil || answer.Holds.Tasks != 1 {
		t.Fatalf("answer.Holds = %v, want a count of 1 task, taken before the backup was made", answer.Holds)
	}

	// Counted before backing up, not after: a file that cannot be counted is
	// refused before any backup of it is made.
	garbage := garbageStore(t)
	if _, err := Do(BackupAdd, garbage, Input{}); err == nil {
		t.Fatalf("Do: want an error, got none: %q is not a store", garbage)
	}
	list, err := Do(BackupList, garbage, Input{})
	if err != nil {
		t.Fatalf("Do(BackupList): %v", err)
	}
	if len(list.Backups) != 0 {
		t.Fatalf("len(list.Backups) = %d, want 0: a backup of a file that cannot be counted must not be made", len(list.Backups))
	}
}
