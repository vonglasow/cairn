package domain

import (
	"testing"
	"time"
)

var madeAt = time.Date(2026, 1, 2, 15, 4, 0, 0, time.UTC)

// TestOfBackup_carriesWhatItHeldWhenCounted pins the wire: what a machine
// reads for a backup carries its number, a moment written the way a person
// reads one, and what it held — only when that was counted.
func TestOfBackup_carriesWhatItHeldWhenCounted(t *testing.T) {
	backup := Backup{Number: 2, Made: madeAt}
	held := Held{Tasks: 3, Objectives: 1}

	got := OfBackup(backup, &held)

	want := BackupShape{Number: 2, Made: FormatMoment(madeAt), Holds: &Held{Tasks: 3, Objectives: 1}}
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
	got := OfBackup(Backup{Number: 1, Made: madeAt}, nil)
	if got.Holds != nil {
		t.Errorf("Holds = %+v, want nil: nothing was counted", got.Holds)
	}
}
