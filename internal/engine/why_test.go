package engine

import (
	"cairn/internal/engine/workspace"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"cairn/internal/store/files"
)

func TestWorkspaceShow_whyNamesTheRungAndThoseItPassedOver(t *testing.T) {
	at := func(h Home, path string) string { return strings.ReplaceAll(path, "~", h.Dir) }
	for _, c := range []struct {
		name             string
		env              map[string]string
		workspace, store string
		said             string
		why              workspace.Why
		says             string
	}{
		{"the default", nil, "", "", twoWorkspaces, workspace.Why{Rung: workspace.RungDefault, Unsaid: []string{workspace.RungWorkspaceFlag, files.StoreVariable}}, "the default (no --workspace, no CAIRN_STORE)"},
		{"the environment on a workspace", map[string]string{files.StoreVariable: "~/h/cairn.db"}, "", "", twoWorkspaces, workspace.Why{Rung: files.StoreVariable, Over: []workspace.Rung{{Rung: workspace.RungDefault, Names: "work"}}}, "CAIRN_STORE — over the default, work"},
		{"the flag", map[string]string{files.StoreVariable: "~/h/cairn.db"}, "home", "", twoWorkspaces, workspace.Why{Rung: workspace.RungWorkspaceFlag, Over: []workspace.Rung{{Rung: files.StoreVariable, Names: "home"}, {Rung: workspace.RungDefault, Names: "work"}}}, "--workspace — over CAIRN_STORE (home) and the default (work)"},
		{"CAIRN_WORKSPACE", map[string]string{workspace.WorkspaceVariable: "home"}, "", "", twoWorkspaces, workspace.Why{Rung: workspace.WorkspaceVariable, Over: []workspace.Rung{{Rung: workspace.RungDefault, Names: "work"}}}, "CAIRN_WORKSPACE — over the default, work"},
		{"CAIRN_STORE over CAIRN_WORKSPACE", map[string]string{files.StoreVariable: "~/e/cairn.db", workspace.WorkspaceVariable: "home"}, "", "", twoWorkspaces, workspace.Why{Rung: files.StoreVariable, Over: []workspace.Rung{{Rung: workspace.WorkspaceVariable, Names: "home"}, {Rung: workspace.RungDefault, Names: "work"}}}, "CAIRN_STORE — over CAIRN_WORKSPACE (home) and the default (work)"},
		{"the flag over CAIRN_WORKSPACE", map[string]string{workspace.WorkspaceVariable: "home"}, "work", "", twoWorkspaces, workspace.Why{Rung: workspace.RungWorkspaceFlag, Over: []workspace.Rung{{Rung: workspace.WorkspaceVariable, Names: "home"}, {Rung: workspace.RungDefault, Names: "work"}}}, "--workspace — over CAIRN_WORKSPACE (home) and the default (work)"},
		{"the store flag", nil, "", "~/x/cairn.db", twoWorkspaces, workspace.Why{Rung: workspace.RungStoreFlag, Over: []workspace.Rung{{Rung: workspace.RungDefault, Names: "work"}}}, "--store — over the default, work"},
		{"the environment on no workspace", map[string]string{files.StoreVariable: "~/e/cairn.db"}, "", "~/x/cairn.db", "", workspace.Why{Rung: workspace.RungStoreFlag, Over: []workspace.Rung{{Rung: files.StoreVariable, Names: "~/e/cairn.db"}}}, "--store — over CAIRN_STORE, ~/e/cairn.db"},
		{"the default over the file", nil, "", "", "store = \"~/s/tasks.db\"\n" + twoWorkspaces, workspace.Why{Rung: workspace.RungDefault, Unsaid: []string{workspace.RungWorkspaceFlag, files.StoreVariable}, Over: []workspace.Rung{{Rung: workspace.RungFile, Names: "~/s/tasks.db"}}}, "the default (no --workspace, no CAIRN_STORE) — over the file, ~/s/tasks.db"},
		{"the file", nil, "", "", "store = \"~/s/tasks.db\"\n", workspace.Why{Rung: workspace.RungFile, Unsaid: []string{workspace.RungWorkspaceFlag, files.StoreVariable, workspace.RungDefault}}, "the file (no --workspace, no CAIRN_STORE, no default)"},
		{"nothing", nil, "", "", "", workspace.Why{Rung: workspace.RungNothing, Unsaid: []string{workspace.RungWorkspaceFlag, files.StoreVariable, workspace.RungDefault, workspace.RungFile}}, "where cairn keeps a store by default (no --workspace, no CAIRN_STORE, no default, no store in the file)"},
	} {
		h := homeHolding(t, c.said, c.env, "w", "h")
		h.Workspace, h.Store = c.workspace, c.store
		answer, err := WorkspaceShow.Run(workedOn{Home: h}, Input{})
		if err != nil {
			t.Errorf("%s: %v", c.name, err)
			continue
		}
		for i := range c.why.Over {
			c.why.Over[i].Names = at(h, c.why.Over[i].Names)
		}
		if got := answer.WorkspaceShown.Why; !reflect.DeepEqual(got, c.why) {
			t.Errorf("%s: why %+v, want %+v", c.name, got, c.why)
		}
	}
}

