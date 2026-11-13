package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"cairn/internal/config"
)

func TestWorkspaceShow_whyNamesTheRungAndThoseItPassedOver(t *testing.T) {
	at := func(h Home, path string) string { return strings.ReplaceAll(path, "~", h.Dir) }
	for _, c := range []struct {
		name             string
		env              map[string]string
		workspace, store string
		said             string
		why              Why
		says             string
	}{
		{"the default", nil, "", "", twoWorkspaces, Why{Rung: rungDefault, Unsaid: []string{rungWorkspaceFlag, config.StoreVariable}}, "the default (no --workspace, no CAIRN_STORE)"},
		{"the environment on a workspace", map[string]string{config.StoreVariable: "~/h/cairn.db"}, "", "", twoWorkspaces, Why{Rung: config.StoreVariable, Over: []Rung{{Rung: rungDefault, Names: "work"}}}, "CAIRN_STORE — over the default, work"},
		{"the flag", map[string]string{config.StoreVariable: "~/h/cairn.db"}, "home", "", twoWorkspaces, Why{Rung: rungWorkspaceFlag, Over: []Rung{{Rung: config.StoreVariable, Names: "home"}, {Rung: rungDefault, Names: "work"}}}, "--workspace — over CAIRN_STORE (home) and the default (work)"},
		{"CAIRN_WORKSPACE", map[string]string{WorkspaceVariable: "home"}, "", "", twoWorkspaces, Why{Rung: WorkspaceVariable, Over: []Rung{{Rung: rungDefault, Names: "work"}}}, "CAIRN_WORKSPACE — over the default, work"},
		{"CAIRN_STORE over CAIRN_WORKSPACE", map[string]string{config.StoreVariable: "~/e/cairn.db", WorkspaceVariable: "home"}, "", "", twoWorkspaces, Why{Rung: config.StoreVariable, Over: []Rung{{Rung: WorkspaceVariable, Names: "home"}, {Rung: rungDefault, Names: "work"}}}, "CAIRN_STORE — over CAIRN_WORKSPACE (home) and the default (work)"},
		{"the flag over CAIRN_WORKSPACE", map[string]string{WorkspaceVariable: "home"}, "work", "", twoWorkspaces, Why{Rung: rungWorkspaceFlag, Over: []Rung{{Rung: WorkspaceVariable, Names: "home"}, {Rung: rungDefault, Names: "work"}}}, "--workspace — over CAIRN_WORKSPACE (home) and the default (work)"},
		{"the store flag", nil, "", "~/x/cairn.db", twoWorkspaces, Why{Rung: rungStoreFlag, Over: []Rung{{Rung: rungDefault, Names: "work"}}}, "--store — over the default, work"},
		{"the environment on no workspace", map[string]string{config.StoreVariable: "~/e/cairn.db"}, "", "~/x/cairn.db", "", Why{Rung: rungStoreFlag, Over: []Rung{{Rung: config.StoreVariable, Names: "~/e/cairn.db"}}}, "--store — over CAIRN_STORE, ~/e/cairn.db"},
		{"the default over the file", nil, "", "", "store = \"~/s/tasks.db\"\n" + twoWorkspaces, Why{Rung: rungDefault, Unsaid: []string{rungWorkspaceFlag, config.StoreVariable}, Over: []Rung{{Rung: rungFile, Names: "~/s/tasks.db"}}}, "the default (no --workspace, no CAIRN_STORE) — over the file, ~/s/tasks.db"},
		{"the file", nil, "", "", "store = \"~/s/tasks.db\"\n", Why{Rung: rungFile, Unsaid: []string{rungWorkspaceFlag, config.StoreVariable, rungDefault}}, "the file (no --workspace, no CAIRN_STORE, no default)"},
		{"nothing", nil, "", "", "", Why{Rung: rungNothing, Unsaid: []string{rungWorkspaceFlag, config.StoreVariable, rungDefault, rungFile}}, "where cairn keeps a store by default (no --workspace, no CAIRN_STORE, no default, no store in the file)"},
	} {
		h := homeHolding(t, c.said, c.env, "w", "h")
		h.Workspace, h.Store = c.workspace, c.store
		answer, err := WorkspaceShow.Run(On{Home: h}, Input{})
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
		if said := RenderText(WorkspaceShow, answer); !strings.Contains(said, "why        "+at(h, c.says)+"\n") || strings.Contains(said, "because") {
			t.Errorf("%s: it says\n%s\nwant why %q", c.name, said, at(h, c.says))
		}
	}
}

