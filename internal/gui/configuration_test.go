package gui

import (
	"cairn/internal/board"
	"cairn/internal/engine/calendar"
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
	w := holding(configuredWith(workspace.Config{})).openingConfiguration().toSection(SettingSection{Name: files.SectionPlaces, First: 13})
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
	w := holding(f).openingConfiguration().Choose(OpenTheFile)
	if in, ok := f.gave("config", "fill"); !ok || in.Given["global"] == "true" {
		t.Errorf("opening the settings filled %+v (%v)", in, ok)
	}
	if path, asked := w.OpenAsked(); !asked || path != "/h/.cairn/settings.toml" {
		t.Errorf("the window opens %q (%v)", path, asked)
	}
}

func TestTheSettingsTakeWhatIsTypedOrPickedWithTheMouse(t *testing.T) {
	f := configuredWith(workspace.Config{Settings: "/h/.cairn/settings.toml"})
	w := holding(f).openingConfiguration()
	middle := func(b geom.Box) geom.Point { return geom.Point{X: b.X + b.W/2, Y: b.Y + b.H/2} }
	shown := func(w Window, name string) (Window, SettingRow) {
		for range 40 {
			for _, one := range w.SettingRows() {
				if one.Key.Name == name && SettingsRows(w.SettingsBox()).In(middle(one.Control)) {
					return w, one
				}
			}
			w = w.scrollingSettings(rowHigh)
		}
		t.Fatalf("no setting %s in sight", name)
		return w, SettingRow{}
	}
	for _, c := range []struct{ key, typed, want string }{
		{"workspace.name", "garden", "garden"},
		{"archive.days", "7", "7"},
		{"focus.session", "90m", "90m"},
		{"window.ghost", "50", "50"},
	} {
		at, row := shown(w, c.key)
		writing := at.Click(middle(row.Control))
		if writing.KeyWriting != c.key {
			t.Errorf("a click on %s writes %q", c.key, writing.KeyWriting)
			continue
		}
		if got := writing.SettingsFootSays(); got != WritingFoot {
			t.Errorf("writing %s, the foot says %q", c.key, got)
		}
		for range 12 {
			writing = writing.Key("backspace")
		}
		_, other := shown(writing, "pages.tasks")
		saved := writing.Types(c.typed).Click(middle(other.Name))
		if in, ok := f.gave("config", "set"); !ok || in.Given["key"] != c.key || in.Given["value"] != c.want || saved.KeyWriting != "" {
			t.Errorf("%s typed then a click elsewhere set %+v (%v), still writing %q", c.key, in.Given, ok, saved.KeyWriting)
		}
	}
	at, row := shown(w, "archive.start")
	dating := at.Click(middle(row.Control))
	if dating.Screen != board.OnCalendar {
		t.Fatalf("a click on archive.start opens %v", dating.Screen)
	}
	box := dating.CalendarBox()
	for i, day := range board.MonthGrid(dating.Day) {
		if day.Day() == 12 && day.Month() == dating.Day.Month() {
			dated := dating.Click(middle(DayAt(box, i)))
			if in, ok := f.gave("config", "set"); !ok || in.Given["key"] != "archive.start" || in.Given["value"] != calendar.FormatDay(day) || dated.Screen != board.OnConfiguration {
				t.Errorf("a day clicked set %+v (%v), back on %v", in.Given, ok, dated.Screen)
			}
		}
	}
}

func TestTheWorkspaceColourIsPickedByAClick(t *testing.T) {
	f := configuredWith(workspace.Config{Settings: "/h/.cairn/settings.toml"})
	w := holding(f).openingConfiguration()
	var colour SettingRow
	for _, row := range w.SettingRows() {
		if row.Key.Name == "workspace.colour" {
			colour = row
		}
	}
	opened := w.Click(geom.Point{X: colour.Control.X + colour.Control.W/2, Y: colour.Control.Y + colour.Control.H/2})
	swatches := opened.ColourSwatches()
	if len(swatches) != len(label.LabelColours) {
		t.Fatalf("a click on the colour opens %d swatches", len(swatches))
	}
	third := swatches[3].At
	picked := opened.Click(geom.Point{X: third.X + third.W/2, Y: third.Y + third.H/2})
	if in, ok := f.gave("config", "set"); !ok || in.Given["key"] != "workspace.colour" || in.Given["value"] != label.LabelColours[3] {
		t.Errorf("a click on a swatch set %+v (%v)", in.Given, ok)
	}
	if len(picked.ColourSwatches()) != 0 {
		t.Error("the swatches stay open once one is picked")
	}
	box := opened.SettingsBox()
	if left := opened.Click(geom.Point{X: box.X + 40, Y: box.Y + box.H - 60}); len(left.ColourSwatches()) != 0 || left.Screen != board.OnConfiguration {
		t.Errorf("a click beside the swatches leaves %d open, on %v", len(left.ColourSwatches()), left.Screen)
	}
}
