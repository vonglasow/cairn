package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cairn/internal/config"
)

func TestLocate_theProjectHoldingAWorkspacesFolderNamesIt(t *testing.T) {
	h := homeHolding(t, "[workspaces]\nhome = \"~/.cairn\"\ncairn = \"~/code/cairn/.cairn\"\n", nil, ".cairn", "code/cairn/.cairn", "code/cairn/internal")
	h.Here = filepath.Join(h.Dir, "code", "cairn", "internal")
	at, err := Locate(h)
	if err != nil || at.Store != filepath.Join(h.Dir, "code", "cairn", ".cairn", "cairn.db") {
		t.Errorf("Locate = %+v, %v; want the store of cairn", at, err)
	}
}

func TestLocate_theNearestFolderWinsAndTheDirectoryComesBeforeTheDefault(t *testing.T) {
	h := homeHolding(t, "[workspaces]\ndefault = \"code\"\ncode = \"~/code/.cairn\"\ncairn = \"~/code/cairn/.cairn\"\n", nil, "code/.cairn", "code/cairn/.cairn")
	h.Here = filepath.Join(h.Dir, "code", "cairn")
	_, a, err := h.settled()
	if err != nil || a.workspace != "cairn" || a.why.Rung != rungDirectory || len(a.why.Over) != 1 || a.why.Over[0].Names != "code" {
		t.Errorf("settled = %+v, %v; want cairn, chosen by the directory over the default", a, err)
	}
}

func TestLocate_insideAFolderThatIsNoProjectsItsWorkspaceAnswers(t *testing.T) {
	h := homeHolding(t, "[workspaces]\ndefault = \"work\"\nwork = \"~/c/work\"\nhome = \"~/c/home\"\n", nil, "c/work", "c/home/notes")
	h.Here = filepath.Join(h.Dir, "c", "home", "notes")
	if at, err := Locate(h); err != nil || at.Store != filepath.Join(h.Dir, "c", "home", "cairn.db") {
		t.Errorf("Locate = %+v, %v; want home's store", at, err)
	}
	h.Here = filepath.Join(h.Dir, "c")
	if at, err := Locate(h); err != nil || at.Store != filepath.Join(h.Dir, "c", "work", "cairn.db") {
		t.Errorf("Locate = %+v, %v; want the default's store beside the folders", at, err)
	}
}

func TestLocate_aFolderInTheHomeAnswersOnlyFromInsideIt(t *testing.T) {
	h := homeHolding(t, "[workspaces]\ndefault = \"work\"\nhome = \"~/.cairn\"\nwork = \"~/w\"\n", nil, ".cairn/notes", "w", "letters")
	h.Here = filepath.Join(h.Dir, "letters")
	if at, err := Locate(h); err != nil || at.Store != filepath.Join(h.Dir, "w", "cairn.db") {
		t.Errorf("Locate = %+v, %v; want the default's store", at, err)
	}
	h.Here = filepath.Join(h.Dir, ".cairn", "notes")
	if at, err := Locate(h); err != nil || at.Store != filepath.Join(h.Dir, ".cairn", "cairn.db") {
		t.Errorf("Locate = %+v, %v; want home's store from inside its folder", at, err)
	}
}

func TestLocate_aDirectoryUnderAFolderGoneNamesNothing(t *testing.T) {
	h := homeHolding(t, "[workspaces]\ndefault = \"work\"\nwork = \"~/w\"\ncairn = \"~/code/cairn/.cairn\"\n", nil, "w", "code/cairn")
	h.Here = filepath.Join(h.Dir, "code", "cairn")
	if at, err := Locate(h); err != nil || at.Store != filepath.Join(h.Dir, "w", "cairn.db") {
		t.Errorf("Locate = %+v, %v; want the default's store", at, err)
	}
}

func TestLocate_aWorkspaceNamingItsNotesKeepsThemThere(t *testing.T) {
	h := homeHolding(t, "[workspaces]\ndefault = \"cairn\"\n\n[workspaces.cairn]\nfolder = \"~/c\"\nnotes = \"~/vault\"\n", nil, "c")
	if at, err := Locate(h); err != nil || at.Notes != filepath.Join(h.Dir, "vault") || at.Store != filepath.Join(h.Dir, "c", "cairn.db") {
		t.Errorf("Locate = %+v, %v; want the store in c and the notes in the vault", at, err)
	}
}

