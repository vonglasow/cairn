package engine

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cairn/internal/config"
)

func homeHolding(t *testing.T, said string, env map[string]string, folders ...string) Home {
	t.Helper()
	home := t.TempDir()
	if said != "" {
		if err := os.MkdirAll(filepath.Dir(config.FileIn(home)), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(config.FileIn(home), []byte(said), 0o600); err != nil {
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

func TestLocate_theDefaultWorkspaceGivesBothPlaces(t *testing.T) {
	h := homeHolding(t, "store = \"~/s/cairn.db\"\n"+twoWorkspaces, nil, "w", "h")
	at, err := Locate(h)
	if err != nil || at.Store != filepath.Join(h.Dir, "w", "cairn.db") || at.Notes != filepath.Join(h.Dir, "w", "notes") {
		t.Errorf("Locate = %+v, %v; want the store and notes of work, before the file's store", at, err)
	}
}

func TestLocate_noDefaultLeavesThePlacesAsTheFileSays(t *testing.T) {
	h := homeHolding(t, "store = \"~/s/cairn.db\"\n[workspaces]\nwork = \"~/w\"\n", nil, "w")
	if at, err := Locate(h); err != nil || at.Store != filepath.Join(h.Dir, "s", "cairn.db") {
		t.Errorf("Locate = %+v, %v; want the file's store", at, err)
	}
}

func TestLocate_aStoreTheEnvironmentNamesNeverTakesTheDefaultsNotes(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, map[string]string{config.StoreVariable: "~/e/cairn.db"}, "w", "h")
	at, err := Locate(h)
	if err != nil || at.Store != filepath.Join(h.Dir, "e", "cairn.db") || at.Notes != filepath.Join(h.Dir, ".cairn", "notes") {
		t.Errorf("Locate = %+v, %v; want the environment's store and the notes as before workspaces", at, err)
	}
}

func TestLocate_aStoreNamedThatIsAWorkspacesTakesItsNotes(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, map[string]string{config.StoreVariable: "~/h/cairn.db"}, "w", "h")
	if at, err := Locate(h); err != nil || at.Notes != filepath.Join(h.Dir, "h", "notes") {
		t.Errorf("Locate = %+v, %v; want home's notes", at, err)
	}
	h = homeHolding(t, twoWorkspaces, map[string]string{config.NotesVariable: "~/n"}, "w", "h")
	h.Store = "~/h/cairn.db"
	if at, err := Locate(h); err != nil || at.Notes != filepath.Join(h.Dir, "n") {
		t.Errorf("Locate = %+v, %v; want the environment's notes before home's", at, err)
	}
}

func TestLocate_theDefaultsStoreWithTheEnvironmentsNotes(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, map[string]string{config.NotesVariable: "~/n"}, "w", "h")
	if at, err := Locate(h); err != nil || at.Store != filepath.Join(h.Dir, "w", "cairn.db") || at.Notes != filepath.Join(h.Dir, "n") {
		t.Errorf("Locate = %+v, %v; want work's store and the environment's notes", at, err)
	}
}

func TestLocate_theEnvironmentNamingBothNeedsNoFolderOfTheDefault(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, map[string]string{config.StoreVariable: "~/e/cairn.db", config.NotesVariable: "~/n"})
	if at, err := Locate(h); err != nil || at.Notes != filepath.Join(h.Dir, "n") {
		t.Errorf("Locate = %+v, %v; want the environment's places, the default unused", at, err)
	}
}

func TestLocate_theWorkspaceFlagComesBeforeTheEnvironment(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, map[string]string{config.StoreVariable: "~/e/cairn.db", config.NotesVariable: "~/n"}, "w", "h")
	h.Workspace = "home"
	at, err := Locate(h)
	if err != nil || at.Store != filepath.Join(h.Dir, "h", "cairn.db") || at.Notes != filepath.Join(h.Dir, "h", "notes") {
		t.Errorf("Locate = %+v, %v; want home's places", at, err)
	}
}

func TestLocate_theStoreFlagComesBeforeTheEnvironmentForTheStoreAlone(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, map[string]string{config.StoreVariable: "~/e/cairn.db"}, "w", "h")
	h.Store = "~/f/cairn.db"
	at, err := Locate(h)
	if err != nil || at.Store != filepath.Join(h.Dir, "f", "cairn.db") || at.Notes != filepath.Join(h.Dir, ".cairn", "notes") {
		t.Errorf("Locate = %+v, %v; want the flag's store and the notes as before workspaces", at, err)
	}
}

