package gui

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/cucumber/godog"

	"cairn/internal/config"
	"cairn/internal/theme"
)

var origins = map[string]string{
	"set here": SetHere, "from the global configuration": FromGlobal, "from its default": ByDefault,
	"from the file": InFile, "read when cairn opens": WhenOpened,
}

func (w *world) settingRow(key string) (SettingRow, error) {
	for _, row := range w.w.SettingRows() {
		if row.Key.Name == key {
			return row, nil
		}
	}
	return SettingRow{}, fmt.Errorf("the settings hold no %q in the scope %q", key, w.w.SettingsScope())
}

func (w *world) clickIn(b Box) {
	w.w = w.w.Hovering(Point{X: b.X + b.W/2, Y: b.Y + b.H/2})
	w.w = w.w.clicked(Point{X: b.X + b.W/2, Y: b.Y + b.H/2})
}

func (w *world) settingInSight(key string) (SettingRow, error) {
	for range 40 {
		row, err := w.settingRow(key)
		if err != nil {
			return row, err
		}
		region := SettingsRows(w.w.SettingsBox())
		if row.At.Y >= region.Y && row.At.Y+row.At.H <= region.Y+region.H {
			return row, nil
		}
		w.w = w.w.scrollingSettings(lineHigh)
	}
	return w.settingRow(key)
}

