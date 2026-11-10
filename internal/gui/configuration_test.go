package gui

import (
	"cairn/internal/build"
	"cairn/internal/engine"
	"slices"
	"strings"
	"testing"
)

func TestTheSystemOpensAFileWithItsOwnCommand(t *testing.T) {
	for _, c := range []struct {
		goos string
		name string
		args []string
	}{
		{"darwin", "open", []string{"-t", "/h/config.toml"}},
		{"linux", "xdg-open", []string{"/h/config.toml"}},
		{"windows", "cmd", []string{"/c", "start", "", "/h/config.toml"}},
	} {
		name, args := OpenCommand(c.goos, "/h/config.toml")
		if name != c.name || !slices.Equal(args, c.args) {
			t.Errorf("on %s the file is opened by %q %q, want %q %q", c.goos, name, args, c.name, c.args)
		}
	}
}

func configuredWith(c engine.Config) *fake {
	f := newFake()
	c.File = "/h/.config/cairn/config.toml"
	c.Store, c.Notes = engine.Emplacement{At: "/h/.cairn/cairn.db"}, engine.Emplacement{At: "/h/.cairn/notes"}
	c.Pages, c.Archive, c.Focus = &engine.PagesConfig{Tasks: true, Objectives: true, Time: "on"}, &engine.Archive{Days: 14}, &engine.FocusConfig{}
	c.Approach, c.Time, c.Recurrence = &engine.ApproachConfig{}, &engine.TimeConfig{Deleted: "kept", Weekends: "shown"}, &engine.RecurrenceConfig{}
	if c.Window == nil {
		c.Window = &engine.WindowConfig{Theme: "default"}
	}
	f.config = &c
	return f
}

func TestAPathIsCopiedByTheMarkBeforeIt(t *testing.T) {
	w := holding(configuredWith(engine.Config{})).openingConfiguration().toSection(SettingSection{Name: "Places", First: 13})
	marks := w.CopyMarks()
	if len(marks) != 2 {
		t.Fatalf("the configuration offers %d paths to copy: %+v", len(marks), marks)
	}
	store := marks[0]
	if got, _ := w.clicked(Point{X: store.At.X + store.At.W/2, Y: store.At.Y + store.At.H/2}).Copying(); got != "/h/.cairn/cairn.db" {
		t.Errorf("the store's mark copies %q", got)
	}
}

func TestTheAboutBoxCopiesWhereCairnIsKept(t *testing.T) {
	w := holding(newFake())
	w.screen, w.question = OnAsking, aboutCairn
	marks := w.CopyMarks()
	if len(marks) != 1 || marks[0].Said != build.Home {
		t.Fatalf("About offers to copy %+v", marks)
	}
}

func TestAnOpenTasksContextIsCopiedByItsMark(t *testing.T) {
	w := holding(newFake(), engine.Card{Number: 1, Title: "A", State: "todo", Context: "/h/notes/1"})
	w = w.clicked(middleOf(w, 1, 0)).Key("enter").Key("e")
	marks := w.CopyMarks()
	if len(marks) != 1 || marks[0].Said != "/h/notes/1" {
		t.Fatalf("an open task offers to copy %+v", marks)
	}
	if said := w.ContextSays(); len(said) != 2 || said[1].Box.X < marks[0].At.X+marks[0].At.W {
		t.Errorf("the context %+v is written over its mark %+v", said, marks[0].At)
	}
}

func TestAKeyTheStoresSettingsWriteIsSaidToComeFromThem(t *testing.T) {
	w := holding(configuredWith(engine.Config{Settings: "/h/.cairn/settings.toml", Written: []string{"focus.session", "time.weekends"}, WrittenHere: []string{"time.weekends"}})).openingConfiguration()
	from := map[string]string{}
	for _, row := range w.SettingRows() {
		from[row.Key.Name] = row.Comes
	}
	if from["time.weekends"] != SetHere || from["focus.session"] != FromGlobal || from["time.deleted"] != ByDefault {
		t.Errorf("the keys come from %v", from)
	}
}

func TestTheGlobalScopeSaysWhatTheWorkspaceSetsOfItsOwn(t *testing.T) {
	w := holding(configuredWith(engine.Config{Settings: "/h/.cairn/settings.toml", Written: []string{"time.weekends"}, WrittenHere: []string{"time.weekends"}, Global: map[string]string{"time.weekends": "shown"}})).openingConfiguration().scoped(true)
	for _, row := range w.SettingRows() {
		if row.Key.Name == "time.weekends" && (row.Comes != InFile || !strings.Contains(row.Theirs, "sets its own")) {
			t.Errorf("the weekends read %+v", row)
		}
		if row.Key.Store {
			t.Errorf("the global scope offers %s", row.Key.Name)
		}
	}
}

func TestTheConfigurationSaysTheSchemeTheWindowWears(t *testing.T) {
	w := holding(configuredWith(engine.Config{Window: &engine.WindowConfig{Theme: "nord"}, Settings: "/h/.cairn/settings.toml", WrittenHere: []string{"window.theme"}, Global: map[string]string{"window.theme": "default"}})).openingConfiguration()
	for _, row := range w.SettingRows() {
		if row.Key.Name == "window.theme" {
			if row.Value != "nord" || row.Comes != SetHere || row.Aside != "← global: default" {
				t.Errorf("the scheme is said %+v", row)
			}
			return
		}
	}
	t.Error("the configuration does not say the scheme the window wears")
}