func TestLocate_aStoreFlagNotAbsoluteIsRefused(t *testing.T) {
	h := homeHolding(t, "", nil)
	h.Store = "f/cairn.db"
	if _, err := Locate(h); !errors.Is(err, config.ErrNotAbsolute) || !strings.Contains(err.Error(), "--store") {
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

func TestLocate_anUnknownWorkspaceIsRefusedWithTheOnesThereAre(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, nil, "w", "h")
	h.Workspace = "play"
	_, err := Locate(h)
	if !errors.Is(err, ErrNoSuchWorkspace) || !strings.Contains(err.Error(), `"play"`) || !strings.Contains(err.Error(), "home, work") {
		t.Errorf("Locate = %v, want play refused, naming home and work", err)
	}
}

func TestLocate_anUnknownWorkspaceWithNoneNamedSaysSo(t *testing.T) {
	h := homeHolding(t, "", nil)
	h.Workspace = "play"
	if _, err := Locate(h); !errors.Is(err, ErrNoSuchWorkspace) || !strings.Contains(err.Error(), "none is named") {
		t.Errorf("Locate = %v, want play refused, saying none is named", err)
	}
}

func TestLocate_aDefaultNamingNoWorkspaceIsRefused(t *testing.T) {
	h := homeHolding(t, "[workspaces]\ndefault = \"play\"\nwork = \"~/w\"\n", nil, "w")
	if _, err := Locate(h); !errors.Is(err, ErrNoSuchWorkspace) || !strings.Contains(err.Error(), `"play"`) {
		t.Errorf("Locate = %v, want the default play refused", err)
	}
}

func TestLocate_aWorkspaceWhoseFolderIsNotThereIsRefusedNamingIt(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, nil, "w")
	h.Workspace = "home"
	if _, err := Locate(h); !errors.Is(err, ErrWorkspaceNotThere) || !strings.Contains(err.Error(), filepath.Join(h.Dir, "h")) {
		t.Errorf("Locate = %v, want home refused, naming its folder", err)
	}
}

func TestLocate_aConfigurationThatCannotBeReadIsRefusedAsBefore(t *testing.T) {
	h := homeHolding(t, "[workspaces]\nwork = \"w\"\n", nil)
	if _, err := Locate(h); !errors.Is(err, config.ErrNotAbsolute) {
		t.Errorf("Locate = %v, want the folder refused as not absolute", err)
	}
	if _, err := WorkspaceList.Run(On{Home: h}, Input{}); !errors.Is(err, config.ErrNotAbsolute) {
		t.Errorf("workspace list = %v, want the same refusal", err)
	}
}

func TestWorkspaceShow_saysWhichAnswersAndWhy(t *testing.T) {
	for _, c := range []struct {
		name           string
		env            map[string]string
		workspace      string
		store          string
		said           string
		answers, cause string
		add            bool
	}{
		{"the default", nil, "", "", twoWorkspaces, "work", "it is the default", false},
		{"the flag", nil, "home", "", twoWorkspaces, "home", "--workspace names it", false},
		{"the store flag", nil, "", "~/x/cairn.db", twoWorkspaces, "", "--store names it", true},
		{"the store flag on a workspace", nil, "", "~/h/cairn.db", twoWorkspaces, "home", "--store names it", false},
		{"the environment", map[string]string{config.StoreVariable: "~/e/cairn.db"}, "", "", twoWorkspaces, "", "CAIRN_STORE names it", true},
		{"the file", nil, "", "", "store = \"~/s/cairn.db\"\n", "", "the file names it", true},
		{"nothing", nil, "", "", "", "", "nothing names it", true},
	} {
		h := homeHolding(t, c.said, c.env, "w", "h")
		h.Workspace, h.Store = c.workspace, c.store
		answer, err := WorkspaceShow.Run(On{Home: h}, Input{})
		if err != nil {
			t.Errorf("%s: %v", c.name, err)
			continue
		}
		shown := *answer.WorkspaceShown
		if shown.Workspace != c.answers || !strings.HasPrefix(shown.Because, c.cause) || (shown.Add != "") != c.add {
			t.Errorf("%s: shown %+v, want %q answering because %q, suggesting to add it %v", c.name, shown, c.answers, c.cause, c.add)
		}
		said := RenderText(WorkspaceShow, answer)
		if !strings.Contains(said, "because    "+c.cause) || !strings.Contains(said, "store      "+shown.Store+"\n") || !strings.Contains(said, "notes      "+shown.Notes+"\n") {
			t.Errorf("%s: it says\n%s", c.name, said)
		}
	}
}

