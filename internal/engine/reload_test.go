package engine

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"cairn/internal/config"
)

func answered(t *testing.T, h Home) Config {
	t.Helper()
	got, err := DoFrom(ConfigShow, h, Input{}, Acting{})
	if err != nil {
		t.Fatal(err)
	}
	return *got.Config
}

func TestPagesOf_readsThePagesTheWeekAndTheSeriesOffAnAnswer(t *testing.T) {
	h := homeHolding(t, "[pages]\nobjectives = false\ntime = \"hidden\"\n[time]\nweekends = \"hidden\"\n[series]\non = true\n", nil)
	p := PagesOf(answered(t, h))
	if p.Shows(PageObjectives) || p.Shows(PageTime) || !p.Shows(PageTasks) || p.WeekDays() != 5 || !p.Recurs() {
		t.Errorf("the pages read %+v", p)
	}
	if p := PagesOf(Config{}); !p.Shows(PageTasks) || p.Recurs() || p.WeekDays() != 7 {
		t.Errorf("an answer saying nothing reads %+v", p)
	}
}

func TestConfigChanges_namesEachKeyThatMovedAndFromWhat(t *testing.T) {
	h := homeHolding(t, homeIsDefault, nil, "h")
	was := answered(t, h)
	settingsOf(t, h, "h", "[pages]\ntime = \"off\"\n[window]\ntheme = \"nord\"\n")
	now := answered(t, h)
	changes := ConfigChanges(was, now)
	var said []string
	for _, c := range changes {
		said = append(said, c.String())
	}
	if !slices.Equal(said, []string{"window.theme default → nord", "pages.time on → off"}) {
		t.Errorf("the changes are %v", said)
	}
	if ConfigChanges(now, now) != nil {
		t.Error("an answer changed from itself")
	}
}

