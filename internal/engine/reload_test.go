package engine

import (
	"cairn/internal/engine/workspace"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"cairn/internal/store/files"
)

func answered(t *testing.T, h Home) workspace.Config {
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
	if p := PagesOf(workspace.Config{}); !p.Shows(PageTasks) || p.Recurs() || p.WeekDays() != 7 {
		t.Errorf("an answer saying nothing reads %+v", p)
	}
}

func TestConfigChanges_namesEachKeyThatMovedAndFromWhat(t *testing.T) {
	h := homeHolding(t, homeIsDefault, nil, "h")
	was := answered(t, h)
	settingsOf(t, h, "h", "[pages]\ntime = \"off\"\n[window]\ntheme = \"nord\"\n")
	now := answered(t, h)
	changes := workspace.ConfigChanges(was, now)
	var said []string
	for _, c := range changes {
		said = append(said, c.String())
	}
	if !slices.Equal(said, []string{"window.theme default → nord", "pages.time on → off"}) {
		t.Errorf("the changes are %v", said)
	}
	if workspace.ConfigChanges(now, now) != nil {
		t.Error("an answer changed from itself")
	}
}

func TestConfigStamp_watchesTheThemeOfOnesOwnInForce(t *testing.T) {
	h := homeHolding(t, "[window]\ntheme = \"mine\"\n", nil)
	folder := filepath.Join(filepath.Dir(files.FileIn(h.Dir)), "themes")
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
	stamp := workspace.StampOf(answered(t, h))
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
	trouble := workspace.TroubleIn(broken, last)
	if trouble.File != path || trouble.Line != 4 || !strings.Contains(trouble.Said, `"hiden"`) || strings.HasPrefix(trouble.Said, path) {
		t.Errorf("the trouble is %+v", trouble)
	}
	if got := trouble.String(); !strings.HasPrefix(got, "line 4: ") {
		t.Errorf("it says %q", got)
	}
	if err := os.WriteFile(files.FileIn(h.Dir), []byte("[[[\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if trouble := workspace.TroubleIn(answered(t, h), last); trouble.File != files.FileIn(h.Dir) || trouble.Line != 0 {
		t.Errorf("the global trouble is %+v", trouble)
	}
}

func TestSchemeOf_readsTheSchemeWhereConfigShowSaysItIs(t *testing.T) {
	h := homeHolding(t, "", nil)
	for _, seen := range answered(t, h).Themes {
		scheme, err := workspace.SchemeOf(seen)
		if err != nil || scheme.Slug != seen.Slug {
			t.Errorf("%s reads %+v, %v", seen.Slug, scheme.Slug, err)
		}
	}
	if _, err := workspace.SchemeOf(workspace.ThemeSeen{Slug: "mine", From: "/nowhere/mine.yaml"}); err == nil {
		t.Error("a scheme of one's own gone is read")
	}
}

func TestPlacesOfAnswer_sayWhereTheSettingsAre(t *testing.T) {
	h := homeHolding(t, homeIsDefault, nil, "h")
	if got := workspace.SettingsOf(answered(t, h)); got != filepath.Join(h.Dir, "h", files.SettingsFile) {
		t.Errorf("the settings are at %q", got)
	}
	if workspace.SettingsOf(workspace.Config{}) != "" {
		t.Error("an answer with no store has settings")
	}
}

func TestValueIn_readsEveryTableTheAnswerHolds(t *testing.T) {
	c := workspace.Config{
		Pages:      &workspace.PagesConfig{Tasks: true, Objectives: false, Time: "hidden"},
		Archive:    &workspace.Archive{Start: "03:00", Days: 9},
		Focus:      &workspace.FocusConfig{Session: "2h"},
		Approach:   &workspace.ApproachConfig{Label: "discuss"},
		Time:       &workspace.TimeConfig{Deleted: "kept", Weekends: "shown"},
		Window:     &workspace.WindowConfig{KeepSizes: true, Theme: "nord", Intro: "fade", Ghost: 70},
		Recurrence: &workspace.RecurrenceConfig{On: true},
		Workspace:  &workspace.Colouring{Colour: "teal", Name: "mine"},
		Plugin:     workspace.Emplacement{At: "/p"},
	}
	for key, want := range map[string]string{
		"pages.tasks": "true", "pages.objectives": "false", "pages.time": "hidden",
		"archive.start": "03:00", "archive.days": "9", "focus.session": "2h", "approach.label": "discuss",
		"time.deleted": "kept", "time.weekends": "shown",
		"window.keep_sizes": "true", "window.theme": "nord", "window.intro": "fade", "window.ghost": "70",
		"series.on": "true", "workspace.colour": "teal", "workspace.name": "mine", "plugin": "/p",
	} {
		if got := workspace.ValueIn(c, key); got != want {
			t.Errorf("%s is %q, want %q", key, got, want)
		}
	}
	if got := workspace.ValueIn(workspace.Config{}, "pages.tasks"); got != "" {
		t.Errorf("an answer holding no table gives pages.tasks %q", got)
	}
}
