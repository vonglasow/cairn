package gui

import (
	"cairn/internal/board"
	"cairn/internal/engine/label"
	"cairn/internal/engine/workspace"
	"cairn/internal/gui/geom"
	"slices"
	"testing"
	"time"

	"cairn/internal/store/files"
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

func configuredWith(c workspace.Config) *fake {
	f := newFake()
	c.File = "/h/.config/cairn/config.toml"
	c.Store, c.Notes = workspace.Emplacement{At: "/h/.cairn/cairn.db"}, workspace.Emplacement{At: "/h/.cairn/notes"}
	c.Pages, c.Archive, c.Focus = &workspace.PagesConfig{Tasks: true, Objectives: true, Time: "on"}, &workspace.Archive{Days: 14}, &workspace.FocusConfig{}
	c.Approach, c.Time, c.Recurrence = &workspace.ApproachConfig{}, &workspace.TimeConfig{Deleted: "kept", Weekends: "shown"}, &workspace.RecurrenceConfig{}
	if c.Window == nil {
		c.Window = &workspace.WindowConfig{Theme: "default"}
	}
	f.config = &c
	return f
}

func TestAPathIsCopiedByTheMarkBeforeIt(t *testing.T) {
	w := holding(configuredWith(workspace.Config{})).openingConfiguration().toSection(SettingSection{Name: "Places", First: 13})
	marks := w.CopyMarks()
	if len(marks) != 3 {
		t.Fatalf("the configuration offers %d paths to copy: %+v", len(marks), marks)
	}
	store := marks[0]
	if got, _ := w.clicked(geom.Point{X: store.At.X + store.At.W/2, Y: store.At.Y + store.At.H/2}).Copying(); got != "/h/.cairn/cairn.db" {
		t.Errorf("the store's mark copies %q", got)
	}
}

func TestTheSettingsTakeADayOnTheCalendarAndAColourOnTheSwatches(t *testing.T) {
	f := configuredWith(workspace.Config{Settings: "/h/.cairn/settings.toml"})
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
	if dating.Screen != board.OnCalendar {
		t.Fatalf("archive.start opens %v, want the calendar", dating.Screen)
	}
	dated := dating.dated(time.Date(2026, 11, 2, 0, 0, 0, 0, time.Local))
	if in, ok := f.gave("config", "set"); dated.Screen != board.OnConfiguration || !ok || in.Given["key"] != "archive.start" || in.Given["value"] != "2026-11-02" {
		t.Errorf("the day taken set %+v (%v), back on %v", in, ok, dated.Screen)
	}
	colouring := w.acting(row(w, "workspace.colour"), -1)
	if got := len(colouring.ColourSwatches()); got != len(label.LabelColours) {
		t.Fatalf("the colour opens %d swatches, want %d", got, len(label.LabelColours))
	}
	colouring.Key("right").Key("enter")
	if in, ok := f.gave("config", "set"); !ok || in.Given["key"] != "workspace.colour" || !slices.Contains(label.LabelColours, in.Given["value"]) {
		t.Errorf("the swatch taken set %+v (%v)", in, ok)
	}
}

func TestAClickOutsideTheSettingsClosesThemAskingFirstOnlyWhenSomethingTypedIsNotSaved(t *testing.T) {
	w := holding(configuredWith(workspace.Config{Settings: "/h/.cairn/settings.toml"})).openingConfiguration()
	box := w.SettingsBox()
	outside := geom.Point{X: box.X / 2, Y: box.Y + box.H/2}
	if closed := w.Click(outside); closed.Screen != board.OnTheBoard {
		t.Errorf("a click outside the settings leaves the window on %v", closed.Screen)
	}
	var typed Window
	for _, row := range w.SettingRows() {
		if row.Key.Kind == files.KindText {
			typed = w.acting(row, -1).Key("x")
			break
		}
	}
	asked := typed.Click(outside)
	if asked.Screen != board.OnAsking || !slices.Equal(asked.Offers(), []string{"Throw away", "Keep editing"}) {
		t.Fatalf("a click outside with a value typed leaves %v offering %v", asked.Screen, asked.Offers())
	}
	if kept := asked.Choose("Keep editing"); kept.Screen != board.OnConfiguration || kept.KeyWriting == "" {
		t.Errorf("keep editing goes to %v, writing %q", kept.Screen, kept.KeyWriting)
	}
	if thrown := asked.Choose("Throw away"); thrown.Screen != board.OnTheBoard {
		t.Errorf("throw away goes to %v", thrown.Screen)
	}
}

func TestOpeningAFileOfTheConfigurationFillsItFirst(t *testing.T) {
	f := configuredWith(workspace.Config{Settings: "/h/.cairn/settings.toml", File: "/c/config.toml"})
	w := holding(f).openingConfiguration().Choose(OpenSettings)
	if in, ok := f.gave("config", "fill"); !ok || in.Given["global"] == "true" {
		t.Errorf("opening the settings filled %+v (%v)", in, ok)
	}
	if path, asked := w.OpenAsked(); !asked || path != "/h/.cairn/settings.toml" {
		t.Errorf("the window opens %q (%v)", path, asked)
	}
}