func TestWorkspaceShow_suggestsTheFolderOfAStoreNamedCairnDB(t *testing.T) {
	h := homeHolding(t, "", map[string]string{config.StoreVariable: "~/e/cairn.db"})
	answer, err := WorkspaceShow.Run(On{Home: h}, Input{})
	if err != nil {
		t.Fatal(err)
	}
	said := RenderText(WorkspaceShow, answer)
	if want := "cairn workspace add <name> " + filepath.Join(h.Dir, "e"); !strings.Contains(said, want) || !strings.Contains(said, "no workspace names this store") {
		t.Errorf("it says\n%s\nwant it suggesting %q", said, want)
	}
	h = homeHolding(t, "", map[string]string{config.StoreVariable: "~/e/tasks.db"})
	answer, _ = WorkspaceShow.Run(On{Home: h}, Input{})
	if answer.WorkspaceShown.Add != "cairn workspace add <name> <folder>" {
		t.Errorf("suggests %q for a store not named cairn.db", answer.WorkspaceShown.Add)
	}
}

func TestWorkspaceShow_saysWhyTheWorkspaceItShowsAnswers(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, nil, "w", "h")
	answer, err := WorkspaceShow.Run(On{Home: h}, Input{})
	if err != nil {
		t.Fatal(err)
	}
	said := RenderText(WorkspaceShow, answer)
	if !strings.Contains(said, "workspace  work\n") || !strings.Contains(said, "folder     "+filepath.Join(h.Dir, "w")+"\n") || strings.Contains(said, "workspace add") {
		t.Errorf("it says\n%s", said)
	}
	h.Workspace = "play"
	if _, err := WorkspaceShow.Run(On{Home: h}, Input{}); !errors.Is(err, ErrNoSuchWorkspace) {
		t.Errorf("workspace show = %v, want play refused", err)
	}
}

func TestWorkspaceList_namesEachAndTheDefault(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, nil)
	answer, err := WorkspaceList.Run(On{Home: h}, Input{})
	if err != nil {
		t.Fatal(err)
	}
	want := []Workspace{{Name: "home", Folder: filepath.Join(h.Dir, "h")}, {Name: "work", Folder: filepath.Join(h.Dir, "w"), Default: true}}
	if got := answer.Workspaces.Workspaces; len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("Workspaces = %+v, want %+v", got, want)
	}
	said := RenderText(WorkspaceList, answer)
	if !strings.Contains(said, "home  "+want[0].Folder+"\n") || !strings.Contains(said, "work  "+want[1].Folder+"  default\n") {
		t.Errorf("it says\n%s", said)
	}
	if got := Of(WorkspaceList, answer); got.(Workspaces).Workspaces[1] != want[1] {
		t.Errorf("Of = %+v", got)
	}
}

func TestWorkspaceList_saysWhenNoneIsNamed(t *testing.T) {
	answer, err := WorkspaceList.Run(On{Home: homeHolding(t, "", nil)}, Input{})
	if err != nil {
		t.Fatal(err)
	}
	if said := RenderText(WorkspaceList, answer); !strings.Contains(said, "no workspace is named") || !strings.Contains(said, "cairn workspace add") {
		t.Errorf("it says\n%s", said)
	}
}

func adding(name, folder string) Input {
	return Input{Given: map[string]string{"name": name, "folder": folder}}
}

func TestWorkspaceAdd_writesItAndMakesItsFolder(t *testing.T) {
	h := homeHolding(t, "notes = \"~/vault\"\n", nil)
	answer, err := WorkspaceAdd.Run(On{Home: h}, adding("play", "~/p"))
	if err != nil {
		t.Fatal(err)
	}
	read, _ := os.ReadFile(config.FileIn(h.Dir))
	if want := "notes = \"~/vault\"\n\n[workspaces]\nplay = \"~/p\"\n"; string(read) != want {
		t.Errorf("the file reads %q, want %q", read, want)
	}
	if info, err := os.Stat(filepath.Join(h.Dir, "p", "notes")); err != nil || !info.IsDir() {
		t.Errorf("the notes folder: %v", err)
	}
	added := *answer.WorkspaceAdded
	if added.Name != "play" || added.Folder != filepath.Join(h.Dir, "p") || added.Store != filepath.Join(h.Dir, "p", "cairn.db") || added.Written != config.FileIn(h.Dir) {
		t.Errorf("added %+v", added)
	}
	if said := RenderText(WorkspaceAdd, answer); !strings.Contains(said, "cairn workspace use play") || !strings.Contains(said, added.Store) {
		t.Errorf("it says\n%s", said)
	}
	if Of(WorkspaceAdd, answer).(WorkspaceAdded) != added {
		t.Error("Of does not answer what was added")
	}
}

