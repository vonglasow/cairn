package engine

// TestConfigShow_isOfferedAndAnswersAnObject is the fiche's named test: the
// operation exists, engine.All() offers it at Read level without opening the
// store, and it answers an object naming the two emplacements and what named
// each — read off config.Places through the Tell/Told membrane the config
// package already poses, never recomputed as a string comparison.

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"cairn/internal/config"
)

func TestConfigShow_isOfferedAndAnswersAnObject(t *testing.T) {
	found := false
	for _, op := range All() {
		if op.Name() == "config_show" {
			found = true
		}
	}
	if !found {
		t.Fatal("All() does not offer config_show")
	}

	if ConfigShow.Level != Read {
		t.Errorf("Level = %v, want Read", ConfigShow.Level)
	}
	if ConfigShow.Needs != AConfiguration {
		t.Errorf("Needs = %v, want AConfiguration", ConfigShow.Needs)
	}
	if kind := reflect.TypeOf(ConfigShow.Answers).Kind(); kind != reflect.Struct {
		t.Errorf("Answers is a %v, want a struct", kind)
	}

	home := t.TempDir()
	dir := filepath.Join(home, ".config", "cairn")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("could not set up the fixture: %v", err)
	}
	content := "store = \"/tmp/quelquepart/cairn.db\"\n"
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(content), 0o600); err != nil {
		t.Fatalf("could not set up the fixture: %v", err)
	}
	env := func(key string) (string, bool) {
		if key == config.NotesVariable {
			return "/somewhere/of/my/own/notes", true
		}
		return "", false
	}
	reading, err := config.NewReading(home, env)
	if err != nil {
		t.Fatalf("NewReading returned an error: %v", err)
	}

	answer, err := ConfigShow.Run(On{Reading: reading}, Input{})
	if err != nil {
		t.Fatalf("Run returned an error: %v", err)
	}
	if answer.Config == nil {
		t.Fatal("answer.Config is nil, want an object")
	}
	if want := filepath.Join(dir, "config.toml"); answer.Config.File != want {
		t.Errorf("File = %q, want %q", answer.Config.File, want)
	}
	if want := (Emplacement{At: "/tmp/quelquepart/cairn.db", By: "the file"}); answer.Config.Store != want {
		t.Errorf("Store = %+v, want %+v", answer.Config.Store, want)
	}
	if want := (Emplacement{At: "/somewhere/of/my/own/notes", By: "the environment"}); answer.Config.Notes != want {
		t.Errorf("Notes = %+v, want %+v", answer.Config.Notes, want)
	}
}

// TestConfigShow_asksForAConfigurationRatherThanAStore is the fiche's named
// test: ConfigShow does not ask for the store to be opened at all, not even
// the way a backup does — a backup still needs the path places() resolves
// to find itself beside where the store would be, and ConfigShow has no such
// use for it. The two must not share one value, or the day another operation
// wants ConfigShow's tolerant reading, nothing will tell them apart.
func TestConfigShow_asksForAConfigurationRatherThanAStore(t *testing.T) {
	if ConfigShow.Needs != AConfiguration {
		t.Errorf("ConfigShow.Needs = %v, want AConfiguration", ConfigShow.Needs)
	}
	for _, op := range []Operation{BackupAdd, BackupList, BackupShow, BackupDelete, BackupRestore} {
		if op.Needs != NoStore {
			t.Errorf("%s %s: Needs = %v, want NoStore unchanged", op.Thing, op.Verb, op.Needs)
		}
	}
}
