package migration

import (
	"encoding/json"
	"testing"
	"time"

	"cairn/internal/domain"
)

// TestShape_marshalsUnderTheNamesAScriptReads pins the wire: the field names
// are what a script reads, and a script that reads one has no way to notice
// a renamed tag unless it is checked here.
func TestShape_marshalsUnderTheNamesAScriptReads(t *testing.T) {
	encoded, err := json.Marshal(OfMigration(builtChange(true)))
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var doc map[string]any
	if err := json.Unmarshal(encoded, &doc); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	for _, key := range []string{"number", "name", "applied", "made"} {
		if _, present := doc[key]; !present {
			t.Errorf("no %q key in %s", key, encoded)
		}
	}
}

func TestOfMigration_carriesEveryFieldAScriptReads(t *testing.T) {
	got := OfMigration(builtChange(true))
	made := domain.FormatMoment(time.Date(2026, 1, 2, 15, 4, 0, 0, time.UTC))
	want := Shape{Number: 2, Name: "a task can hang under another", Applied: true, Made: made}
	if got != want {
		t.Errorf("OfMigration() = %+v, want %+v", got, want)
	}
}

// A change waiting has no moment to report yet: Made is left out until it is
// applied, not written as a zero value nobody asked about.
func TestOfMigration_leavesMadeOutUntilApplied(t *testing.T) {
	got := OfMigration(builtChange(false))
	if got.Made != "" {
		t.Errorf("Made = %q, want none: nothing was applied yet", got.Made)
	}
}

func TestOfRolledBack_namesTheChangeAndWhatItCost(t *testing.T) {
	got := OfRolledBack(builtChange(true))
	want := RolledBack{RolledBack: "a task can hang under another", Lost: "which task each sub-task hangs under"}
	if got != want {
		t.Errorf("OfRolledBack() = %+v, want %+v", got, want)
	}
}
