package engine

import (
	"cairn/internal/engine/workspace"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cairn/internal/store/files"
)

func TestLocate_theProjectHoldingAWorkspacesFolderNamesIt(t *testing.T) {
	h := homeHolding(t, "[workspaces]\nhome = \"~/.cairn\"\ncairn = \"~/code/cairn/.cairn\"\n", nil, ".cairn", "code/cairn/.cairn", "code/cairn/internal")
	h.Here = filepath.Join(h.Dir, "code", "cairn", "internal")
	at, err := Locate(h)
	if err != nil || at.Store != filepath.Join(h.Dir, "code", "cairn", ".cairn", "cairn.db") {
		t.Errorf("Locate = %+v, %v; want the store of cairn", at, err)
	}
}

func TestLocate_aDirectoryUnderAFolderGoneNamesNothing(t *testing.T) {
	h := homeHolding(t, "[workspaces]\ndefault = \"work\"\nwork = \"~/w\"\ncairn = \"~/code/cairn/.cairn\"\n", nil, "w", "code/cairn")
	h.Here = filepath.Join(h.Dir, "code", "cairn")
	if at, err := Locate(h); err != nil || at.Store != filepath.Join(h.Dir, "w", "cairn.db") {
		t.Errorf("Locate = %+v, %v; want the default's store", at, err)
	}
}

func TestLocate_upgradesTheConfigurationOnItsOwn(t *testing.T) {
	h := homeHolding(t, "notes = \"~/vault\"\n", nil)
	if _, err := Locate(h); err != nil {
		t.Fatal(err)
	}
	if said, _ := os.ReadFile(files.FileIn(h.Dir)); strings.Contains(string(said), "\nnotes = \"~/vault\"\n[") || strings.HasPrefix(string(said), "notes") {
		t.Errorf("the configuration still names its notes at the top:\n%s", said)
	}
}

func notesMovedAway(t *testing.T) Home {
	t.Helper()
	h := homeHolding(t, "[workspaces]\ndefault = \"cairn\"\ncairn = \"~/c\"\n", nil, "c/notes/1")
	if err := os.WriteFile(filepath.Join(h.Dir, "c", "notes", "1", "summary.md"), []byte("x\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, found, err := NotesToMove(h); err != nil || found {
		t.Fatalf("NotesToMove on the notes where they are = %v, %v", found, err)
	}
	if err := os.WriteFile(files.FileIn(h.Dir), []byte("[workspaces]\ndefault = \"cairn\"\n\n[workspaces.cairn]\nfolder = \"~/c\"\nnotes = \"~/vault\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	return h
}

func TestNotesToMove_aTargetHoldingSomethingIsBlocked(t *testing.T) {
	h := notesMovedAway(t)
	if err := os.MkdirAll(filepath.Join(h.Dir, "vault"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(h.Dir, "vault", "Inbox.md"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	move, found, err := NotesToMove(h)
	if err != nil || !found || !move.Blocked {
		t.Fatalf("NotesToMove = %+v, %v, %v; want it blocked", move, found, err)
	}
	if err := MoveNotes(h, move); err == nil {
		t.Error("MoveNotes onto a folder holding something went through")
	}
}

func TestNotesToMove_notesTheEnvironmentNamesAreNotTheWorkspaces(t *testing.T) {
	h := notesMovedAway(t)
	h.Lookup = func(key string) (string, bool) {
		if key == files.NotesVariable {
			return filepath.Join(h.Dir, "elsewhere"), true
		}
		return "", false
	}
	if _, found, err := NotesToMove(h); err != nil || found {
		t.Errorf("NotesToMove = %v, %v; want nothing under CAIRN_NOTES", found, err)
	}
}

func TestWorkspaceShow_mainNobodyConfiguredIsSaidAndConfiguresNothing(t *testing.T) {
	answer, err := WorkspaceShow.Run(workedOn{Home: homeHolding(t, "", nil)}, Input{})
	if err != nil {
		t.Fatal(err)
	}
	if shown := answer.WorkspaceShown; shown.Workspace != files.UpgradedMain || !shown.Implicit || shown.Configured() != "" {
		t.Errorf("shown %+v, want main, said implicit, configuring nothing", shown)
	}
	answer, err = WorkspaceShow.Run(workedOn{Home: homeHolding(t, twoWorkspaces, nil, "w", "h")}, Input{})
	if err != nil || answer.WorkspaceShown.Implicit || answer.WorkspaceShown.Configured() != "work" {
		t.Errorf("shown %+v, %v; want work, configured", answer.WorkspaceShown, err)
	}
}

func TestLocate_aStoreNoWorkspaceNamesTakesTheNotesOfMain(t *testing.T) {
	h := homeHolding(t, "[workspaces]\ndefault = \"work\"\nwork = \"~/w\"\n\n[workspaces.main]\nfolder = \"~/m\"\nnotes = \"~/vault\"\n", map[string]string{files.StoreVariable: "~/e/cairn.db"}, "w", "m")
	if at, err := Locate(h); err != nil || at.Notes != filepath.Join(h.Dir, "vault") {
		t.Errorf("Locate = %+v, %v; want the notes of main", at, err)
	}
}

func TestWithin_saysWhetherADirectoryIsUnderAFolder(t *testing.T) {
	for _, c := range []struct {
		dir, root string
		want      bool
	}{{"/a/b", "/a", true}, {"/a", "/a", true}, {"/a", "/a/b", false}, {"/c", "/a/b", false}, {"/a/bc", "/a/b", false}} {
		if got := workspace.Within(c.dir, c.root); got != c.want {
			t.Errorf("within(%q, %q) = %v, want %v", c.dir, c.root, got, c.want)
		}
	}
}

func TestMoveNotes_takesAnEmptyFolderAndSaysWhatStoppedIt(t *testing.T) {
	h := notesMovedAway(t)
	if err := os.MkdirAll(filepath.Join(h.Dir, "vault"), 0o700); err != nil {
		t.Fatal(err)
	}
	move, _, _ := NotesToMove(h)
	if err := MoveNotes(h, move); err != nil {
		t.Fatalf("MoveNotes onto an empty folder = %v", err)
	}
	if _, err := os.Stat(filepath.Join(h.Dir, "vault", "1", "summary.md")); err != nil {
		t.Error(err)
	}
	move.From, move.To = filepath.Join(h.Dir, "gone"), filepath.Join(h.Dir, "else")
	if err := MoveNotes(h, move); err == nil {
		t.Error("MoveNotes from a folder gone went through")
	}
}

func TestConfigSet_aServerKeptOnMainNobodyNamedSetsTheGlobalFile(t *testing.T) {
	h := Keep(homeHolding(t, "", nil))
	if _, err := setting(h, ConfigSet, map[string]string{"key": "pages.time", "value": "off"}); err != nil {
		t.Fatal(err)
	}
	if got := fileRead(t, files.FileIn(h.Dir)); got != "[pages]\ntime = \"off\"\n" {
		t.Errorf("the file became %q", got)
	}
}
