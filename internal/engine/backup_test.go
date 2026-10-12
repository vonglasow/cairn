package engine

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"cairn/internal/domain"
)

func TestBackupList_listsEveryBackup(t *testing.T) {
	path := storeAt(t)
	if _, err := Do(BackupNew, path, Input{}); err != nil {
		t.Fatalf("Do(BackupNew): %v", err)
	}
	if _, err := Do(BackupNew, path, Input{}); err != nil {
		t.Fatalf("Do(BackupNew): %v", err)
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
	if _, err := Do(TaskNew, path, Input{Given: map[string]string{"title": "one"}}); err != nil {
		t.Fatalf("Do(TaskNew): %v", err)
	}
	made, err := Do(BackupNew, path, Input{})
	if err != nil {
		t.Fatalf("Do(BackupNew): %v", err)
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
	made, err := Do(BackupNew, path, Input{})
	if err != nil {
		t.Fatalf("Do(BackupNew): %v", err)
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
	if _, err := Do(BackupNew, path, Input{}); err != nil {
		t.Fatalf("Do(BackupNew): %v", err)
	}
	latest, err := Do(BackupNew, path, Input{})
	if err != nil {
		t.Fatalf("Do(BackupNew): %v", err)
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
	if answer.Restored == nil || *answer.Restored != latest.Backup.Number {
		t.Fatalf("answer.Restored = %v, want %d", answer.Restored, latest.Backup.Number)
	}

	// A number a backup does carry restores that one, not whichever the
	// number happens to resolve to by accident.
	first, err := Do(BackupRestore, path, Input{Given: map[string]string{"number": "1"}})
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

	// Counted before backing up, not after: a file that cannot be counted is
	// refused before any backup of it is made.
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

var madeAt = time.Date(2026, 1, 2, 15, 4, 0, 0, time.UTC)

// TestOfBackup_carriesWhatItHeldWhenCounted pins the wire: what a machine
// reads for a backup carries its number, a moment written the way a person
// reads one, and what it held — only when that was counted.
func TestOfBackup_carriesWhatItHeldWhenCounted(t *testing.T) {
	backup := domain.Backup{Number: 2, Made: madeAt}
	held := domain.Held{Tasks: 3, Objectives: 1}

	got := OfBackup(backup, &held)

	want := BackupShape{Number: 2, Made: domain.FormatMoment(madeAt), Holds: &domain.Held{Tasks: 3, Objectives: 1}}
	if got.Number != want.Number || got.Made != want.Made {
		t.Errorf("OfBackup() = %+v, want %+v", got, want)
	}
	if got.Holds == nil || *got.Holds != *want.Holds {
		t.Errorf("Holds = %+v, want %+v", got.Holds, want.Holds)
	}
}

// Making a backup does not open it, so nothing counted must read as nothing
// held — never as a zero, which would read as an empty store.
func TestOfBackup_leavesHoldsOutWhenNothingWasCounted(t *testing.T) {
	got := OfBackup(domain.Backup{Number: 1, Made: madeAt}, nil)
	if got.Holds != nil {
		t.Errorf("Holds = %+v, want nil: nothing was counted", got.Holds)
	}
}

// The store is written to a file named, and that file is a store holding what
// this one held; a file already there is refused.
func TestBackupExport_writesTheStoreToTheFileNamed(t *testing.T) {
	path := storeAt(t)
	if _, err := Do(TaskNew, path, Input{Given: map[string]string{"title": "one"}}); err != nil {
		t.Fatalf("Do(TaskNew): %v", err)
	}
	to := filepath.Join(t.TempDir(), "elsewhere.db")
	answer, err := Do(BackupExport, path, Input{Given: map[string]string{"to": to}})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if answer.Exported == nil || *answer.Exported != to || answer.Holds == nil || answer.Holds.Tasks != 1 {
		t.Fatalf("answered %v holding %v", answer.Exported, answer.Holds)
	}
	if got := RenderText(BackupExport, answer); got != to+" holds 1 task(s) and 0 objective(s)\n" {
		t.Errorf("it says %q", got)
	}
	if shape, ok := Of(BackupExport, answer).(Exported); !ok || shape.Exported != to || shape.Holds.Tasks != 1 {
		t.Errorf("a machine reads %#v", Of(BackupExport, answer))
	}
	listed, err := Do(TaskList, to, Input{})
	if err != nil || len(listed.Tasks) != 1 {
		t.Errorf("the file lists %v, %v", listed.Tasks, err)
	}
	if _, err := Do(BackupExport, path, Input{Given: map[string]string{"to": to}}); !errors.Is(err, os.ErrExist) {
		t.Errorf("exporting onto it again answered %v", err)
	}
}

// A store from a file replaces this one, which is backed up first; the file
// is left where it was.
func TestBackupImport_putsTheStoreFromAFileInPlace(t *testing.T) {
	path := storeAt(t)
	if _, err := Do(TaskNew, path, Input{Given: map[string]string{"title": "here"}}); err != nil {
		t.Fatalf("Do(TaskNew): %v", err)
	}
	from := storeAt(t)
	if _, err := Do(TaskNew, from, Input{Given: map[string]string{"title": "from elsewhere"}}); err != nil {
		t.Fatalf("Do(TaskNew): %v", err)
	}
	answer, err := Do(BackupImport, path, Input{Given: map[string]string{"from": from}})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if answer.Imported == nil || *answer.Imported != from || answer.Restored == nil || *answer.Restored != 1 {
		t.Fatalf("answered %v, put back %v", answer.Imported, answer.Restored)
	}
	if got := RenderText(BackupImport, answer); got != "the store now holds what "+from+" held — backup 1; a new backup of what it held before was made first\n" {
		t.Errorf("it says %q", got)
	}
	if shape, ok := Of(BackupImport, answer).(Imported); !ok || shape.Imported != from || shape.Backup != 1 {
		t.Errorf("a machine reads %#v", Of(BackupImport, answer))
	}
	listed, err := Do(TaskList, path, Input{})
	if err != nil || len(listed.Tasks) != 1 || !strings.Contains(RenderText(TaskList, listed), "from elsewhere") {
		t.Errorf("the store lists %v, %v", listed.Tasks, err)
	}
	if kept, _ := Do(BackupList, path, Input{}); len(kept.Backups) != 2 {
		t.Errorf("%d backups, want the one taken in and the one set aside", len(kept.Backups))
	}
	if BackupImport.Level != Full || BackupExport.Level != Write {
		t.Errorf("import is level %v and export %v", BackupImport.Level, BackupExport.Level)
	}
}