func TestWorkspaceAdd_writesTheFileWhenThereIsNone(t *testing.T) {
	h := homeHolding(t, "", nil)
	if _, err := WorkspaceAdd.Run(On{Home: h}, adding("play", "~/p")); err != nil {
		t.Fatal(err)
	}
	if read, _ := os.ReadFile(config.FileIn(h.Dir)); !strings.HasPrefix(string(read), "[workspaces]\nplay = ") {
		t.Errorf("the file reads %q", read)
	}
}

func TestWorkspaceAdd_refusesWhatItCannotName(t *testing.T) {
	for _, c := range []struct {
		name, folder string
		want         error
		names        string
	}{
		{"Work", "~/p", ErrNotAWorkspaceName, `"Work"`},
		{"my_work", "~/p", ErrNotAWorkspaceName, `"my_work"`},
		{"", "~/p", ErrNotAWorkspaceName, `""`},
		{"default", "~/p", ErrNotAWorkspaceName, `"default"`},
		{"work", "~/p", ErrWorkspaceKnown, `"work"`},
		{"play", "p", config.ErrNotAbsolute, `"p"`},
	} {
		h := homeHolding(t, twoWorkspaces, nil)
		if _, err := WorkspaceAdd.Run(On{Home: h}, adding(c.name, c.folder)); !errors.Is(err, c.want) || !strings.Contains(err.Error(), c.names) {
			t.Errorf("add %q %q = %v, want %v naming it", c.name, c.folder, err, c.want)
		}
		if read, _ := os.ReadFile(config.FileIn(h.Dir)); string(read) != twoWorkspaces {
			t.Errorf("add %q %q changed the file: %q", c.name, c.folder, read)
		}
		if _, err := os.Stat(filepath.Join(h.Dir, "p")); err == nil {
			t.Errorf("add %q %q made the folder", c.name, c.folder)
		}
	}
}

func TestWorkspaceAdd_refusesAFileItCannotChange(t *testing.T) {
	h := homeHolding(t, "workspaces = { work = \"/w\" }\n", nil)
	if _, err := WorkspaceAdd.Run(On{Home: h}, adding("play", "~/p")); !errors.Is(err, config.ErrNotWritten) {
		t.Errorf("add = %v, want it refused", err)
	}
	if _, err := os.Stat(filepath.Join(h.Dir, "p")); err == nil {
		t.Error("it made the folder of a workspace it did not write")
	}
}

func TestWorkspaceAddAndUse_needAHome(t *testing.T) {
	missing := errors.New("no home")
	for _, op := range []Operation{WorkspaceAdd, WorkspaceUse} {
		if _, err := op.Run(On{Home: Home{Missing: missing}}, adding("play", "/p")); !errors.Is(err, ErrStore) || !errors.Is(err, missing) {
			t.Errorf("%s = %v, want the missing home said", op.Name(), err)
		}
	}
}

func using(name string) Input { return Input{Given: map[string]string{"name": name}} }

func TestWorkspaceUse_makesItTheDefaultAndSaysWhatItWas(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, nil, "w", "h")
	answer, err := WorkspaceUse.Run(On{Home: h}, using("home"))
	if err != nil {
		t.Fatal(err)
	}
	if *answer.WorkspaceUsed != (WorkspaceUsed{Default: "home", Was: "work"}) {
		t.Errorf("used %+v", *answer.WorkspaceUsed)
	}
	read, _ := os.ReadFile(config.FileIn(h.Dir))
	if want := strings.Replace(twoWorkspaces, `default = "work"`, `default = "home"`, 1); string(read) != want {
		t.Errorf("the file reads %q, want %q", read, want)
	}
	if said := RenderText(WorkspaceUse, answer); said != "the default workspace is now \"home\"; it was \"work\"\n" {
		t.Errorf("it says %q", said)
	}
	if Of(WorkspaceUse, answer).(WorkspaceUsed).Default != "home" {
		t.Error("Of does not answer the default")
	}
}

func TestWorkspaceUse_saysWhenThereWasNoDefault(t *testing.T) {
	h := homeHolding(t, "[workspaces]\nwork = \"~/w\"\n", nil, "w")
	answer, err := WorkspaceUse.Run(On{Home: h}, using("work"))
	if err != nil {
		t.Fatal(err)
	}
	if said := RenderText(WorkspaceUse, answer); said != "the default workspace is now \"work\"; there was none\n" {
		t.Errorf("it says %q", said)
	}
}

