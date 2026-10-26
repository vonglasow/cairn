package engine

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
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
	answer, err := ConfigShow.Run(On{Home: Home{Dir: home, Lookup: env}}, Input{})
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

func TestConfigShow_asksForAConfigurationRatherThanAStore(t *testing.T) {
	if ConfigShow.Needs != AConfiguration {
		t.Errorf("ConfigShow.Needs = %v, want AConfiguration", ConfigShow.Needs)
	}
	for _, op := range []Operation{BackupNew, BackupList, BackupShow, BackupDelete, BackupRestore} {
		if op.Needs != NoStore {
			t.Errorf("%s %s: Needs = %v, want NoStore unchanged", op.Thing, op.Verb, op.Needs)
		}
	}
}

func TestConfigInit_writesTheExampleAndNeverOverAFile(t *testing.T) {
	home := t.TempDir()
	answer, err := ConfigInit.Run(On{Home: Home{Dir: home}}, Input{})
	if err != nil {
		t.Fatal(err)
	}
	written, err := os.ReadFile(config.FileIn(home))
	if err != nil || string(written) != config.Example {
		t.Fatalf("it wrote %q, %v", written, err)
	}
	if answer.Written == nil || *answer.Written != config.FileIn(home) {
		t.Errorf("it says it wrote %v", answer.Written)
	}
	if said := RenderText(ConfigInit, answer); !strings.Contains(said, config.FileIn(home)) {
		t.Errorf("it says %q", said)
	}
	if _, err := ConfigInit.Run(On{Home: Home{Dir: home}}, Input{}); !errors.Is(err, ErrConfigAlreadyThere) {
		t.Errorf("writing it again = %v, want ErrConfigAlreadyThere", err)
	}
	if ConfigInit.Level != Write || ConfigInit.Needs != AConfiguration {
		t.Errorf("config init is level %v and needs %v", ConfigInit.Level, ConfigInit.Needs)
	}
}

func TestConfigShowSaysTheWindowKeepsItsSizes(t *testing.T) {
	said := configText(Config{Window: &WindowConfig{KeepSizes: true}})
	if !strings.Contains(said, "window      opens with the sizes it was left with") {
		t.Errorf("config show says:\n%s", said)
	}
	if said := configText(Config{}); strings.Contains(said, "window") {
		t.Errorf("with the sizes not kept, config show says:\n%s", said)
	}
}
