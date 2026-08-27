package trace

// How an act crosses to a machine. Every fact a line shows, and the moment as a
// machine reads one — never as a person does, since a machine reading this has
// to parse it without guessing at a locale.

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestOfAct_carriesEveryFactALineShows(t *testing.T) {
	got := OfAct(acted(4, Changed, "priority", "normal", "high", "mcp", 9, 15))
	for _, tc := range []struct {
		what      string
		got, want any
	}{
		{"the task", got.Task, 4},
		{"what was done", got.Did, Changed},
		{"the attribute", got.Field, "priority"},
		{"what it was", got.Was, "normal"},
		{"what it became", got.Became, "high"},
		{"the surface", got.Through, "mcp"},
	} {
		if tc.got != tc.want {
			t.Errorf("%s = %v, want %v", tc.what, tc.got, tc.want)
		}
	}
}

// The moment crosses in the one format a machine reads without guessing, and in
// UTC, so what a script gets does not shift with whoever ran the command.
func TestOfAct_writesTheMomentTheWayAMachineReadsOne(t *testing.T) {
	got := OfAct(acted(1, FocusTaken, "", "", "", "", 9, 15)).At
	when, err := time.Parse(time.RFC3339, got)
	if err != nil {
		t.Fatalf("at = %q, which is not a moment a machine can read: %v", got, err)
	}
	if when.Location() != time.UTC && when.UTC() != when {
		t.Errorf("at = %q, want it in UTC", got)
	}
}

// What was absent is left out rather than crossing as an empty string: a machine
// asking whether an attribute changed should not have to tell "" from absent.
func TestOfAct_leavesOutWhatWasAbsent(t *testing.T) {
	written, err := json.Marshal(OfAct(acted(1, FocusTaken, "", "", "", "", 9, 15)))
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	for _, absent := range []string{"field", "was", "became", "through"} {
		if strings.Contains(string(written), absent) {
			t.Errorf("%q crosses although nothing was there: %s", absent, written)
		}
	}
	for _, present := range []string{"task", "did", "at"} {
		if !strings.Contains(string(written), present) {
			t.Errorf("%q does not cross, and every act has one: %s", present, written)
		}
	}
}
