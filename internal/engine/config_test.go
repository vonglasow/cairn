package engine

import (
	"cairn/internal/engine/workspace"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cairn/internal/store/files"
)

func TestConfigInit_writesTheExampleAndNeverOverAFile(t *testing.T) {
	home := t.TempDir()
	answer, err := ConfigInit.Run(workedOn{Home: Home{Dir: home}}, Input{})
	if err != nil {
		t.Fatal(err)
	}
	written, err := os.ReadFile(files.FileIn(home))
	if err != nil || string(written) != files.Example {
		t.Fatalf("it wrote %q, %v", written, err)
	}
	if answer.Written == nil || *answer.Written != files.FileIn(home) {
		t.Errorf("it says it wrote %v", answer.Written)
	}
	if said := RenderText(ConfigInit, answer); !strings.Contains(said, files.FileIn(home)) {
		t.Errorf("it says %q", said)
	}
	if _, err := ConfigInit.Run(workedOn{Home: Home{Dir: home}}, Input{}); !errors.Is(err, workspace.ErrConfigAlreadyThere) {
		t.Errorf("writing it again = %v, want ErrConfigAlreadyThere", err)
	}
	if ConfigInit.Level != write || ConfigInit.Needs != aConfiguration {
		t.Errorf("config init is level %v and needs %v", ConfigInit.Level, ConfigInit.Needs)
	}
}

func shownWith(t *testing.T, content string) workspace.Config {
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
	answer, err := ConfigShow.Run(workedOn{Home: Home{Dir: home, Lookup: func(string) (string, bool) { return "", false }}}, Input{})
	if err != nil || answer.Config == nil {
		t.Fatalf("config show: %v", err)
	}
	return *answer.Config
}

func TestConfigShow_answersWhatTheFileSays(t *testing.T) {
	c := shownWith(t, "[pages]\ntime = false\n[archive]\nstart = 2026-10-21\ndays = 7\n[focus]\nsession = \"1h\"\n[time]\ndeleted = \"taken\"\n[window]\nkeep_sizes = true\n")
	if *c.Pages != (workspace.PagesConfig{Tasks: true, Objectives: true, Time: "hidden"}) || *c.Archive != (workspace.Archive{Start: "2026-10-21", Days: 7}) ||
		c.Focus.Session != "1h" || c.Time.Deleted != "taken" || !c.Window.KeepSizes {
		t.Errorf("config show answers %+v %+v %+v %+v %+v", c.Pages, c.Archive, c.Focus, c.Time, c.Window)
	}
}

func TestConfigText_readsBackAsTheSameConfiguration(t *testing.T) {
	first := shownWith(t, fmt.Sprintf("[workspaces]\ndefault = \"main\"\n\n[workspaces.main]\nfolder = %q\nnotes = \"/x/notes\"\n\n[pages]\nobjectives = false\n", t.TempDir())+"[archive]\nstart = 2026-10-21\ndays = 10\n[focus]\nsession = \"45m\"\n[time]\ndeleted = \"taken\"\n[window]\nkeep_sizes = true\n")
	if first.Notes.At != "/x/notes" {
		t.Fatalf("the notes are at %q, want those of main", first.Notes.At)
	}
	again := shownWith(t, workspace.ConfigText(first))
	if again.Problem != "" {
		t.Fatalf("what config show says is refused as a configuration: %s\n%s", again.Problem, workspace.ConfigText(first))
	}
	if again.Store.At != first.Store.At || again.Notes.At != first.Notes.At || again.Plugin.At != first.Plugin.At ||
		*again.Pages != *first.Pages || *again.Archive != *first.Archive || *again.Focus != *first.Focus ||
		*again.Time != *first.Time || *again.Window != *first.Window {
		t.Errorf("read back, it says:\n%s\nwhere it said:\n%s", workspace.ConfigText(again), workspace.ConfigText(first))
	}
}