func TestWorkspaceShow_aServerSaysWhyAsItWasWhenItStarted(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, nil, "w", "h")
	h.Workspace = "home"
	kept := Keep(h)
	if kept.Kept == nil || !kept.Kept.Why.Started || kept.Kept.Why.Rung != rungWorkspaceFlag {
		t.Fatalf("kept %+v, want why --workspace, settled as the server started", kept.Kept)
	}
	answer, err := WorkspaceShow.Run(On{Home: kept}, Input{})
	if err != nil {
		t.Fatal(err)
	}
	if said := RenderText(WorkspaceShow, answer); !strings.Contains(said, "why        when this server started, --workspace — over the default, work\n") {
		t.Errorf("it says\n%s", said)
	}
	said, err := json.Marshal(answer.WorkspaceShown)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(said), `"why":{"rung":"--workspace","over":[{"rung":"the default","names":"work"}],"started":true}`) {
		t.Errorf("a machine reads %s", said)
	}
}

func TestWorkspaceShow_saysThePagesOnAndWhereThatIsSaid(t *testing.T) {
	for _, c := range []struct {
		name, said, settings string
		pages                []string
		from, says           string
	}{
		{"by default", twoWorkspaces, "", []string{"tasks", "objectives", "time"}, "by default", "tasks, objectives and time, by default"},
		{"the global file", "[pages]\nobjectives = false\ntime = \"off\"\n" + twoWorkspaces, "", []string{"tasks"}, "from the global configuration", "tasks, from the global configuration"},
		{"the store's settings", "[pages]\nobjectives = false\n" + twoWorkspaces, "[pages]\ntasks = false\ntime = \"off\"\n", nil, "set in the store's settings", "none, set in the store's settings"},
		{"the time alone said here", "[pages]\nobjectives = false\n" + twoWorkspaces, "[pages]\ntime = \"hidden\"\n", []string{"tasks"}, "set in the store's settings", "tasks, set in the store's settings"},
	} {
		h := homeHolding(t, c.said, nil, "w", "h")
		if c.settings != "" {
			settingsIn(t, filepath.Join(h.Dir, "w"), c.settings)
		}
		answer, err := WorkspaceShow.Run(On{Home: h}, Input{})
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if shown := answer.WorkspaceShown; !slices.Equal(shown.Pages, c.pages) || shown.PagesFrom != c.from {
			t.Errorf("%s: pages %v %s, want %v %s", c.name, shown.Pages, shown.PagesFrom, c.pages, c.from)
		}
		if said := RenderText(WorkspaceShow, answer); !strings.Contains(said, "pages      "+c.says+"\n") {
			t.Errorf("%s: it says\n%s", c.name, said)
		}
	}
}

func TestWorkspaceList_marksTheTimeEachWorkspaceKeeps(t *testing.T) {
	h := homeHolding(t, "[pages]\ntime = \"hidden\"\n"+twoWorkspaces+"play = \"~/p\"\n", nil, "w", "h", "p")
	settingsIn(t, filepath.Join(h.Dir, "h"), "[pages]\ntime = \"off\"\n")
	settingsIn(t, filepath.Join(h.Dir, "p"), "[pages]\ntime = \"on\"\n")
	answer, err := WorkspaceList.Run(On{Home: h}, Input{})
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
	said := RenderText(WorkspaceList, answer)
	for _, want := range []string{"home  " + filepath.Join(h.Dir, "h") + "  time off  ", "work  " + filepath.Join(h.Dir, "w") + "  default  time hidden  "} {
		if !strings.Contains(said, want) {
			t.Errorf("it says\n%s\nwant %q", said, want)
		}
	}
	if strings.Contains(said, "play  "+filepath.Join(h.Dir, "p")+"  time") {
		t.Errorf("a workspace whose time is on is marked:\n%s", said)
	}
}

