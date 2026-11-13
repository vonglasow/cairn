package config

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pelletier/go-toml/v2"
)

func TestRead_aWorkspaceWrittenAsATableNamesItsNotes(t *testing.T) {
	read, err := readThis(t, "[workspaces]\ndefault = \"cairn\"\n\n[workspaces.cairn]\nfolder = \"/code/cairn/.cairn\"\nnotes = \"~/vaults/cairn\"\n")
	if err != nil {
		t.Fatal(err)
	}
	if got := read.Workspaces(); len(got) != 1 || got[0].Folder != "/code/cairn/.cairn" || !strings.HasSuffix(got[0].Notes, filepath.Join("vaults", "cairn")) || !filepath.IsAbs(got[0].Notes) {
		t.Errorf("Workspaces() = %+v, want cairn at its folder with its notes under the home", got)
	}
}

func TestRead_aWorkspaceWrittenTheShortWayNamesNoNotes(t *testing.T) {
	read, err := readThis(t, "[workspaces]\nwork = \"/w\"\n")
	if err != nil || len(read.Workspaces()) != 1 || read.Workspaces()[0].Notes != "" {
		t.Errorf("Read = %+v, %v; want work and no notes of its own", read.Workspaces(), err)
	}
}

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

func TestUpgraded_leavesAFileWithNoTopLevelPlaceAlone(t *testing.T) {
	up, err := Upgraded([]byte("plugin = \"~/p\"\n[workspaces]\nwork = \"/w\"\n"), "/home/me")
	if err != nil || up.Data != nil {
		t.Errorf("Upgraded = %q, %v; want nothing to do", up.Data, err)
	}
}

func TestUpgraded_makesMainOfTheTopLevelPlacesTheDefault(t *testing.T) {
	up, doc := upgradedOf(t, "# mine\nstore = \"~/data/cairn.db\"\nnotes = \"~/data/notes\"\nplugin = \"~/p\"\n\n[pages]\ntasks = true\n")
	spaces, _ := doc["workspaces"].(map[string]any)
	if spaces["default"] != "main" || spaces["main"] != "~/data" || doc["store"] != nil || doc["notes"] != nil || doc["plugin"] != "~/p" {
		t.Errorf("Upgraded wrote:\n%s\nwant main at ~/data the default, no top-level store or notes, the rest kept", up.Data)
	}
	if !strings.HasPrefix(string(up.Data), "# mine\n") || up.Name != "main" || up.Folder != "/home/me/data" || !up.Default {
		t.Errorf("Upgraded = %+v:\n%s", up, up.Data)
	}
}

func TestUpgraded_keepsTheNotesInUseWhenTheyAreNotInTheFolder(t *testing.T) {
	up, doc := upgradedOf(t, "store = \"~/data/cairn.db\"\n")
	spaces, _ := doc["workspaces"].(map[string]any)
	main, _ := spaces["main"].(map[string]any)
	if spaces["default"] != "main" || main["folder"] != "~/data" || main["notes"] != "~/.cairn/notes" {
		t.Errorf("Upgraded wrote:\n%s\nwant main at ~/data with the notes of ~/.cairn", up.Data)
	}
	if up.Notes != "/home/me/.cairn/notes" {
		t.Errorf("Upgraded said the notes %q", up.Notes)
	}
}

func TestUpgraded_onlyNotesMakesMainAtTheUsualFolder(t *testing.T) {
	up, doc := upgradedOf(t, "notes = \"~/vault\"\n")
	spaces, _ := doc["workspaces"].(map[string]any)
	main, _ := spaces["main"].(map[string]any)
	if main["folder"] != "~/.cairn" || main["notes"] != "~/vault" {
		t.Errorf("Upgraded wrote:\n%s\nwant main at ~/.cairn with the notes of ~/vault", up.Data)
	}
}

func TestUpgraded_joinsTheWorkspaceAlreadyAtThatFolder(t *testing.T) {
	up, doc := upgradedOf(t, "store = \"~/.cairn/cairn.db\"\n[workspaces]\nhome = \"/home/me/.cairn\"\n")
	spaces, _ := doc["workspaces"].(map[string]any)
	if spaces["default"] != "home" || spaces["main"] != nil || spaces["home"] != "/home/me/.cairn" || up.Name != "home" {
		t.Errorf("Upgraded wrote:\n%s\nwant home the default and no main", up.Data)
	}
}

func TestUpgraded_joiningAWorkspaceGivesItTheNotesInUse(t *testing.T) {
	up, doc := upgradedOf(t, "store = \"~/w/cairn.db\"\nnotes = \"~/vault\"\n[workspaces]\nwork = \"~/w\"\nplay = \"~/p\"\n")
	spaces, _ := doc["workspaces"].(map[string]any)
	work, _ := spaces["work"].(map[string]any)
	if spaces["default"] != "work" || work["folder"] != "~/w" || work["notes"] != "~/vault" || spaces["play"] != "~/p" {
		t.Errorf("Upgraded wrote:\n%s\nwant work with the notes of ~/vault, the default", up.Data)
	}
}

func TestUpgraded_aDefaultAlreadyNamedStaysTheDefault(t *testing.T) {
	up, doc := upgradedOf(t, "store = \"~/data/cairn.db\"\n[workspaces]\ndefault = \"work\"\nwork = \"~/w\"\n")
	spaces, _ := doc["workspaces"].(map[string]any)
	main, _ := spaces["main"].(map[string]any)
	if spaces["default"] != "work" || main["folder"] != "~/data" || up.Default {
		t.Errorf("Upgraded wrote:\n%s\nwant work still the default, and main beside it", up.Data)
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

func TestUpgraded_aWorkspaceNamingItsNotesKeepsThem(t *testing.T) {
	up, doc := upgradedOf(t, "store = \"~/w/cairn.db\"\nnotes = \"~/vault\"\n[workspaces]\n\n[workspaces.work]\nfolder = \"~/w\"\nnotes = \"~/own\"\n")
	spaces, _ := doc["workspaces"].(map[string]any)
	work, _ := spaces["work"].(map[string]any)
	if spaces["default"] != "work" || work["notes"] != "~/own" || up.Notes != "/home/me/own" {
		t.Errorf("Upgraded = %+v:\n%s\nwant work's own notes kept", up, up.Data)
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
