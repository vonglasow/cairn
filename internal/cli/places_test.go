package cli

import (
	"path/filepath"
	"testing"
)

// With no home to put the default under and nothing naming the notes, the
// notes are left unsaid rather than sent somewhere relative to wherever the
// command ran; the store, named by the environment, is still found. Named by
// the environment, they are said whatever the home.
func TestPlaces_leaveTheNotesUnsaidWithNoHomeToPutThemUnder(t *testing.T) {
	store := filepath.Join(t.TempDir(), "cairn.db")
	t.Setenv("HOME", "")
	t.Setenv("CAIRN_STORE", store)
	t.Setenv("CAIRN_NOTES", "")
	at, err := places()
	if err != nil {
		t.Fatalf("places: %v", err)
	}
	if at.Store != store || at.Notes != "" {
		t.Errorf("places = %+v, want the store and no notes", at)
	}

	notes := filepath.Join(t.TempDir(), "vault")
	t.Setenv("CAIRN_NOTES", notes)
	at, err = places()
	if err != nil {
		t.Fatalf("places: %v", err)
	}
	if at.Notes != notes {
		t.Errorf("the notes are at %q, want %q", at.Notes, notes)
	}
}
