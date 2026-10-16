package engine

import (
	"path/filepath"
	"testing"
)

func TestPlaces_keepADraftBesideTheStore(t *testing.T) {
	at := Places{Store: filepath.Join("/somewhere", "cairn.db")}
	if got := at.Draft(); got != filepath.Join("/somewhere", "cairn.db.draft.json") {
		t.Errorf("the draft of a new task is kept at %q", got)
	}
}
