package engine

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"cairn/internal/config"
	"cairn/internal/trace"
)

func settingsIn(t *testing.T, folder, said string) {
	t.Helper()
	if err := os.MkdirAll(folder, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(folder, config.SettingsFile), []byte(said), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestPagesFrom_readsTheSettingsOfTheStoreInUse(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, nil, "w", "h")
	settingsIn(t, filepath.Join(h.Dir, "h"), "[pages]\nobjectives = false\n[time]\nweekends = \"hidden\"\n")
	if !PagesFrom(h).Shows("objectives") {
		t.Error("the default workspace takes the objectives of another's settings")
	}
	h.Workspace = "home"
	if pages := PagesFrom(h); pages.Shows("objectives") || pages.WeekDays() != 5 {
		t.Error("the settings of home do not turn its objectives off and hide its weekends")
	}
}

func TestKeepsSizes_readsTheSettingsBesideTheStore(t *testing.T) {
	h := homeHolding(t, "", map[string]string{config.StoreVariable: "~/s/cairn.db"})
	settingsIn(t, filepath.Join(h.Dir, "s"), "[window]\nkeep_sizes = true\n")
	if !KeepsSizes(h) {
		t.Error("keep_sizes = true beside the store does not keep the window's sizes")
	}
}

func TestLocate_refusesSettingsThatSayWhatBelongsToTheMachine(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, nil, "w", "h")
	settingsIn(t, filepath.Join(h.Dir, "w"), "store = \"~/x/cairn.db\"\n")
	_, err := Locate(h)
	if !errors.Is(err, config.ErrMachineKey) || RefusalOf(err).Kind != KindRefused {
		t.Errorf("Locate = %v, want ErrMachineKey, refused rather than a store unusable", err)
	}
	if RefusalOf(err).Reason != "config-machine-key" {
		t.Errorf("the refusal is %+v", RefusalOf(err))
	}
	settingsIn(t, filepath.Join(h.Dir, "w"), "[focus]\nsession = \"\"\n")
	if _, err := Locate(h); RefusalOf(err).Reason != "config-empty" {
		t.Errorf("Locate = %v, want an empty string refused", err)
	}
}

func TestLocate_settingsThatCannotBeReadLeaveTheStoreUnusable(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, nil, "w", "h")
	if err := os.MkdirAll(filepath.Join(h.Dir, "w", config.SettingsFile), 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := Locate(h); RefusalOf(err).Kind != KindStore {
		t.Errorf("Locate = %v, want the store unusable", err)
	}
}

func TestLocate_settingsThatAreNotTOMLAreRefusedNamingThem(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, nil, "w", "h")
	settingsIn(t, filepath.Join(h.Dir, "w"), "[pages\n")
	_, err := Locate(h)
	if !errors.Is(err, config.ErrNotTOML) || !strings.Contains(err.Error(), filepath.Join(h.Dir, "w", config.SettingsFile)) {
		t.Errorf("Locate = %v, want ErrNotTOML naming the settings", err)
	}
}

func TestDelete_takesItsTimeWhenTheSettingsOfItsStoreSaySo(t *testing.T) {
	path := storeAt(t, "one")
	settingsIn(t, filepath.Dir(path), "[time]\ndeleted = \"taken\"\n")
	if _, err := DoFrom(TaskDelete, homeSaying(t, path, ""), Agreed(given("number", "1")), Acting{}); err != nil {
		t.Fatal(err)
	}
	e, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = e.Close() }()
	acts, err := e.history.ActsOn(1)
	if err != nil {
		t.Fatal(err)
	}
	if went := acts[len(acts)-1].Carried().Was; !strings.HasSuffix(went, trace.WentWithItsTime) {
		t.Errorf("the deletion took %q, want its time with it", went)
	}
}

func TestSessionLength_isTheOneTheSettingsOfTheStoreSay(t *testing.T) {
	path := storeAt(t, "one")
	settingsIn(t, filepath.Dir(path), "[focus]\nsession = \"25m\"\n")
	e, err := OpenAt(Places{Store: path, Home: homeSaying(t, path, "[focus]\nsession = \"1h\"\n")})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = e.Close() }()
	if length, err := e.sessionLength(); err != nil || length.String() != "25m0s" {
		t.Errorf("sessionLength = %v, %v; want 25m", length, err)
	}
	settingsIn(t, filepath.Dir(path), "[pages\n")
	if length, err := e.sessionLength(); err != nil || length != 0 {
		t.Errorf("with settings it cannot read, sessionLength = %v, %v; want none, said where the places are read", length, err)
	}
}

