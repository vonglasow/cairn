package gui

import (
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"cairn/internal/config"
	"cairn/internal/engine"
)

func TestAToggleIsATrackOnTheControlOffAndOnTheAccentOn(t *testing.T) {
	s := Dark()
	box := Box{X: 10, Y: 10, W: toggleWide, H: toggleHigh}
	off, drawn := drawnAlone(t, 60, 40, func(k Kit) { Toggle{}.Draw(k, box) })
	if got := pixel(off, box.X+box.W-6, box.Y+box.H/2); got != s.Colour.Surface.ControlLit {
		t.Errorf("a toggle off is %v at its right", got)
	}
	if got := pixel(off, box.X+3+7, box.Y+box.H/2); got != s.Colour.Ink.Faint {
		t.Errorf("a toggle off has its knob %v at its left", got)
	}
	if pixel(off, box.X, box.Y) == s.Colour.Surface.ControlLit {
		t.Error("a toggle is square")
	}
	if want := []Drawing{{Part: "toggle", At: box, Layer: LayerBoard}}; !slices.Equal(drawn, want) {
		t.Errorf("the toggle recorded %+v", drawn)
	}
	on, _ := drawnAlone(t, 60, 40, func(k Kit) { Toggle{On: true}.Draw(k, box) })
	if got := pixel(on, box.X+6, box.Y+box.H/2); got != s.Colour.Accent.Base {
		t.Errorf("a toggle on is %v at its left", got)
	}
	if got, want := pixel(on, box.X+box.W-3-8, box.Y+box.H/2-1), s.Colour.Ink.OnAccent; abs(int(got.R)-int(want.R))+abs(int(got.G)-int(want.G))+abs(int(got.B)-int(want.B)) > 6 {
		t.Errorf("a toggle on has its knob %v at its right", got)
	}
}

func TestTheSettingsPanelIsPinnedAtBothSizes(t *testing.T) {
	for _, c := range []struct {
		wide, high int
		want       Box
		nav        bool
	}{{1200, 800, Box{X: 170, Y: 50, W: 860, H: 700}, true}, {720, 560, Box{X: 16, Y: 16, W: 688, H: 528}, false}} {
		got := SettingsBox(c.wide, c.high)
		if got != c.want {
			t.Errorf("at %d by %d the settings are %+v, want %+v", c.wide, c.high, got, c.want)
		}
		if nav := SettingsNav(got); (nav.W > 0) != c.nav {
			t.Errorf("at %d by %d the sections are %+v", c.wide, c.high, nav)
		}
	}
}

func TestEveryRowKeepsItsNameItsControlAndItsOriginApart(t *testing.T) {
	f := newFake()
	f.config = &engine.Config{File: "/h/.config/cairn/config.toml", Store: engine.Emplacement{At: "/h/.cairn/cairn.db"},
		Pages: &engine.PagesConfig{Tasks: true, Objectives: true, Time: "on"}, Archive: &engine.Archive{Days: 14},
		Focus: &engine.FocusConfig{}, Approach: &engine.ApproachConfig{}, Time: &engine.TimeConfig{Deleted: "kept", Weekends: "shown"},
		Window: &engine.WindowConfig{Theme: "default"}, Recurrence: &engine.RecurrenceConfig{}}
	for _, size := range [][2]int{{1200, 800}, {720, 560}} {
		w := holding(f).Resized(size[0], size[1]).openingConfiguration()
		rows := w.SettingRows()
		if len(rows) == 0 {
			t.Fatalf("at %v the settings hold no row", size)
		}
		for _, row := range rows {
			if row.Control.W > 0 && (cross(row.Name, row.Control) || cross(row.Control, row.From)) {
				t.Errorf("at %v the row %s lays %+v %+v %+v over each other", size, row.Key.Name, row.Name, row.Control, row.From)
			}
		}
	}
}

