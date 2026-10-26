package engine

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPlaces_keepADraftBesideTheStore(t *testing.T) {
	at := Places{Store: filepath.Join("/somewhere", "cairn.db")}
	if got := at.Draft(); got != filepath.Join("/somewhere", "cairn.db.draft.json") {
		t.Errorf("the draft of a new task is kept at %q", got)
	}
}

func TestPlaces_keepTheWindowsSizesBesideTheStore(t *testing.T) {
	at := Places{Store: filepath.Join("/somewhere", "cairn.db")}
	if got := at.Sizes(); got != filepath.Join("/somewhere", "cairn.db.sizes.json") {
		t.Errorf("the window's sizes are kept at %q", got)
	}
}

func TestKeepsSizes_readsTheConfiguration(t *testing.T) {
	home := t.TempDir()
	path := filepath.Join(home, ".config", "cairn", "config.toml")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if KeepsSizes(Home{Dir: home}) {
		t.Error("with no file, the window keeps its sizes")
	}
	if err := os.WriteFile(path, []byte("[window]\nkeep_sizes = true\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if !KeepsSizes(Home{Dir: home}) {
		t.Error("keep_sizes = true does not keep the window's sizes")
	}
	if KeepsSizes(Home{}) {
		t.Error("with no home, the window keeps its sizes")
	}
}