func (w *world) configurationSteps(sc *godog.ScenarioContext) {
	sc.Given(`^a configuration file naming the store$`, func() error {
		path := config.FileIn(w.home)
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			return err
		}
		return os.WriteFile(path, []byte(fmt.Sprintf("store = %q\n", w.path)), 0o600)
	})
	sc.Then(`^the configuration reads "([^"]*)" (set here|from the global configuration|from its default|from the file|read when cairn opens)$`, func(key, from string) error {
		row, err := w.settingRow(key)
		if err != nil {
			return err
		}
		if row.Comes != origins[from] {
			return fmt.Errorf("%q reads %q %s, not %s", key, row.Value, row.Comes, origins[from])
		}
		return nil
	})
	sc.Then(`^the configuration reads "([^"]*)" (set here|from the global configuration), (?:"([^"]*)"|nothing) beside it$`, func(key, from, aside string) error {
		row, err := w.settingRow(key)
		if err != nil {
			return err
		}
		if row.Comes != origins[from] || row.Aside != aside {
			return fmt.Errorf("%q reads %q %s, %q beside it", key, row.Value, row.Comes, row.Aside)
		}
		return nil
	})
	sc.Then(`^the panel reads the settings of "([^"]*)" over the global configuration$`, func(name string) error {
		want := filepath.Join(w.spaces.FolderOf(name), config.SettingsFile) + " over " + config.FileIn(w.spaces.Dir)
		if got := w.w.SettingsPath(); got != want {
			return fmt.Errorf("the panel reads %q, not %q", got, want)
		}
		return nil
	})
	sc.Then(`^the colour of the workspace reads "([^"]*)", beside its swatch$`, func(colour string) error {
		row, err := w.settingRow("workspace.colour")
		if err != nil {
			return err
		}
		if row.Value != colour || row.Key.Kind != config.KindColour {
			return fmt.Errorf("the colour reads %q", row.Value)
		}
		return nil
	})
	sc.Then(`^the time reads "([^"]*)" of (".*")$`, func(on, list string) error {
		row, err := w.settingRow("pages.time")
		if err != nil {
			return err
		}
		if row.Value != on || !slices.Equal(row.Words, quoted(list)) || len(row.Parts) != len(row.Words) {
			return fmt.Errorf("the time reads %q of %q", row.Value, row.Words)
		}
		return nil
	})
	sc.Then(`^the settings are written in "([^"]*)"$`, func(scope string) error {
		if got := w.w.SettingsScope(); got != scope {
			return fmt.Errorf("the settings are written in %q, not %q", got, scope)
		}
		return nil
	})
	sc.When(`^I choose the scope "([^"]*)"$`, func(scope string) error {
		for i, said := range w.w.scopeWords() {
			if said == scope {
				w.clickIn(w.w.ScopeSegments()[i])
				return nil
			}
		}
		return fmt.Errorf("no scope %q: %v", scope, w.w.scopeWords())
	})
	sc.When(`^I click the (?:switch|control) of "([^"]*)"$`, func(key string) error {
		row, err := w.settingInSight(key)
		if err != nil {
			return err
		}
		w.clickIn(row.Control)
		return nil
	})
	sc.When(`^I click "([^"]*)" on the row "([^"]*)"$`, func(said, key string) error {
		row, err := w.settingInSight(key)
		if err != nil {
			return err
		}
		if at := slices.Index(row.Words, said); at >= 0 {
			w.clickIn(row.Parts[at])
			return nil
		}
		w.w = w.w.Hovering(Point{X: row.Name.X + 2, Y: row.Name.Y + 2})
		if row, err = w.settingRow(key); err != nil {
			return err
		}
		if at := slices.Index(row.Offers, said); at >= 0 {
			w.clickIn(row.OfferAt[at])
			return nil
		}
		return fmt.Errorf("the row %q offers %v and %v", key, row.Words, row.Offers)
	})
	sc.When(`^I write "([^"]*)" in the row "([^"]*)"$`, func(said, key string) error {
		row, err := w.settingInSight(key)
		if err != nil {
			return err
		}
		w.clickIn(row.Control)
		w.w = w.w.Types(said).Key("enter")
		return nil
	})
	sc.Then("^the row \"([^\"]*)\" says `([^`]*)`$", func(key, said string) error {
		row, err := w.settingRow(key)
		if err != nil {
			return err
		}
		if row.Said != said {
			return fmt.Errorf("the row %q says %q", key, row.Said)
		}
		return nil
	})
	sc.Then(`^the row "([^"]*)" offers nothing to change$`, func(key string) error {
		row, err := w.settingInSight(key)
		if err != nil {
			return err
		}
		before := w.w.Configured()
		w.clickIn(row.Control)
		if !row.Fixed || len(row.Offers) > 0 || w.w.Configured().File != before.File || w.w.keyWriting != "" {
			return fmt.Errorf("the row %q offers %+v", key, row)
		}
		return nil
	})
	sc.Then(`^the row "([^"]*)" is not offered$`, func(key string) error {
		row, err := w.settingRow(key)
		if err != nil {
			return err
		}
		if !row.Off || len(row.Offers) > 0 {
			return fmt.Errorf("the row %q is offered: %+v", key, row)
		}
		return nil
	})
	sc.When(`^I walk the chooser to "([^"]*)"$`, func(slug string) error {
		for range 40 {
			if w.w.WornScheme() == slug {
				return nil
			}
			w.w = w.w.Key("right")
		}
		return fmt.Errorf("the chooser never reached %q: it is on %q", slug, w.w.WornScheme())
	})
	sc.Then(`^the window wears "([^"]*)"$`, func(slug string) error {
		scheme, err := theme.Find("", slug)
		if err != nil {
			return err
		}
		if got := w.w.Skin().Colour.Surface.Back; got != scheme.Colours().Surface.Back {
			return fmt.Errorf("the window wears %v, %q wears %v (%s)", got, slug, scheme.Colours().Surface.Back, w.w.WornScheme())
		}
		return nil
	})
	sc.Then(`^the banner says "([^"]*)"$`, func(said string) error {
		var all []string
		board := w.w
		board.screen = OnTheBoard
		for _, b := range board.BoardBanners() {
			all = append(all, b.Said, b.Then)
		}
		if !strings.Contains(strings.Join(all, "\n"), said) {
			return fmt.Errorf("the window says %q", all)
		}
		return nil
	})
	sc.Then(`^the banner says nothing of a change$`, func() error {
		if _, changes := w.w.ReadAgain(); len(changes) > 0 {
			return fmt.Errorf("the window says %v changed", changes)
		}
		return nil
	})
	sc.Then(`^the banner says nothing of a file that cannot be read$`, func() error {
		if t, ok := w.w.Trouble(); ok {
			return fmt.Errorf("the window says %+v", t)
		}
		return nil
	})
	sc.Then(`^the window asks the system to open the configuration file$`, func() error {
		got, asked := w.w.OpenAsked()
		if !asked {
			return errors.New("the window asked the system to open nothing")
		}
		if want := config.FileIn(w.home); got != want {
			return fmt.Errorf("the window asks the system to open %q, not %q", got, want)
		}
		return nil
	})
	sc.Then(`^a configuration file exists$`, func() error {
		if _, err := os.Stat(config.FileIn(w.home)); err != nil {
			return fmt.Errorf("no configuration file: %w", err)
		}
		return nil
	})
}
