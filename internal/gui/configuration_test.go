package gui

import (
	"slices"
	"strings"
	"testing"
	"time"

	"cairn/internal/build"
	"cairn/internal/config"
	"cairn/internal/domain"
	"cairn/internal/engine"
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
	if len(marks) != 3 {
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

func TestTheSettingsTakeADayOnTheCalendarAndAColourOnTheSwatches(t *testing.T) {
	f := configuredWith(engine.Config{Settings: "/h/.cairn/settings.toml"})
	w := holding(f).openingConfiguration()
	row := func(w Window, name string) SettingRow {
		for _, one := range w.SettingRows() {
			if one.Key.Name == name {
				return one
			}
		}
		t.Fatalf("no setting %s among %d rows", name, len(w.SettingRows()))
		return SettingRow{}
	}
	dating := w.acting(row(w, "archive.start"), -1)
	if dating.Screen() != OnCalendar {
		t.Fatalf("archive.start opens %v, want the calendar", dating.Screen())
	}
	dated := dating.dated(time.Date(2026, 11, 2, 0, 0, 0, 0, time.Local))
	if in, ok := f.gave("config", "set"); dated.Screen() != OnConfiguration || !ok || in.Given["key"] != "archive.start" || in.Given["value"] != "2026-11-02" {
		t.Errorf("the day taken set %+v (%v), back on %v", in, ok, dated.Screen())
	}
	colouring := w.acting(row(w, "workspace.colour"), -1)
	if got := len(colouring.ColourSwatches()); got != len(domain.LabelColours) {
		t.Fatalf("the colour opens %d swatches, want %d", got, len(domain.LabelColours))
	}
	colouring.Key("right").Key("enter")
	if in, ok := f.gave("config", "set"); !ok || in.Given["key"] != "workspace.colour" || !slices.Contains(domain.LabelColours, in.Given["value"]) {
		t.Errorf("the swatch taken set %+v (%v)", in, ok)
	}
}

func TestAClickOutsideTheSettingsClosesThemAskingFirstOnlyWhenSomethingTypedIsNotSaved(t *testing.T) {
	w := holding(configuredWith(engine.Config{Settings: "/h/.cairn/settings.toml"})).openingConfiguration()
	box := w.SettingsBox()
	outside := Point{X: box.X / 2, Y: box.Y + box.H/2}
	if closed := w.Click(outside); closed.Screen() != OnTheBoard {
		t.Errorf("a click outside the settings leaves the window on %v", closed.Screen())
	}
	var typed Window
	for _, row := range w.SettingRows() {
		if row.Key.Kind == config.KindText {
			typed = w.acting(row, -1).Key("x")
			break
		}
	}
	asked := typed.Click(outside)
	if asked.Screen() != OnAsking || !slices.Equal(asked.Offers(), []string{"Throw away", "Keep editing"}) {
		t.Fatalf("a click outside with a value typed leaves %v offering %v", asked.Screen(), asked.Offers())
	}
	if kept := asked.Choose("Keep editing"); kept.Screen() != OnConfiguration || kept.keyWriting == "" {
		t.Errorf("keep editing goes to %v, writing %q", kept.Screen(), kept.keyWriting)
	}
	if thrown := asked.Choose("Throw away"); thrown.Screen() != OnTheBoard {
		t.Errorf("throw away goes to %v", thrown.Screen())
	}
}

func TestOpeningAFileOfTheConfigurationFillsItFirst(t *testing.T) {
	f := configuredWith(engine.Config{Settings: "/h/.cairn/settings.toml", File: "/c/config.toml"})
	w := holding(f).openingConfiguration().Choose(OpenSettings)
	if in, ok := f.gave("config", "fill"); !ok || in.Given["global"] == "true" {
		t.Errorf("opening the settings filled %+v (%v)", in, ok)
	}
	if path, asked := w.OpenAsked(); !asked || path != "/h/.cairn/settings.toml" {
		t.Errorf("the window opens %q (%v)", path, asked)
	}
}
