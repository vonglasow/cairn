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

func shownWith(t *testing.T, content string) Config {
	t.Helper()
	home := t.TempDir()
	dir := filepath.Join(home, ".config", "cairn")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if content != "" {
		if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	answer, err := ConfigShow.Run(On{Home: Home{Dir: home, Lookup: func(string) (string, bool) { return "", false }}}, Input{})
	if err != nil || answer.Config == nil {
		t.Fatalf("config show: %v", err)
	}
	return *answer.Config
}

func TestConfigShow_answersEveryTableWhenTheFileSaysNothing(t *testing.T) {
	c := shownWith(t, "")
	if c.Pages == nil || *c.Pages != (PagesConfig{Tasks: true, Objectives: true, Time: true}) {
		t.Errorf("Pages = %+v, want every page on", c.Pages)
	}
	if c.Archive == nil || *c.Archive != (Archive{Days: 14}) {
		t.Errorf("Archive = %+v, want no start, in periods of 14 days", c.Archive)
	}
	if c.Focus == nil || c.Focus.Session != "" {
		t.Errorf("Focus = %+v, want a session that is not said", c.Focus)
	}
	if c.Time == nil || c.Time.Deleted != "kept" {
		t.Errorf("Time = %+v, want a deleted task's time kept", c.Time)
	}
	if c.Window == nil || c.Window.KeepSizes {
		t.Errorf("Window = %+v, want the sizes not kept", c.Window)
	}
}

func TestConfigShow_answersWhatTheFileSays(t *testing.T) {
	c := shownWith(t, "[pages]\ntime = false\n[archive]\nstart = 2026-10-21\ndays = 7\n[focus]\nsession = \"1h\"\n[time]\ndeleted = \"taken\"\n[window]\nkeep_sizes = true\n")
	if *c.Pages != (PagesConfig{Tasks: true, Objectives: true}) || *c.Archive != (Archive{Start: "2026-10-21", Days: 7}) ||
		c.Focus.Session != "1h" || c.Time.Deleted != "taken" || !c.Window.KeepSizes {
		t.Errorf("config show answers %+v %+v %+v %+v %+v", c.Pages, c.Archive, c.Focus, c.Time, c.Window)
	}
}

func TestConfigText_saysEveryKeyUnderItsTable(t *testing.T) {
	said := configText(shownWith(t, ""))
	for _, line := range []string{"\n[pages]\ntasks = true\nobjectives = true\ntime = true\n", "\n[archive]\n", "days = 14", "\n[focus]\n", "\n[time]\ndeleted = \"kept\"\n", "\n[window]\nkeep_sizes = false\n", "store = \"", "notes = \"", "plugin = \""} {
		if !strings.Contains(said, line) {
			t.Errorf("config show does not say %q:\n%s", line, said)
		}
	}
}

func TestConfigText_readsBackAsTheSameConfiguration(t *testing.T) {
	first := shownWith(t, "notes = \"/x/notes\"\n[pages]\nobjectives = false\n[archive]\nstart = 2026-10-21\ndays = 10\n[focus]\nsession = \"45m\"\n[time]\ndeleted = \"taken\"\n[window]\nkeep_sizes = true\n")
	again := shownWith(t, configText(first))
	if again.Problem != "" {
		t.Fatalf("what config show says is refused as a configuration: %s\n%s", again.Problem, configText(first))
	}
	if again.Store.At != first.Store.At || again.Notes.At != first.Notes.At || again.Plugin.At != first.Plugin.At ||
		*again.Pages != *first.Pages || *again.Archive != *first.Archive || *again.Focus != *first.Focus ||
		*again.Time != *first.Time || *again.Window != *first.Window {
		t.Errorf("read back, it says:\n%s\nwhere it said:\n%s", configText(again), configText(first))
	}
}

func TestConfigText_saysWhatNamedEachPlaceAndWhatIsNotSaid(t *testing.T) {
	said := configText(shownWith(t, ""))
	for _, words := range []string{"named by nothing", "nothing is archived on its own", "never asked to be renewed"} {
		if !strings.Contains(said, words) {
			t.Errorf("config show does not say %q:\n%s", words, said)
		}
	}
}

func TestConfigShow_saysWhetherTheWeekendsAreShown(t *testing.T) {
	if c := shownWith(t, ""); c.Time == nil || c.Time.Weekends != "shown" {
		t.Errorf("Time = %+v, want the weekends shown", c.Time)
	}
	c := shownWith(t, "[time]\nweekends = \"hidden\"\n")
	if c.Time.Weekends != "hidden" {
		t.Errorf("Time = %+v, want the weekends hidden", c.Time)
	}
	if said := configText(c); !strings.Contains(said, "\n[time]\ndeleted = \"kept\"\nweekends = \"hidden\"\n") {
		t.Errorf("config show does not say the weekends are hidden:\n%s", said)
	}
	if again := shownWith(t, configText(c)); again.Problem != "" || again.Time.Weekends != "hidden" {
		t.Errorf("read back, the weekends are %+v (%s)", again.Time, again.Problem)
	}
}
