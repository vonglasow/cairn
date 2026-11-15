package gui

import (
	"cairn/internal/board"
	"cairn/internal/engine/workspace"
	"cairn/internal/gui/geom"
	"fmt"
	"strings"
	"testing"
	"time"

	"cairn/internal/store/files"
)

func settingsScreens(size [2]int) []struct {
	name string
	w    Window
} {
	c := workspace.Config{Settings: "/h/.cairn/settings.toml", Written: []string{"pages.time", "archive.start"}, WrittenHere: []string{"window.theme", "pages.objectives", "workspace.colour"},
		Global: map[string]string{"window.theme": "default", "pages.objectives": "true"}, Window: &workspace.WindowConfig{Theme: "nord"},
		Workspace: &workspace.Colouring{Colour: "green"}, Themes: []workspace.ThemeSeen{{Slug: "default", From: "shipped"}, {Slug: "nord", From: "shipped", InForce: true}, {Slug: "dracula", From: "shipped"}, {Slug: "neon", From: "/nowhere/neon.yaml", Problem: "line 3: base0D is not a colour"}}}
	f := configuredWith(c)
	f.shown = &workspace.Shown{Workspace: "home", Store: "/h/.cairn/cairn.db", Colour: "green"}
	w := holding(f, card(1, "Fix the garden gate", "backlog")).Resized(size[0], size[1]).At(time.Now()).Look().ReadingConfiguration()
	open := w.openingConfiguration()
	refused := open
	refused.RefusedKey, refused.RefusedSay = "pages.time", `"hiden" is not a word pages.time takes (on, hidden, off)`
	writing := open
	writing.KeyWriting, writing.KeyWritten = "focus.session", board.Written{}.Types("1 hour")
	told := w
	told.Changed, told.ChangedIn = []workspace.ConfigChange{{Key: "pages.time", Was: "hidden", Now: "off"}, {Key: "window.theme", Was: "default", Now: "nord"}}, "settings.toml"
	troubled := w
	troubled.Troubled = &workspace.Trouble{File: "/h/.cairn/settings.toml", Line: 4, Said: `pages.time "hiden": the time is "on", "hidden" or "off"`}
	troubledPanel := troubled.openingConfiguration()
	troubledPanel.Troubled = troubled.Troubled
	return []struct {
		name string
		w    Window
	}{
		{"the settings of a workspace", open.Hovering(geom.Point{X: 600, Y: 260})},
		{"the settings scrolled", open.scrollingSettings(400)},
		{"the global settings", open.scoped(true)},
		{"the global places", open.scoped(true).scrollingSettings(2000)},
		{"a row walked to", open.Key("down").Key("down")},
		{"a value refused", refused},
		{"a field written", writing},
		{"the theme chooser", open.openingChooser().Key("right")},
		{"the intro chooser", open.openingIntros().Key("down").Key("down")},
		{"the intro chooser walked down", open.openingIntros().Key("up").onIntro(len(files.Intros()) - 1)},
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