func TestLocate_noWorkspaceChosenIsMainInTheUsualFolder(t *testing.T) {
	h := homeHolding(t, "", nil)
	_, a, err := h.settled()
	if err != nil || a.workspace != config.UpgradedMain || a.folder != filepath.Join(h.Dir, ".cairn") {
		t.Errorf("settled = %+v, %v; want main at ~/.cairn", a, err)
	}
}

func TestUpgradeConfiguration_turnsTheTopLevelPlacesIntoMainOnceAndSaysSo(t *testing.T) {
	h := homeHolding(t, "store = \"~/data/cairn.db\"\n", nil)
	said, err := UpgradeConfiguration(h)
	if err != nil || !strings.Contains(said, `"main"`) || !strings.Contains(said, config.FileIn(h.Dir)+".v0.11.0") {
		t.Errorf("UpgradeConfiguration = %q, %v; want it saying main and where the old file is", said, err)
	}
	if kept, err := os.ReadFile(config.FileIn(h.Dir) + ".v0.11.0"); err != nil || string(kept) != "store = \"~/data/cairn.db\"\n" {
		t.Errorf("the old file kept reads %q, %v", kept, err)
	}
	if at, err := Locate(h); err != nil || at.Store != filepath.Join(h.Dir, "data", "cairn.db") || at.Notes != filepath.Join(h.Dir, ".cairn", "notes") {
		t.Errorf("Locate = %+v, %v; want the places as they were", at, err)
	}
	if again, err := UpgradeConfiguration(h); err != nil || again != "" {
		t.Errorf("UpgradeConfiguration again = %q, %v; want nothing said", again, err)
	}
}

func TestUpgradeConfiguration_aStoreNoWorkspaceCanHoldStaysWhereItIs(t *testing.T) {
	h := homeHolding(t, "store = \"~/data/tasks.db\"\n", nil)
	if said, err := UpgradeConfiguration(h); err == nil || said != "" {
		t.Errorf("UpgradeConfiguration = %q, %v; want it refused", said, err)
	}
	if at, err := Locate(h); err != nil || at.Store != filepath.Join(h.Dir, "data", "tasks.db") {
		t.Errorf("Locate = %+v, %v; want the store the file names, still", at, err)
	}
}

