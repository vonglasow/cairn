package engine

import (
	"cairn/internal/engine/backup"
	"errors"
	"strconv"
	"testing"
)

func TestBackupShow_saysWhenABackupWasMadeAndWhatItHolds(t *testing.T) {
	path := storeAt(t)
	if _, err := Do(TaskNew, path, Input{Given: map[string]string{"title": "one"}}); err != nil {
		t.Fatalf("Do(TaskNew): %v", err)
	}
	made, err := Do(BackupNew, path, Input{})
	if err != nil {
		t.Fatalf("Do(BackupNew): %v", err)
	}

	answer, err := Do(backupShow, path, Input{Given: map[string]string{"number": strconv.Itoa(made.Backup.Number)}})
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
	made, err := Do(BackupNew, path, Input{})
	if err != nil {
		t.Fatalf("Do(BackupNew): %v", err)
	}

	answer, err := Do(backupDelete, path, Input{Confirmed: true, Given: map[string]string{"number": strconv.Itoa(made.Backup.Number)}})
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
	if _, err := Do(BackupNew, path, Input{}); err != nil {
		t.Fatalf("Do(BackupNew): %v", err)
	}
	latest, err := Do(BackupNew, path, Input{})
	if err != nil {
		t.Fatalf("Do(BackupNew): %v", err)
	}

	if _, err := Do(BackupRestore, path, Input{Confirmed: true, Given: map[string]string{"number": "99"}}); !errors.Is(err, backup.ErrNoSuchBackup) {
		t.Fatalf("err = %v, want %v", err, backup.ErrNoSuchBackup)
	}

	answer, err := Do(BackupRestore, path, Input{Confirmed: true})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if answer.Restored == nil || *answer.Restored != latest.Backup.Number {
		t.Fatalf("answer.Restored = %v, want %d", answer.Restored, latest.Backup.Number)
	}

	first, err := Do(BackupRestore, path, Input{Confirmed: true, Given: map[string]string{"number": "1"}})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if first.Restored == nil || *first.Restored != 1 {
		t.Fatalf("first.Restored = %v, want 1", first.Restored)
	}
}

func TestBackupAdd_countsBeforeMaking(t *testing.T) {
	path := storeAt(t)
	if _, err := Do(TaskNew, path, Input{Given: map[string]string{"title": "one"}}); err != nil {
		t.Fatalf("Do(TaskNew): %v", err)
	}

	answer, err := Do(BackupNew, path, Input{})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if answer.Backup == nil {
		t.Fatalf("answer.Backup = nil, want a backup")
	}
	if answer.Holds == nil || answer.Holds.Tasks != 1 {
		t.Fatalf("answer.Holds = %v, want a count of 1 task, taken before the backup was made", answer.Holds)
	}

	garbage := garbageStore(t)
	if _, err := Do(BackupNew, garbage, Input{}); err == nil {
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