func TestConfigStamp_movesOnlyWhenWhatAFileSaysChanges(t *testing.T) {
	h := homeHolding(t, homeIsDefault, nil, "h")
	path := settingsOf(t, h, "h", "[pages]\ntime = \"off\"\n")
	c := answered(t, h)
	stamp := StampOf(c)
	if _, moved := stamp.Moved(); moved {
		t.Error("the stamp moved with nothing written")
	}
	later := time.Now().Add(time.Minute)
	if err := os.Chtimes(path, later, later); err != nil {
		t.Fatal(err)
	}
	stamp, moved := stamp.Moved()
	if moved {
		t.Error("a file touched and not changed moved the stamp")
	}
	settingsOf(t, h, "h", "[pages]\ntime = \"on\"\n")
	if _, moved := stamp.Moved(); !moved {
		t.Error("a file changed did not move the stamp")
	}
	before := StampOf(c)
	if err := os.WriteFile(config.FileIn(h.Dir), []byte(homeIsDefault+"\n[focus]\nsession = \"1h\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, moved := before.Moved(); !moved {
		t.Error("the global file changed did not move the stamp")
	}
}

func TestConfigStamp_watchesTheThemeOfOnesOwnInForce(t *testing.T) {
	h := homeHolding(t, "[window]\ntheme = \"mine\"\n", nil)
	folder := filepath.Join(filepath.Dir(config.FileIn(h.Dir)), "themes")
	if err := os.MkdirAll(folder, 0o700); err != nil {
		t.Fatal(err)
	}
	scheme, err := os.ReadFile(filepath.Join("..", "theme", "schemes", "nord.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(folder, "mine.yaml"), scheme, 0o600); err != nil {
		t.Fatal(err)
	}
	stamp := StampOf(answered(t, h))
	if err := os.WriteFile(filepath.Join(folder, "mine.yaml"), append(scheme, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, moved := stamp.Moved(); !moved {
		t.Error("the scheme worn changed did not move the stamp")
	}
}

func TestTroubleIn_namesTheFileTheLineAndWhy(t *testing.T) {
	h := homeHolding(t, homeIsDefault, nil, "h")
	last := answered(t, h)
	path := settingsOf(t, h, "h", "# mine\n[pages]\n\ntime = \"hiden\"\n")
	broken := answered(t, h)
	trouble := TroubleIn(broken, last)
	if trouble.File != path || trouble.Line != 4 || !strings.Contains(trouble.Said, `"hiden"`) || strings.HasPrefix(trouble.Said, path) {
		t.Errorf("the trouble is %+v", trouble)
	}
	if got := trouble.String(); !strings.HasPrefix(got, "line 4: ") {
		t.Errorf("it says %q", got)
	}
	if err := os.WriteFile(config.FileIn(h.Dir), []byte("[[[\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if trouble := TroubleIn(answered(t, h), last); trouble.File != config.FileIn(h.Dir) || trouble.Line != 0 {
		t.Errorf("the global trouble is %+v", trouble)
	}
}

func TestSchemeOf_readsTheSchemeWhereConfigShowSaysItIs(t *testing.T) {
	h := homeHolding(t, "", nil)
	for _, seen := range answered(t, h).Themes {
		scheme, err := SchemeOf(seen)
		if err != nil || scheme.Slug != seen.Slug {
			t.Errorf("%s reads %+v, %v", seen.Slug, scheme.Slug, err)
		}
	}
	if _, err := SchemeOf(ThemeSeen{Slug: "mine", From: "/nowhere/mine.yaml"}); err == nil {
		t.Error("a scheme of one's own gone is read")
	}
}

func TestPlacesOfAnswer_sayWhereTheSettingsAre(t *testing.T) {
	h := homeHolding(t, homeIsDefault, nil, "h")
	if got := SettingsOf(answered(t, h)); got != filepath.Join(h.Dir, "h", config.SettingsFile) {
		t.Errorf("the settings are at %q", got)
	}
	if SettingsOf(Config{}) != "" {
		t.Error("an answer with no store has settings")
	}
}

func TestValueIn_readsEveryTableTheAnswerHolds(t *testing.T) {
	c := Config{
		Pages:      &PagesConfig{Tasks: true, Objectives: false, Time: "hidden"},
		Archive:    &Archive{Start: "03:00", Days: 9},
		Focus:      &FocusConfig{Session: "2h"},
		Approach:   &ApproachConfig{Label: "discuss"},
		Time:       &TimeConfig{Deleted: "kept", Weekends: "shown"},
		Window:     &WindowConfig{KeepSizes: true, Theme: "nord", Intro: "fade", Ghost: 70},
		Recurrence: &RecurrenceConfig{On: true},
		Workspace:  &WorkspaceConfig{Colour: "teal", Name: "mine"},
		Plugin:     Emplacement{At: "/p"},
	}
	for key, want := range map[string]string{
		"pages.tasks": "true", "pages.objectives": "false", "pages.time": "hidden",
		"archive.start": "03:00", "archive.days": "9", "focus.session": "2h", "approach.label": "discuss",
		"time.deleted": "kept", "time.weekends": "shown",
		"window.keep_sizes": "true", "window.theme": "nord", "window.intro": "fade", "window.ghost": "70",
		"series.on": "true", "workspace.colour": "teal", "workspace.name": "mine", "plugin": "/p",
	} {
		if got := ValueIn(c, key); got != want {
			t.Errorf("%s is %q, want %q", key, got, want)
		}
	}
	if got := ValueIn(Config{}, "pages.tasks"); got != "" {
		t.Errorf("an answer holding no table gives pages.tasks %q", got)
	}
}

func TestStampOf_watchesOnlyTheSchemeWornFromAFile(t *testing.T) {
	folder := t.TempDir()
	file, worn, other := filepath.Join(folder, "config.toml"), filepath.Join(folder, "worn.toml"), filepath.Join(folder, "other.toml")
	for _, path := range []string{file, worn, other} {
		if err := os.WriteFile(path, []byte("a"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	stamp := StampOf(Config{File: file, Themes: []ThemeSeen{{Slug: "o", From: other}, {Slug: "w", From: worn, InForce: true}}})
	if err := os.WriteFile(other, []byte("bb"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, moved := stamp.Moved(); moved {
		t.Error("a scheme not worn moved the stamp")
	}
	if err := os.WriteFile(worn, []byte("bb"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, moved := stamp.Moved(); !moved {
		t.Error("the scheme worn changed and the stamp did not move")
	}
}

func TestMoved_readsAgainOnlyWhatTheSizeOrTheTimeSaysChanged(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	at := time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC)
	wrote := func(said string, when time.Time) {
		t.Helper()
		if err := os.WriteFile(path, []byte(said), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(path, when, when); err != nil {
			t.Fatal(err)
		}
	}
	wrote("ab", at)
	stamp := StampOf(Config{File: path})
	if _, moved := stamp.Moved(); moved {
		t.Error("a file left alone moved the stamp")
	}
	wrote("cd", at)
	if _, moved := stamp.Moved(); moved {
		t.Error("a file of the same size and time was read again")
	}
	wrote("ab", at.Add(time.Hour))
	if _, moved := stamp.Moved(); moved {
		t.Error("a file saved with nothing changed moved the stamp")
	}
	wrote("cd", at.Add(time.Hour))
	if _, moved := stamp.Moved(); !moved {
		t.Error("a file of the same size saying something else did not move the stamp")
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if _, moved := stamp.Moved(); !moved {
		t.Error("a file gone did not move the stamp")
	}
}