func TestConfigured_isTheGlobalConfigurationWithTheSettingsBesideTheStore(t *testing.T) {
	path := storeAt(t, "one")
	settingsIn(t, filepath.Dir(path), "[archive]\nstart = 2026-10-21\n")
	e, err := OpenAt(Places{Store: path, Home: homeSaying(t, path, "[archive]\ndays = 7\n")})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = e.Close() }()
	read, err := e.configured()
	if start, days, on := read.Archive(); err != nil || start != "2026-10-21" || days != 7 || !on {
		t.Errorf("configured archives from %q every %d days (%v), %v; want 2026-10-21 every 7", start, days, on, err)
	}
}

func TestIgnored_saysWhatTheGlobalFileAndTheSettingsPassOver(t *testing.T) {
	h := homeHolding(t, "colour = 1\n", map[string]string{config.StoreVariable: "~/s/cairn.db"})
	settingsIn(t, filepath.Join(h.Dir, "s"), "[pages]\ncalendar = false\n")
	said := Ignored(h)
	if len(said) != 2 || !strings.Contains(said[1], filepath.Join(h.Dir, "s", config.SettingsFile)) {
		t.Errorf("Ignored = %q, want the global key and the settings' one, naming the settings", said)
	}
	if Ignored(Home{}) != nil {
		t.Error("with no home, something is said to be ignored")
	}
	settingsIn(t, filepath.Join(h.Dir, "s"), "store = \"~/x\"\n")
	if Ignored(h) != nil {
		t.Error("settings refused are said to pass keys over")
	}
}

func TestConfigShow_saysWhereEachStoreKeyComesFrom(t *testing.T) {
	h := homeHolding(t, "[focus]\nsession = \"1h\"\n[approach]\nlabel = \"discuss\"\n", map[string]string{config.StoreVariable: "~/s/cairn.db"})
	settingsIn(t, filepath.Join(h.Dir, "s"), "[time]\nweekends = \"hidden\"\n[approach]\nlabel = \"\"\n")
	answer, err := ConfigShow.Run(On{Home: h}, Input{})
	if err != nil {
		t.Fatal(err)
	}
	c := *answer.Config
	settings := filepath.Join(h.Dir, "s", config.SettingsFile)
	if c.Settings != settings || !slices.Equal(c.WrittenHere, []string{"approach.label", "time.weekends"}) || c.Time.Weekends != config.WeekendsHidden {
		t.Errorf("Config = %+v, want the settings read over the global file", c)
	}
	said := configText(c)
	for _, line := range []string{
		"# the store's settings: " + settings + "\n",
		"weekends = \"hidden\"   # set in the store's settings; global: \"shown\"\n",
		"session = \"1h\"   # from the global configuration\n",
		"deleted = \"kept\"   # by default\n",
		"tasks = true   # by default\n",
		"days = 14   # by default\n",
		"keep_sizes = false   # by default\n",
		"# label is said empty in the store's settings: no label asks for the approach to be agreed first; global: \"discuss\"\n",
	} {
		if !strings.Contains(said, line) {
			t.Errorf("config show does not say %q:\n%s", line, said)
		}
	}
}

func TestConfigShow_saysNothingOfSettingsThatAreNotThere(t *testing.T) {
	said := configText(shownWith(t, "[focus]\nsession = \"1h\"\n"))
	if strings.Contains(said, "settings") || strings.Contains(said, "# by default") {
		t.Errorf("with no settings, config show says:\n%s", said)
	}
}

func TestConfigShow_saysSettingsItCannotMakeSenseOfAsTheProblem(t *testing.T) {
	h := homeHolding(t, "", map[string]string{config.StoreVariable: "~/s/cairn.db"})
	settingsIn(t, filepath.Join(h.Dir, "s"), "plugin = \"~/p\"\n")
	answer, err := ConfigShow.Run(On{Home: h}, Input{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(answer.Config.Problem, "plugin") || answer.Config.Pages != nil {
		t.Errorf("Config = %+v, want the problem and no key", answer.Config)
	}
}

func TestWorkspaceAdd_writesSettingsThatSayNothing(t *testing.T) {
	h := homeHolding(t, "", nil)
	if _, err := WorkspaceAdd.Run(On{Home: h}, adding("play", "~/p")); err != nil {
		t.Fatal(err)
	}
	if read, _ := os.ReadFile(filepath.Join(h.Dir, "p", config.SettingsFile)); string(read) != config.SettingsHeader {
		t.Errorf("the settings read %q", read)
	}
}

func TestWorkspaceAdd_leavesSettingsAlreadyThere(t *testing.T) {
	h := homeHolding(t, "", nil)
	settingsIn(t, filepath.Join(h.Dir, "p"), "[pages]\ntime = false\n")
	if _, err := WorkspaceAdd.Run(On{Home: h}, adding("play", "~/p")); err != nil {
		t.Fatal(err)
	}
	if read, _ := os.ReadFile(filepath.Join(h.Dir, "p", config.SettingsFile)); string(read) != "[pages]\ntime = false\n" {
		t.Errorf("the settings read %q", read)
	}
}