func settingsScreens(size [2]int) []struct {
	name string
	w    Window
} {
	c := engine.Config{Settings: "/h/.cairn/settings.toml", Written: []string{"pages.time", "archive.start"}, WrittenHere: []string{"window.theme", "pages.objectives", "workspace.colour"},
		Global: map[string]string{"window.theme": "default", "pages.objectives": "true"}, Window: &engine.WindowConfig{Theme: "nord"},
		Workspace: &engine.WorkspaceConfig{Colour: "green"}, Themes: []engine.ThemeSeen{{Slug: "default", From: "shipped"}, {Slug: "nord", From: "shipped", InForce: true}, {Slug: "dracula", From: "shipped"}, {Slug: "neon", From: "/nowhere/neon.yaml", Problem: "line 3: base0D is not a colour"}}}
	f := configuredWith(c)
	f.shown = &engine.WorkspaceShown{Workspace: "home", Store: "/h/.cairn/cairn.db", Colour: "green"}
	w := holding(f, card(1, "Fix the garden gate", "backlog")).Resized(size[0], size[1]).At(time.Now()).Look().ReadingConfiguration()
	open := w.openingConfiguration()
	refused := open
	refused.refusedKey, refused.refusedSay = "pages.time", `"hiden" is not a word pages.time takes (on, hidden, off)`
	writing := open
	writing.keyWriting, writing.keyWritten = "focus.session", Written{}.Types("1 hour")
	told := w
	told.changed, told.changedIn = []engine.ConfigChange{{Key: "pages.time", Was: "hidden", Now: "off"}, {Key: "window.theme", Was: "default", Now: "nord"}}, "settings.toml"
	troubled := w
	troubled.trouble = &engine.Trouble{File: "/h/.cairn/settings.toml", Line: 4, Said: `pages.time "hiden": the time is "on", "hidden" or "off"`}
	troubledPanel := troubled.openingConfiguration()
	troubledPanel.trouble = troubled.trouble
	return []struct {
		name string
		w    Window
	}{
		{"the settings of a workspace", open.Hovering(Point{X: 600, Y: 260})},
		{"the settings scrolled", open.scrollingSettings(400)},
		{"the global settings", open.scoped(true)},
		{"the global places", open.scoped(true).scrollingSettings(2000)},
		{"a row walked to", open.Key("down").Key("down")},
		{"a value refused", refused},
		{"a field written", writing},
		{"the theme chooser", open.openingChooser().Key("right")},
		{"the intro chooser", open.openingIntros().Key("down").Key("down")},
		{"the intro chooser walked down", open.openingIntros().Key("up").onIntro(len(config.Intros()) - 1)},
		{"the board told of a change", told},
		{"the board with a file it cannot read", troubled},
		{"the settings it cannot read", troubledPanel},
	}
}

func TestTheSettingsDrawNoPartOverItsNeighbourAndEveryTextWhereItFits(t *testing.T) {
	for _, size := range [][2]int{{720, 560}, {1200, 800}} {
		var wrong []string
		for _, s := range settingsScreens(size) {
			if s.w.Wrong() != nil {
				wrong = append(wrong, s.name+": "+s.w.Wrong().Error())
			}
			drawn := Drawn{all: partsDrawn(s.w)}
			for _, crossed := range drawn.Overlaps() {
				wrong = append(wrong, fmt.Sprintf("%s: %s %+v over %s %+v", s.name, crossed[0].Part, crossed[0].At, crossed[1].Part, crossed[1].At))
			}
			for _, stray := range drawn.Strays() {
				wrong = append(wrong, fmt.Sprintf("%s: %s %+v runs out of %s %+v", s.name, stray[0].Part, stray[0].At, stray[1].Part, stray[1].At))
			}
			for _, one := range textDrawn(s.w) {
				if one.box.W <= 0 || one.box.H <= 0 || one.high > one.box.H+textSlack {
					wrong = append(wrong, fmt.Sprintf("%s: %q in %+v", s.name, one.said, one.box))
				}
			}
		}
		if len(wrong) > 0 {
			t.Errorf("at %v, %d problem(s):\n%s", size, len(wrong), strings.Join(wrong, "\n"))
		}
	}
}