func TestWorkspaceUse_refusesAWorkspaceNotThere(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, nil, "w")
	if _, err := WorkspaceUse.Run(On{Home: h}, using("play")); !errors.Is(err, ErrNoSuchWorkspace) {
		t.Errorf("use play = %v", err)
	}
	if _, err := WorkspaceUse.Run(On{Home: h}, using("home")); !errors.Is(err, ErrWorkspaceNotThere) {
		t.Errorf("use home = %v", err)
	}
	if read, _ := os.ReadFile(config.FileIn(h.Dir)); string(read) != twoWorkspaces {
		t.Errorf("the file was changed: %q", read)
	}
	if _, err := WorkspaceUse.Run(On{Home: homeHolding(t, "[workspaces]\n", nil)}, using("x")); !errors.Is(err, ErrNoSuchWorkspace) {
		t.Errorf("use x = %v", err)
	}
}

func TestWorkspaceUse_refusesAFileItCannotChange(t *testing.T) {
	h := homeHolding(t, "workspaces = { work = \"~/w\" }\n", nil, "w")
	if _, err := WorkspaceUse.Run(On{Home: h}, using("work")); !errors.Is(err, config.ErrNotWritten) {
		t.Errorf("use = %v, want it refused", err)
	}
}

func TestWorkspaces_areDescribedForEverySurface(t *testing.T) {
	for _, op := range []Operation{WorkspaceList, WorkspaceAdd, WorkspaceUse, WorkspaceShow} {
		if op.Needs != AConfiguration || op.Thing != "workspace" {
			t.Errorf("%s needs %v", op.Name(), op.Needs)
		}
	}
	for _, op := range []Operation{WorkspaceList, WorkspaceShow} {
		if _, found := Called(op.Thing, op.Verb); !found {
			t.Errorf("%s is not among All(), so not offered to an agent", op.Name())
		}
	}
	for _, op := range []Operation{WorkspaceAdd, WorkspaceUse} {
		if _, found := Called(op.Thing, op.Verb); found {
			t.Errorf("%s is among All(), so offered to an agent", op.Name())
		}
	}
	if typed := ForAPersonOnly(); len(typed) != 2 || typed[0].Name() != WorkspaceAdd.Name() || typed[1].Name() != WorkspaceUse.Name() {
		t.Errorf("ForAPersonOnly() = %d operations, want workspace add and use", len(typed))
	}
	if WorkspaceList.Level != Read || WorkspaceShow.Level != Read || WorkspaceAdd.Level != Write || WorkspaceUse.Level != Write {
		t.Error("listing and showing read, adding and using write")
	}
}

func TestConfigShow_saysTheWorkspacesAndWhatNamedEachPlace(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, nil, "w", "h")
	answer, err := ConfigShow.Run(On{Home: h}, Input{})
	if err != nil {
		t.Fatal(err)
	}
	c := *answer.Config
	if c.Store != (Emplacement{At: filepath.Join(h.Dir, "w", "cairn.db"), By: `the default workspace, "work"`}) || c.Plugin.At != filepath.Join(h.Dir, "w", "claude-plugin") {
		t.Errorf("config %+v", c)
	}
	if len(c.Workspaces) != 2 || !c.Workspaces[1].Default {
		t.Errorf("workspaces %+v", c.Workspaces)
	}
	said := RenderText(ConfigShow, answer)
	if !strings.Contains(said, "\n[workspaces]\ndefault = \"work\"\nhome = \""+filepath.Join(h.Dir, "h")+"\"\nwork = ") {
		t.Errorf("it says\n%s", said)
	}
	h.Workspace = "play"
	answer, err = ConfigShow.Run(On{Home: h}, Input{})
	if err != nil || !strings.Contains(answer.Config.Problem, "play") {
		t.Errorf("config show = %+v, %v; want the workspace said as the problem", answer.Config, err)
	}
	h.Workspace, h.Store = "home", ""
	answer, _ = ConfigShow.Run(On{Home: h}, Input{})
	if answer.Config.Store.By != "--workspace" {
		t.Errorf("store named by %q", answer.Config.Store.By)
	}
}

func TestConfigShow_saysWhenNoWorkspaceIsNamed(t *testing.T) {
	answer, err := ConfigShow.Run(On{Home: homeHolding(t, "", nil)}, Input{})
	if err != nil {
		t.Fatal(err)
	}
	if said := RenderText(ConfigShow, answer); !strings.Contains(said, "\n[workspaces]\n# none is named") {
		t.Errorf("it says\n%s", said)
	}
}
