package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cairn/internal/config"
)

func TestKeep_aServerKeepsTheDefaultItStartedOn(t *testing.T) {
	h := Keep(homeHolding(t, twoWorkspaces, nil, "w", "h"))
	if err := os.WriteFile(config.FileIn(h.Dir), []byte(strings.Replace(twoWorkspaces, `"work"`, `"home"`, 1)), 0o600); err != nil {
		t.Fatal(err)
	}
	at, err := Locate(h)
	if err != nil || at.Store != filepath.Join(h.Dir, "w", "cairn.db") || at.Notes != filepath.Join(h.Dir, "w", "notes") {
		t.Errorf("Locate = %+v, %v; want work's places, the default it started on", at, err)
	}
	answer, err := WorkspaceShow.Run(On{Home: h}, Input{})
	if err != nil {
		t.Fatal(err)
	}
	if shown := answer.WorkspaceShown; shown.Workspace != "work" || shown.Because != "it was the default of [workspaces] when this server started" || !strings.HasPrefix(shown.Switch, "not from here") {
		t.Errorf("shown %+v, want work, kept since the start, and how to switch", shown)
	}
}

func TestKeep_saysWhyAsItWasWhenTheServerStarted(t *testing.T) {
	for _, c := range []struct {
		name             string
		env              map[string]string
		workspace, store string
		said             string
		because          string
	}{
		{"the flag", nil, "home", "", twoWorkspaces, "--workspace, given when this server started"},
		{"the store flag", nil, "", "~/x/cairn.db", twoWorkspaces, "--store, given when this server started"},
		{"the environment", map[string]string{config.StoreVariable: "~/e/cairn.db"}, "", "", twoWorkspaces, "CAIRN_STORE, set when this server started"},
		{"the file", nil, "", "", "store = \"~/s/cairn.db\"\n", "the file named it when this server started"},
		{"nothing", nil, "", "", "", "nothing named it when this server started: it is where cairn keeps a store by default"},
	} {
		h := homeHolding(t, c.said, c.env, "w", "h")
		h.Workspace, h.Store = c.workspace, c.store
		kept := Keep(h)
		if kept.Kept == nil || kept.Kept.Because != c.because {
			t.Errorf("%s: kept %+v, want because %q", c.name, kept.Kept, c.because)
		}
	}
}

func TestKeep_whatCannotBeSettledIsFoundAtEachCall(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, nil, "h")
	if kept := Keep(h); kept.Kept != nil {
		t.Errorf("kept %+v from a default whose folder is not there", kept.Kept)
	}
	if kept := Keep(Home{Missing: ErrStore}); kept.Kept != nil {
		t.Errorf("kept %+v with no home", kept.Kept)
	}
}

func TestKeep_isCarriedWhole(t *testing.T) {
	h := Keep(homeHolding(t, twoWorkspaces, nil, "w", "h"))
	said, err := json.Marshal(h.Kept)
	if err != nil {
		t.Fatal(err)
	}
	var back Started
	if err := json.Unmarshal(said, &back); err != nil || back != *h.Kept {
		t.Errorf("carried %s, read back %+v, %v", said, back, err)
	}
}

func TestWorkspaceShow_saysTheTimeAndWhereItIsSaid(t *testing.T) {
	for _, c := range []struct {
		name, said, settings, mode, from string
	}{
		{"by default", twoWorkspaces, "", "on", "by default"},
		{"the global file", "[pages]\ntime = \"hidden\"\n" + twoWorkspaces, "", "hidden", "from the global configuration"},
		{"the store's settings", "[pages]\ntime = \"hidden\"\n" + twoWorkspaces, "[pages]\ntime = \"off\"\n", "off", "set in the store's settings"},
	} {
		h := homeHolding(t, c.said, nil, "w", "h")
		if c.settings != "" {
			if err := os.WriteFile(filepath.Join(h.Dir, "w", config.SettingsFile), []byte(c.settings), 0o600); err != nil {
				t.Fatal(err)
			}
		}
		answer, err := WorkspaceShow.Run(On{Home: h}, Input{})
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		shown := answer.WorkspaceShown
		if shown.Time != c.mode || shown.TimeFrom != c.from || shown.Switch != "" {
			t.Errorf("%s: shown %+v, want time %q %s and nothing on switching", c.name, shown, c.mode, c.from)
		}
		if said := RenderText(WorkspaceShow, answer); !strings.Contains(said, "time       "+c.mode+", "+c.from+"\n") || strings.Contains(said, "switch") {
			t.Errorf("%s: it says\n%s", c.name, said)
		}
	}
}

func TestWorkspaceShow_refusesSettingsItCannotRead(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, nil, "w", "h")
	if err := os.WriteFile(filepath.Join(h.Dir, "w", config.SettingsFile), []byte("not toml ["), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := WorkspaceShow.Run(On{Home: h}, Input{}); err == nil {
		t.Error("workspace show answered over settings that are not TOML")
	}
}

func TestWorkspaceShow_aServerSaysSwitchingIsNotDoneThere(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, nil, "w", "h")
	h.Workspace = "home"
	answer, err := WorkspaceShow.Run(On{Home: Keep(h)}, Input{})
	if err != nil {
		t.Fatal(err)
	}
	said := RenderText(WorkspaceShow, answer)
	for _, want := range []string{"workspace  home\n", "because    --workspace, given when this server started\n", "switch     not from here: this server keeps its store; another workspace is another server, which the person starts\n"} {
		if !strings.Contains(said, want) {
			t.Errorf("it says\n%s\nwant %q", said, want)
		}
	}
}
