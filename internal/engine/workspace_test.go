package engine

import (
	"cairn/internal/engine/workspace"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cairn/internal/store/files"
)

func homeHolding(t *testing.T, said string, env map[string]string, folders ...string) Home {
	t.Helper()
	home := t.TempDir()
	if said != "" {
		if err := os.MkdirAll(filepath.Dir(files.FileIn(home)), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(files.FileIn(home), []byte(said), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	for _, folder := range folders {
		if err := os.MkdirAll(filepath.Join(home, folder), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	return Home{Dir: home, Lookup: func(key string) (string, bool) {
		said, ok := env[key]
		return strings.ReplaceAll(said, "~", home), ok
	}}
}

const twoWorkspaces = "[workspaces]\ndefault = \"work\"\nwork = \"~/w\"\nhome = \"~/h\"\n"

func TestLocate_aStoreNamedThatIsAWorkspacesTakesItsNotes(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, map[string]string{files.StoreVariable: "~/h/cairn.db"}, "w", "h")
	if at, err := Locate(h); err != nil || at.Notes != filepath.Join(h.Dir, "h", "notes") {
		t.Errorf("Locate = %+v, %v; want home's notes", at, err)
	}
	h = homeHolding(t, twoWorkspaces, map[string]string{files.NotesVariable: "~/n"}, "w", "h")
	h.Store = "~/h/cairn.db"
	if at, err := Locate(h); err != nil || at.Notes != filepath.Join(h.Dir, "n") {
		t.Errorf("Locate = %+v, %v; want the environment's notes before home's", at, err)
	}
}

func TestLocate_aStoreFlagNotAbsoluteIsRefused(t *testing.T) {
	h := homeHolding(t, "", nil)
	h.Store = "f/cairn.db"
	if _, err := Locate(h); !errors.Is(err, files.ErrNotAbsolute) || !strings.Contains(err.Error(), "--store") {
		t.Errorf("Locate = %v, want --store refused as not absolute", err)
	}
}

func TestLocate_bothFlagsAreNotUnderstood(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, nil, "w", "h")
	h.Workspace, h.Store = "home", "/f/cairn.db"
	_, err := Locate(h)
	if !errors.Is(err, errMisread) || !strings.Contains(err.Error(), "--workspace") || !strings.Contains(err.Error(), "--store") {
		t.Errorf("Locate = %v, want both flags not understood together", err)
	}
}

func TestLocate_aConfigurationThatCannotBeReadIsRefusedAsBefore(t *testing.T) {
	h := homeHolding(t, "[workspaces]\nwork = \"w\"\n", nil)
	if _, err := Locate(h); !errors.Is(err, files.ErrNotAbsolute) {
		t.Errorf("Locate = %v, want the folder refused as not absolute", err)
	}
	if _, err := WorkspaceList.Run(workedOn{Home: h}, Input{}); !errors.Is(err, files.ErrNotAbsolute) {
		t.Errorf("workspace list = %v, want the same refusal", err)
	}
}

func TestWorkspaceList_saysWhenNoneIsNamed(t *testing.T) {
	answer, err := WorkspaceList.Run(workedOn{Home: homeHolding(t, "", nil)}, Input{})
	if err != nil {
		t.Fatal(err)
	}
	if listed := answer.Workspaces.Workspaces; len(listed) != 0 {
		t.Errorf("it lists %+v", listed)
	}
}

func adding(name, folder string) Input {
	return Input{Given: map[string]string{"name": name, "folder": folder}}
}

func TestWorkspaceAdd_writesItAndMakesItsFolder(t *testing.T) {
	h := homeHolding(t, "plugin = \"~/plug\"\n", nil)
	answer, err := WorkspaceAdd.Run(workedOn{Home: h}, adding("play", "~/p"))
	if err != nil {
		t.Fatal(err)
	}
	read, _ := os.ReadFile(files.FileIn(h.Dir))
	if want := "plugin = \"~/plug\"\n\n[workspaces]\nplay = \"~/p\"\n"; string(read) != want {
		t.Errorf("the file reads %q, want %q", read, want)
	}
	if info, err := os.Stat(filepath.Join(h.Dir, "p", "notes")); err != nil || !info.IsDir() {
		t.Errorf("the notes folder: %v", err)
	}
	added := *answer.WorkspaceAdded
	if added.Name != "play" || added.Folder != filepath.Join(h.Dir, "p") || added.Store != filepath.Join(h.Dir, "p", "cairn.db") || added.Written != files.FileIn(h.Dir) {
		t.Errorf("added %+v", added)
	}
	if Of(WorkspaceAdd, answer).(workspace.Added) != added {
		t.Error("Of does not answer what was added")
	}
}

func TestWorkspaceAdd_writesTheColourItIsGiven(t *testing.T) {
	h := homeHolding(t, "", nil)
	in := adding("play", "~/p")
	in.Given["colour"] = "teal"
	if _, err := WorkspaceAdd.Run(workedOn{Home: h}, in); err != nil {
		t.Fatal(err)
	}
	read, _ := os.ReadFile(filepath.Join(h.Dir, "p", files.SettingsFile))
	if !strings.Contains(string(read), "[workspace]\ncolour = \"teal\"") {
		t.Errorf("the settings read %q", read)
	}
	in = adding("game", "~/g")
	in.Given["colour"] = "plaid"
	if _, err := WorkspaceAdd.Run(workedOn{Home: h}, in); err == nil || !strings.Contains(err.Error(), "plaid") {
		t.Errorf("a colour cairn has not is taken: %v", err)
	}
	if _, err := os.Stat(filepath.Join(h.Dir, "g")); err == nil {
		t.Error("a refused colour made the folder")
	}
}

func TestWorkspaceAdd_refusesWhatItCannotName(t *testing.T) {
	for _, c := range []struct {
		name, folder string
		want         error
		names        string
	}{
		{"Work", "~/p", workspace.ErrNotAWorkspaceName, `"Work"`},
		{"my_work", "~/p", workspace.ErrNotAWorkspaceName, `"my_work"`},
		{"", "~/p", workspace.ErrNotAWorkspaceName, `""`},
		{"default", "~/p", workspace.ErrNotAWorkspaceName, `"default"`},
		{"work", "~/p", workspace.ErrWorkspaceKnown, `"work"`},
		{"play", "p", files.ErrNotAbsolute, `"p"`},
	} {
		h := homeHolding(t, twoWorkspaces, nil)
		if _, err := WorkspaceAdd.Run(workedOn{Home: h}, adding(c.name, c.folder)); !errors.Is(err, c.want) || !strings.Contains(err.Error(), c.names) {
			t.Errorf("add %q %q = %v, want %v naming it", c.name, c.folder, err, c.want)
		}
		if read, _ := os.ReadFile(files.FileIn(h.Dir)); string(read) != twoWorkspaces {
			t.Errorf("add %q %q changed the file: %q", c.name, c.folder, read)
		}
		if _, err := os.Stat(filepath.Join(h.Dir, "p")); err == nil {
			t.Errorf("add %q %q made the folder", c.name, c.folder)
		}
	}
}

func TestWorkspaceAdd_refusesAFileItCannotChange(t *testing.T) {
	h := homeHolding(t, "workspaces = { work = \"/w\" }\n", nil)
	if _, err := WorkspaceAdd.Run(workedOn{Home: h}, adding("play", "~/p")); !errors.Is(err, files.ErrNotWritten) {
		t.Errorf("add = %v, want it refused", err)
	}
	if _, err := os.Stat(filepath.Join(h.Dir, "p")); err == nil {
		t.Error("it made the folder of a workspace it did not write")
	}
}

func TestWorkspaceAddAndUse_needAHome(t *testing.T) {
	missing := errors.New("no home")
	for _, op := range []Operation{WorkspaceAdd, WorkspaceUse} {
		if _, err := op.Run(workedOn{Home: Home{Missing: missing}}, adding("play", "/p")); !errors.Is(err, errStore) || !errors.Is(err, missing) {
			t.Errorf("%s = %v, want the missing home said", op.Name(), err)
		}
	}
}

func using(name string) Input { return Input{Given: map[string]string{"name": name}} }

func TestWorkspaceUse_makesItTheDefaultAndSaysWhatItWas(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, nil, "w", "h")
	answer, err := WorkspaceUse.Run(workedOn{Home: h}, using("home"))
	if err != nil {
		t.Fatal(err)
	}
	if *answer.WorkspaceUsed != (workspace.Used{Default: "home", Was: "work"}) {
		t.Errorf("used %+v", *answer.WorkspaceUsed)
	}
	read, _ := os.ReadFile(files.FileIn(h.Dir))
	if want := strings.Replace(twoWorkspaces, `default = "work"`, `default = "home"`, 1); string(read) != want {
		t.Errorf("the file reads %q, want %q", read, want)
	}
	if Of(WorkspaceUse, answer).(workspace.Used).Default != "home" {
		t.Error("Of does not answer the default")
	}
}

func TestWorkspaceUse_refusesAWorkspaceNotThere(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, nil, "w")
	if _, err := WorkspaceUse.Run(workedOn{Home: h}, using("play")); !errors.Is(err, workspace.ErrNoSuchWorkspace) {
		t.Errorf("use play = %v", err)
	}
	if _, err := WorkspaceUse.Run(workedOn{Home: h}, using("home")); !errors.Is(err, workspace.ErrWorkspaceNotThere) {
		t.Errorf("use home = %v", err)
	}
	if read, _ := os.ReadFile(files.FileIn(h.Dir)); string(read) != twoWorkspaces {
		t.Errorf("the file was changed: %q", read)
	}
	if _, err := WorkspaceUse.Run(workedOn{Home: homeHolding(t, "[workspaces]\n", nil)}, using("x")); !errors.Is(err, workspace.ErrNoSuchWorkspace) {
		t.Errorf("use x = %v", err)
	}
}

func TestWorkspaceUse_refusesAFileItCannotChange(t *testing.T) {
	h := homeHolding(t, "workspaces = { work = \"~/w\" }\n", nil, "w")
	if _, err := WorkspaceUse.Run(workedOn{Home: h}, using("work")); !errors.Is(err, files.ErrNotWritten) {
		t.Errorf("use = %v, want it refused", err)
	}
}

func TestConfigShow_saysTheWorkspacesAndWhatNamedEachPlace(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, nil, "w", "h")
	answer, err := ConfigShow.Run(workedOn{Home: h}, Input{})
	if err != nil {
		t.Fatal(err)
	}
	c := *answer.Config
	if c.Store != (workspace.Emplacement{At: filepath.Join(h.Dir, "w", "cairn.db"), By: `the default workspace, "work"`}) || c.Plugin.At != filepath.Join(h.Dir, "w", "claude-plugin") {
		t.Errorf("config %+v", c)
	}
	if len(c.Workspaces) != 2 || !c.Workspaces[1].Default {
		t.Errorf("workspaces %+v", c.Workspaces)
	}
	h.Workspace = "play"
	answer, err = ConfigShow.Run(workedOn{Home: h}, Input{})
	if err != nil || !strings.Contains(answer.Config.Problem, "play") {
		t.Errorf("config show = %+v, %v; want the workspace said as the problem", answer.Config, err)
	}
	h.Workspace, h.Store = "home", ""
	answer, _ = ConfigShow.Run(workedOn{Home: h}, Input{})
	if answer.Config.Store.By != "--workspace" {
		t.Errorf("store named by %q", answer.Config.Store.By)
	}
}