func TestWorkspaceList_leavesUnmarkedTheTimeOfSettingsItCannotRead(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, nil, "w", "h")
	settingsIn(t, filepath.Join(h.Dir, "h"), "not toml [")
	answer, err := WorkspaceList.Run(On{Home: h}, Input{})
	if err != nil {
		t.Fatal(err)
	}
	for _, one := range answer.Workspaces.Workspaces {
		if one.Time != "" {
			t.Errorf("%s is marked time %q", one.Name, one.Time)
		}
	}
}

func TestConfigShow_saysWhatTheGlobalConfigurationGivesBesideAKeySetHere(t *testing.T) {
	h := homeHolding(t, "[focus]\nsession = \"1h\"\n[approach]\nlabel = \"discuss\"\n", map[string]string{config.StoreVariable: "~/s/cairn.db"})
	settingsIn(t, filepath.Join(h.Dir, "s"), "[focus]\nsession = \"25m\"\n[time]\nweekends = \"hidden\"\n[approach]\nlabel = \"\"\n[archive]\nstart = 2026-01-01\ndays = 30\n[pages]\ntasks = false\ntime = \"off\"\n[window]\nkeep_sizes = true\n[workspace]\ncolour = \"teal\"\n")
	answer, err := ConfigShow.Run(On{Home: h}, Input{})
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
	said := configText(*answer.Config)
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

func TestConfigShow_saysTheColourOfTheWorkspaceAndWhereItComesFrom(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, nil, "w", "h")
	answer, err := ConfigShow.Run(On{Home: h}, Input{})
	if err != nil {
		t.Fatal(err)
	}
	if c := answer.Config; c.Workspace == nil || c.Workspace.Colour != colourOf("work", "") {
		t.Fatalf("Config = %+v, want the colour read off work", c.Workspace)
	}
	if said := configText(*answer.Config); !strings.Contains(said, "\n[workspace]\ncolour = \""+colourOf("work", "")+"\"   # read off the workspace's name\n\n[workspaces]") {
		t.Errorf("config show says\n%s", said)
	}
	loose := homeHolding(t, "", map[string]string{config.StoreVariable: "~/s/cairn.db"})
	answer, err = ConfigShow.Run(On{Home: loose}, Input{})
	if err != nil {
		t.Fatal(err)
	}
	if answer.Config.Workspace != nil || strings.Contains(configText(*answer.Config), "[workspace]\n") {
		t.Errorf("a store no workspace names says a colour: %+v", answer.Config.Workspace)
	}
}

func TestConfigInit_namesOneWorkspaceTheDefaultAndMakesItsFolder(t *testing.T) {
	home := t.TempDir()
	if _, err := ConfigInit.Run(On{Home: Home{Dir: home}}, Input{}); err != nil {
		t.Fatal(err)
	}
	at, err := Locate(Home{Dir: home})
	if err != nil {
		t.Fatal(err)
	}
	if at.Store != filepath.Join(home, ".cairn", "cairn.db") || at.Notes != filepath.Join(home, ".cairn", "notes") {
		t.Errorf("places %+v, want those cairn has always used", at)
	}
	if _, err := os.Stat(filepath.Join(home, ".cairn", "notes")); err != nil {
		t.Errorf("the folder of the default workspace is not made: %v", err)
	}
	read, err := config.Read(home, func(string) (string, bool) { return "", false })
	if err != nil || read.DefaultWorkspace() != "main" {
		t.Errorf("the default is %q, %v; want main", read.DefaultWorkspace(), err)
	}
}
