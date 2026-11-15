package files

import (
	"errors"
	"strings"
	"testing"

	"github.com/pelletier/go-toml/v2"
)

func TestRead_aWorkspaceTableIsRefusedWithoutAFolderOrWithAKeyItDoesNotKnow(t *testing.T) {
	for _, said := range []string{
		"[workspaces.cairn]\nnotes = \"/n\"\n",
		"[workspaces.cairn]\nfolder = \"/f\"\ncolour = \"red\"\n",
		"[workspaces.cairn]\nfolder = \"/f\"\nnotes = \"n\"\n",
		"[workspaces]\ncairn = 3\n",
	} {
		if _, err := readThis(t, said); err == nil || !strings.Contains(err.Error(), "workspaces.cairn") {
			t.Errorf("Read(%q) = %v, want it refused naming workspaces.cairn", said, err)
		}
	}
}

func upgradedOf(t *testing.T, said string) (Upgrade, map[string]any) {
	t.Helper()
	up, err := Upgraded([]byte(said), "/home/me")
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := toml.Unmarshal(up.Data, &doc); err != nil {
		t.Fatalf("%v:\n%s", err, up.Data)
	}
	return up, doc
}

func TestUpgraded_joiningAWorkspaceGivesItTheNotesInUse(t *testing.T) {
	up, doc := upgradedOf(t, "store = \"~/w/cairn.db\"\nnotes = \"~/vault\"\n[workspaces]\nwork = \"~/w\"\nplay = \"~/p\"\n")
	spaces, _ := doc["workspaces"].(map[string]any)
	work, _ := spaces["work"].(map[string]any)
	if spaces["default"] != "work" || work["folder"] != "~/w" || work["notes"] != "~/vault" || spaces["play"] != "~/p" {
		t.Errorf("Upgraded wrote:\n%s\nwant work with the notes of ~/vault, the default", up.Data)
	}
}

func TestUpgraded_refusesWhatItCannotTurnIntoAWorkspace(t *testing.T) {
	for said, want := range map[string]error{
		"store = \"./cairn.db\"\n":      ErrNotAbsolute,
		"store = \"~/data/tasks.db\"\n": ErrNotAWorkspaceStore,
		"store = = \n":                  ErrNotTOML,
	} {
		if _, err := Upgraded([]byte(said), "/home/me"); !errors.Is(err, want) {
			t.Errorf("Upgraded(%q) = %v, want %v", said, err, want)
		}
	}
}

func TestUpgraded_aWorkspacePassedOverForTheDefaultKeepsItsOwnNotes(t *testing.T) {
	up, doc := upgradedOf(t, "store = \"~/w/cairn.db\"\nnotes = \"~/vault\"\n[workspaces]\ndefault = \"play\"\nwork = \"~/w\"\nplay = \"~/p\"\n")
	spaces, _ := doc["workspaces"].(map[string]any)
	if spaces["default"] != "play" || spaces["work"] != "~/w" || up.Name != "work" || up.Notes != "/home/me/w/notes" || up.Default {
		t.Errorf("Upgraded = %+v:\n%s\nwant work left as it was, play still the default", up, up.Data)
	}
}

func TestUpgraded_mainTakenElsewhereIsNotWrittenOver(t *testing.T) {
	up, doc := upgradedOf(t, "store = \"~/data/cairn.db\"\n[workspaces]\nmain = \"~/m\"\n")
	spaces, _ := doc["workspaces"].(map[string]any)
	renamed, _ := spaces["main-v0-11"].(map[string]any)
	if spaces["main"] != "~/m" || renamed["folder"] != "~/data" || spaces["default"] != "main-v0-11" || up.Name != "main-v0-11" {
		t.Errorf("Upgraded = %+v:\n%s\nwant main kept and main-v0-11 the default", up, up.Data)
	}
}