func TestLocate_upgradesTheConfigurationOnItsOwn(t *testing.T) {
	h := homeHolding(t, "notes = \"~/vault\"\n", nil)
	if _, err := Locate(h); err != nil {
		t.Fatal(err)
	}
	if said, _ := os.ReadFile(config.FileIn(h.Dir)); strings.Contains(string(said), "\nnotes = \"~/vault\"\n[") || strings.HasPrefix(string(said), "notes") {
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
	if err := os.WriteFile(config.FileIn(h.Dir), []byte("[workspaces]\ndefault = \"cairn\"\n\n[workspaces.cairn]\nfolder = \"~/c\"\nnotes = \"~/vault\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	return h
}

func TestNotesToMove_saysWhereTheNotesWereAndWhereTheyAreNamed(t *testing.T) {
	h := notesMovedAway(t)
	move, found, err := NotesToMove(h)
	if err != nil || !found || move.Workspace != "cairn" || move.From != filepath.Join(h.Dir, "c", "notes") || move.To != filepath.Join(h.Dir, "vault") || move.Blocked {
		t.Fatalf("NotesToMove = %+v, %v, %v", move, found, err)
	}
	if err := MoveNotes(h, move); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(h.Dir, "vault", "1", "summary.md")); err != nil {
		t.Error(err)
	}
	if _, err := os.Stat(filepath.Join(h.Dir, "c", "notes")); !os.IsNotExist(err) {
		t.Errorf("the old notes are still there: %v", err)
	}
	if _, found, err := NotesToMove(h); err != nil || found {
		t.Errorf("NotesToMove once moved = %v, %v; want nothing", found, err)
	}
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

func TestNotesToMove_anEmptyTargetIsTakenAndNotesGoneAreForgotten(t *testing.T) {
	h := notesMovedAway(t)
	if err := os.MkdirAll(filepath.Join(h.Dir, "vault"), 0o700); err != nil {
		t.Fatal(err)
	}
	move, found, err := NotesToMove(h)
	if err != nil || !found || move.Blocked {
		t.Fatalf("NotesToMove = %+v, %v, %v; want it free to move", move, found, err)
	}
	if err := os.RemoveAll(filepath.Join(h.Dir, "c", "notes")); err != nil {
		t.Fatal(err)
	}
	if _, found, err := NotesToMove(h); err != nil || found {
		t.Errorf("NotesToMove with the old notes gone = %v, %v; want nothing", found, err)
	}
}

func TestNotesToMove_notesTheEnvironmentNamesAreNotTheWorkspaces(t *testing.T) {
	h := notesMovedAway(t)
	h.Lookup = func(key string) (string, bool) {
		if key == config.NotesVariable {
			return filepath.Join(h.Dir, "elsewhere"), true
		}
		return "", false
	}
	if _, found, err := NotesToMove(h); err != nil || found {
		t.Errorf("NotesToMove = %v, %v; want nothing under CAIRN_NOTES", found, err)
	}
}

func TestWorkspaceShow_mainNobodyConfiguredIsSaidAndConfiguresNothing(t *testing.T) {
	answer, err := WorkspaceShow.Run(On{Home: homeHolding(t, "", nil)}, Input{})
	if err != nil {
		t.Fatal(err)
	}
	if shown := answer.WorkspaceShown; shown.Workspace != config.UpgradedMain || !shown.Implicit || shown.Configured() != "" {
		t.Errorf("shown %+v, want main, said implicit, configuring nothing", shown)
	}
	answer, err = WorkspaceShow.Run(On{Home: homeHolding(t, twoWorkspaces, nil, "w", "h")}, Input{})
	if err != nil || answer.WorkspaceShown.Implicit || answer.WorkspaceShown.Configured() != "work" {
		t.Errorf("shown %+v, %v; want work, configured", answer.WorkspaceShown, err)
	}
}

func TestLocate_aStoreNoWorkspaceNamesTakesTheNotesOfMain(t *testing.T) {
	h := homeHolding(t, "[workspaces]\ndefault = \"work\"\nwork = \"~/w\"\n\n[workspaces.main]\nfolder = \"~/m\"\nnotes = \"~/vault\"\n", map[string]string{config.StoreVariable: "~/e/cairn.db"}, "w", "m")
	if at, err := Locate(h); err != nil || at.Notes != filepath.Join(h.Dir, "vault") {
		t.Errorf("Locate = %+v, %v; want the notes of main", at, err)
	}
}

func TestLocate_aFolderGoneIsPassedOverForTheNextThatHoldsTheDirectory(t *testing.T) {
	h := homeHolding(t, "[workspaces]\na = \"~/gone\"\nb = \"~/b\"\n", nil, "b/src")
	h.Here = filepath.Join(h.Dir, "b", "src")
	if at, err := Locate(h); err != nil || at.Store != filepath.Join(h.Dir, "b", "cairn.db") {
		t.Errorf("Locate = %+v, %v; want b's store", at, err)
	}
}

func TestLocate_twoFoldersAsNearTheFirstByNameWins(t *testing.T) {
	h := homeHolding(t, "[workspaces]\na = \"~/q/.cairn\"\nb = \"~/q\"\n", nil, "q/.cairn", "q/src")
	h.Here = filepath.Join(h.Dir, "q", "src")
	if at, err := Locate(h); err != nil || at.Store != filepath.Join(h.Dir, "q", ".cairn", "cairn.db") {
		t.Errorf("Locate = %+v, %v; want a's store", at, err)
	}
}

func TestWithin_saysWhetherADirectoryIsUnderAFolder(t *testing.T) {
	for _, c := range []struct {
		dir, root string
		want      bool
	}{{"/a/b", "/a", true}, {"/a", "/a", true}, {"/a", "/a/b", false}, {"/c", "/a/b", false}, {"/a/bc", "/a/b", false}} {
		if got := within(c.dir, c.root); got != c.want {
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
	if got := fileRead(t, config.FileIn(h.Dir)); got != "[pages]\ntime = \"off\"\n" {
		t.Errorf("the file became %q", got)
	}
}
