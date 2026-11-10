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
	h := homeHolding(t, "[pages]\nobjectives = false\ntime = \"hidden\"\n[time]\nweekends = \"hidden\"\n[recurrence]\non = true\n", nil)
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
