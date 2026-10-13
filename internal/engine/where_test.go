package engine

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cairn/internal/config"
)

func environment(said map[string]string) func(string) (string, bool) {
	return func(name string) (string, bool) {
		value, ok := said[name]
		return value, ok
	}
}

func writeConfiguration(t *testing.T, home, content string) string {
	t.Helper()
	dir := filepath.Join(home, ".config", "cairn")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLocate_leavesTheNotesUnsaidWithNoHomeToPutThemUnder(t *testing.T) {
	store := filepath.Join(t.TempDir(), "cairn.db")
	nowhere := errors.New("no home")
	at, err := Locate(Home{Missing: nowhere, Lookup: environment(map[string]string{config.StoreVariable: store})})
	if err != nil {
		t.Fatalf("Locate: %v", err)
	}
	if at.Store != store || at.Notes != "" {
		t.Errorf("Locate = %+v, want the store and no notes", at)
	}

	notes := filepath.Join(t.TempDir(), "vault")
	at, err = Locate(Home{Missing: nowhere, Lookup: environment(map[string]string{
		config.StoreVariable: store, config.NotesVariable: notes,
	})})
	if err != nil {
		t.Fatalf("Locate: %v", err)
	}
	if at.Notes != notes {
		t.Errorf("the notes are at %q, want %q", at.Notes, notes)
	}
}

func TestLocate_refusesAStoreWithNowhereToGo(t *testing.T) {
	nowhere := errors.New("no home")
	_, err := Locate(Home{Missing: nowhere, Lookup: environment(nil)})
	if !errors.Is(err, ErrStore) || !errors.Is(err, nowhere) {
		t.Errorf("Locate = %v, want ErrStore wrapping why there is no home", err)
	}
}

func TestLocate_passesAMistakeInTheFileThrough(t *testing.T) {
	home := t.TempDir()
	writeConfiguration(t, home, "this is not toml: [[[\n")
	_, err := Locate(Home{Dir: home, Lookup: environment(nil)})
	if !errors.Is(err, config.ErrNotTOML) || errors.Is(err, ErrStore) {
		t.Errorf("Locate = %v, want ErrNotTOML and not ErrStore", err)
	}
}

func TestLocate_foldsAFileItCannotReadIntoTheStore(t *testing.T) {
	home := t.TempDir()
	path := writeConfiguration(t, home, "notes = \"~/vault\"\n")
	if err := os.Chmod(path, 0o000); err != nil {
		t.Fatal(err)
	}
	_, err := Locate(Home{Dir: home, Lookup: environment(nil)})
	if !errors.Is(err, ErrStore) {
		t.Errorf("Locate = %v, want ErrStore", err)
	}
}

func TestLocate_findsBothPlacesTheFileNames(t *testing.T) {
	home := t.TempDir()
	writeConfiguration(t, home, "store = \"~/work/cairn.db\"\nnotes = \"~/vault\"\n")
	at, err := Locate(Home{Dir: home, Lookup: environment(nil)})
	if err != nil {
		t.Fatalf("Locate: %v", err)
	}
	if want := filepath.Join(home, "work", "cairn.db"); at.Store != want {
		t.Errorf("the store is at %q, want %q", at.Store, want)
	}
	if want := filepath.Join(home, "vault"); at.Notes != want {
		t.Errorf("the notes are at %q, want %q", at.Notes, want)
	}
}

func TestDoFrom_showsTheConfigurationItReads(t *testing.T) {
	home := t.TempDir()
	path := writeConfiguration(t, home, "store = \"/tmp/somewhere/cairn.db\"\n")
	answer, err := DoFrom(ConfigShow, Home{Dir: home, Lookup: environment(map[string]string{
		config.NotesVariable: "/somewhere/of/my/own/notes",
	})}, Input{}, ByHand)
	if err != nil {
		t.Fatalf("DoFrom: %v", err)
	}
	if answer.Config == nil || answer.Config.File != path {
		t.Fatalf("Config = %+v, want the file %q", answer.Config, path)
	}
	if want := (Emplacement{At: "/tmp/somewhere/cairn.db", By: "the file"}); answer.Config.Store != want {
		t.Errorf("Store = %+v, want %+v", answer.Config.Store, want)
	}
	if want := (Emplacement{At: "/somewhere/of/my/own/notes", By: "the environment"}); answer.Config.Notes != want {
		t.Errorf("Notes = %+v, want %+v", answer.Config.Notes, want)
	}
}

func TestDoFrom_showsAConfigurationEveryOtherCommandRefuses(t *testing.T) {
	home := t.TempDir()
	writeConfiguration(t, home, "this is not toml: [[[\n")
	answer, err := DoFrom(ConfigShow, Home{Dir: home, Lookup: environment(nil)}, Input{}, ByHand)
	if err != nil {
		t.Fatalf("DoFrom: %v", err)
	}
	if answer.Config == nil || answer.Config.Problem == "" {
		t.Errorf("Config = %+v, want the problem said", answer.Config)
	}
	if _, err := DoFrom(TaskList, Home{Dir: home, Lookup: environment(nil)}, Input{}, ByHand); !errors.Is(err, config.ErrNotTOML) {
		t.Errorf("task list = %v, want ErrNotTOML", err)
	}
}

func TestDoFrom_refusesBeforeLocating(t *testing.T) {
	home := t.TempDir()
	writeConfiguration(t, home, "this is not toml: [[[\n")
	_, err := DoFrom(TaskShow, Home{Dir: home, Lookup: environment(nil)},
		Input{Given: map[string]string{"numbre": "1"}}, ByHand)
	if err == nil || errors.Is(err, config.ErrNotTOML) {
		t.Errorf("DoFrom = %v, want the unknown argument refused first", err)
	}
}

func TestHold_answersConfigShowFromTheHomeItWasOpenedAt(t *testing.T) {
	home := t.TempDir()
	path := writeConfiguration(t, home, "store = \"~/cairn.db\"\n")
	held, err := Hold(Home{Dir: home, Lookup: environment(nil)})
	if err != nil {
		t.Fatalf("Hold: %v", err)
	}
	defer func() { _ = held.Close() }()
	answer, err := held.Ask(ConfigShow, Input{})
	if err != nil {
		t.Fatalf("Ask: %v", err)
	}
	if answer.Config == nil || answer.Config.File != path {
		t.Errorf("Config = %+v, want the file %q", answer.Config, path)
	}
}

func TestHold_saysWhatToTypeAboutAStoreBehind(t *testing.T) {
	home := t.TempDir()
	path := behindStore(t)
	_, err := Hold(Home{Dir: home, Lookup: environment(map[string]string{config.StoreVariable: path})})
	if !errors.Is(err, ErrStoreBehind) || !strings.Contains(err.Error(), "cairn migration apply") {
		t.Errorf("Hold = %v, want the store behind and what to type", err)
	}
}

func TestRenderText_padsBothPlacesToTheLongerOfThem(t *testing.T) {
	for _, tc := range []struct {
		what         string
		store, notes string
	}{
		{"the notes path is longer", "/s/db", "/notes/are/longer"},
		{"the store path is longer", "/store/is/longer/db", "/n"},
		{"both the same length", "/same/len", "/same/len"},
	} {
		said := RenderText(ConfigShow, Answer{Config: &Config{
			File:  "/c.toml",
			Store: Emplacement{At: tc.store, By: "the file"},
			Notes: Emplacement{At: tc.notes, By: "nothing"},
		}})
		var at []int
		for _, line := range strings.Split(said, "\n") {
			if i := strings.Index(line, "named by"); i >= 0 {
				at = append(at, i)
			}
		}
		if len(at) != 2 {
			t.Fatalf("%s: %q names what named it %d times, want twice", tc.what, said, len(at))
		}
		if at[0] != at[1] {
			t.Errorf("%s: the second column starts at %d and %d:\n%s", tc.what, at[0], at[1], said)
		}
		wide := max(len(tc.store), len(tc.notes))
		if got := at[0] - len("store   "); got < wide {
			t.Errorf("%s: the column is %d wide and the longer path is %d:\n%s", tc.what, got, wide, said)
		}
	}
}

func TestRenderText_saysWhatItCouldNotMakeSenseOf(t *testing.T) {
	said := RenderText(ConfigShow, Answer{Config: &Config{File: "/c.toml", Problem: "not TOML"}})
	for _, want := range []string{"/c.toml", "cannot be made sense of: not TOML", "the other commands will refuse"} {
		if !strings.Contains(said, want) {
			t.Errorf("%q does not say %q", said, want)
		}
	}
	if strings.Contains(said, "named by") {
		t.Errorf("%q shows places for a file it could not make sense of", said)
	}
}