func TestWorkspaceShow_saysThePagesOnAndWhereThatIsSaid(t *testing.T) {
	for _, c := range []struct {
		name, said, settings string
		pages                []string
		from, says           string
	}{
		{"by default", twoWorkspaces, "", []string{"tasks", "objectives", "time", "agents"}, "by default", "tasks, objectives, time and agents, by default"},
		{"the global file", "[pages]\nobjectives = false\ntime = \"off\"\n" + twoWorkspaces, "", []string{"tasks", "agents"}, "from the global configuration", "tasks and agents, from the global configuration"},
		{"the store's settings", "[pages]\nobjectives = false\n" + twoWorkspaces, "[pages]\ntasks = false\ntime = \"off\"\n", []string{"agents"}, "set in the store's settings", "agents, set in the store's settings"},
		{"the time alone said here", "[pages]\nobjectives = false\n" + twoWorkspaces, "[pages]\ntime = \"hidden\"\n", []string{"tasks", "agents"}, "set in the store's settings", "tasks and agents, set in the store's settings"},
	} {
		h := homeHolding(t, c.said, nil, "w", "h")
		if c.settings != "" {
			settingsIn(t, filepath.Join(h.Dir, "w"), c.settings)
		}
		answer, err := WorkspaceShow.Run(workedOn{Home: h}, Input{})
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if shown := answer.WorkspaceShown; !slices.Equal(shown.Pages, c.pages) || shown.PagesFrom != c.from {
			t.Errorf("%s: pages %v %s, want %v %s", c.name, shown.Pages, shown.PagesFrom, c.pages, c.from)
		}
	}
}

func TestWorkspaceList_marksTheTimeEachWorkspaceKeeps(t *testing.T) {
	h := homeHolding(t, "[pages]\ntime = \"hidden\"\n"+twoWorkspaces+"play = \"~/p\"\n", nil, "w", "h", "p")
	settingsIn(t, filepath.Join(h.Dir, "h"), "[pages]\ntime = \"off\"\n")
	settingsIn(t, filepath.Join(h.Dir, "p"), "[pages]\ntime = \"on\"\n")
	answer, err := WorkspaceList.Run(workedOn{Home: h}, Input{})
	if err != nil {
		t.Fatal(err)
	}
	times := map[string]string{}
	for _, one := range answer.Workspaces.Workspaces {
		times[one.Name] = one.Time
	}
	if want := map[string]string{"home": "off", "play": "", "work": "hidden"}; times["home"] != want["home"] || times["play"] != want["play"] || times["work"] != want["work"] {
		t.Errorf("times %v, want %v", times, want)
	}
}

func TestConfigShow_saysWhatTheGlobalConfigurationGivesBesideAKeySetHere(t *testing.T) {
	h := homeHolding(t, "[focus]\nsession = \"1h\"\n[approach]\nlabel = \"discuss\"\n", map[string]string{files.StoreVariable: "~/s/cairn.db"})
	settingsIn(t, filepath.Join(h.Dir, "s"), "[focus]\nsession = \"25m\"\n[time]\nweekends = \"hidden\"\n[approach]\nlabel = \"\"\n[archive]\nstart = 2026-01-01\ndays = 30\n[pages]\ntasks = false\ntime = \"off\"\n[window]\nkeep_sizes = true\n[workspace]\ncolour = \"teal\"\n")
	answer, err := ConfigShow.Run(workedOn{Home: h}, Input{})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"focus.session": "1h", "time.weekends": "shown", "approach.label": "discuss",
		"archive.start": "not said", "archive.days": "14", "pages.tasks": "true", "pages.time": "on", "window.keep_sizes": "false",
	}
	if got := answer.Config.Global; len(got) != len(want) {
		t.Errorf("global %v, want %v", got, want)
	}
	for key, value := range want {
		if got := answer.Config.Global[key]; got != value {
			t.Errorf("global %s = %q, want %q", key, got, value)
		}
	}
	said := workspace.ConfigText(*answer.Config)
	for _, line := range []string{
		"session = \"25m\"   # set in the store's settings; global: \"1h\"\n",
		"weekends = \"hidden\"   # set in the store's settings; global: \"shown\"\n",
		"days = 30   # set in the store's settings; global: 14\n",
		"start = 2026-01-01   # set in the store's settings; global: not said\n",
		"tasks = false   # set in the store's settings; global: true\n",
		"objectives = true   # by default\n",
		"# label is said empty in the store's settings: no label asks for the approach to be agreed first; global: \"discuss\"\n",
		"\n[workspace]\ncolour = \"teal\"   # set in the store's settings\n",
	} {
		if !strings.Contains(said, line) {
			t.Errorf("config show does not say %q:\n%s", line, said)
		}
	}
}
