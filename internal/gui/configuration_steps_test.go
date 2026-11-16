package gui

import (
	"cairn/internal/gui/geom"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/cucumber/godog"

	"cairn/internal/board"
	"cairn/internal/store/files"
	"cairn/internal/theme"
)

var origins = map[string]string{
	"only here": SetHere, "from all workspaces": FromGlobal, "from its default": ByDefault,
	"from the file": InFile, "read when cairn opens": WhenOpened, "in its folder": InItsFolder,
}

func (w *world) settingRow(key string) (SettingRow, error) {
	for _, row := range w.w.SettingRows() {
		if row.Key.Name == key {
			return row, nil
		}
	}
	return SettingRow{}, fmt.Errorf("the settings hold no %q in the scope %q", key, w.w.SettingsScope())
}

func (w *world) clickIn(b geom.Box) {
	w.w = w.w.Hovering(geom.Point{X: b.X + b.W/2, Y: b.Y + b.H/2})
	w.w = w.w.clicked(geom.Point{X: b.X + b.W/2, Y: b.Y + b.H/2})
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
	sc.Given(`^a configuration file that names no workspace$`, func() error {
		path := files.FileIn(w.home)
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			return err
		}
		return os.WriteFile(path, []byte("[pages]\ntasks = true\n"), 0o600)
	})
	sc.Then(`^the configuration reads "([^"]*)" (only here|from all workspaces|from its default|from the file|read when cairn opens|in its folder)$`, func(key, from string) error {
		row, err := w.settingRow(key)
		if err != nil {
			return err
		}
		if row.Comes != origins[from] {
			return fmt.Errorf("%q reads %q %s, not %s", key, row.Value, row.Comes, origins[from])
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
		if at := slices.Index(row.Shown(), said); at >= 0 {
			w.clickIn(row.Parts[at])
			return nil
		}
		w.w = w.w.Hovering(geom.Point{X: row.Name.X + 2, Y: row.Name.Y + 2})
		if row, err = w.settingRow(key); err != nil {
			return err
		}
		if at := slices.Index(row.Offers, said); at >= 0 {
			w.clickIn(row.OfferAt[at])
			return nil
		}
		if left := row.Left; left != nil && slices.Contains([]string{NotesLeave, NotesMoveThem}, said) {
			w.clickIn(map[string]geom.Box{NotesLeave: left.Leave, NotesMoveThem: left.Move}[said])
			return nil
		}
		return fmt.Errorf("the row %q offers %v and %v", key, row.Words, row.Offers)
	})
	sc.Then(`^the sections read (.*)$`, func(list string) error {
		var got []string
		for _, section := range w.w.SettingSections() {
			got = append(got, section.Name)
		}
		if want := quoted(list); !slices.Equal(got, want) {
			return fmt.Errorf("the sections read %q, not %q", got, want)
		}
		return nil
	})
	sc.Then(`^the row "([^"]*)" is named "([^"]*)" over "([^"]*)"$`, func(key, name, under string) error {
		row, err := w.settingInSight(key)
		if err != nil {
			return err
		}
		if said, then := w.w.RowLines(row); said != name || then != under {
			return fmt.Errorf("the row %q is named %q over %q", key, said, then)
		}
		return nil
	})
	sc.Then(`^the row "([^"]*)" shows (\d+) %$`, func(key string, want int) error {
		row, err := w.settingInSight(key)
		if err != nil {
			return err
		}
		if got := w.w.PercentShown(row); got != want || len(row.Parts) != 1 {
			return fmt.Errorf("the row %q shows %d %% on %d parts", key, got, len(row.Parts))
		}
		return nil
	})
	sc.When(`^I drag the slider of "([^"]*)" to (\d+) %$`, func(key string, to int) error {
		row, err := w.settingInSight(key)
		if err != nil || len(row.Parts) != 1 {
			return fmt.Errorf("the row %q has no slider: %w", key, err)
		}
		track := row.Parts[0]
		at := func(percent int) geom.Point {
			return geom.Point{X: track.X + (percent-files.LeastGhost)*track.W/(files.MostGhost-files.LeastGhost), Y: track.Y + track.H/2}
		}
		w.w = w.w.PressAt(at(65)).MoveTo(at(55)).MoveTo(at(to)).LetGo(at(to))
		return nil
	})
	sc.When(`^I walk the settings to "([^"]*)"$`, func(key string) error {
		for range len(files.Keys()) {
			if row, err := w.settingRow(key); err == nil && w.w.RowOn >= 0 && w.w.SettingRows()[w.w.RowOn].Key.Name == row.Key.Name {
				return nil
			}
			w.w = w.w.Key("down")
		}
		return fmt.Errorf("the keyboard does not reach %q", key)
	})
	sc.When(`^I rest the pointer on the row "([^"]*)"$`, func(key string) error {
		row, err := w.settingInSight(key)
		if err != nil {
			return err
		}
		w.w = w.w.Hovering(geom.Point{X: row.Name.X + 2, Y: row.Name.Y + 2})
		return nil
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
	sc.When(`^I write the folder "([^"]*)" in the row "([^"]*)"$`, func(folder, key string) error {
		row, err := w.settingInSight(key)
		if err != nil {
			return err
		}
		w.clickIn(row.Control)
		w.w = w.w.Types(filepath.Join(w.spaces.Dir, folder)).Key("enter")
		return nil
	})
	sc.Then(`^the row "([^"]*)" warns (nothing|"[^"]*")$`, func(key, said string) error {
		row, err := w.settingRow(key)
		if err != nil {
			return err
		}
		switch {
		case said == "nothing" && row.Left != nil:
			return fmt.Errorf("the row %q warns %q", key, row.Left.Said)
		case said != "nothing" && (row.Left == nil || !strings.Contains(row.Left.Said, strings.Trim(said, `"`))):
			return fmt.Errorf("the row %q warns %+v", key, row.Left)
		}
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
		before := w.w.Configured
		w.clickIn(row.Control)
		if !row.Fixed || len(row.Offers) > 0 || w.w.Configured.File != before.File || w.w.KeyWriting != "" {
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
	sc.When(`^I walk the intro chooser to "([^"]*)"$`, func(name string) error {
		for range 20 {
			if w.w.IntroShown() == name {
				return nil
			}
			if err := w.press("down"); err != nil {
				return err
			}
		}
		return fmt.Errorf("the intro chooser never reached %q: it is on %q", name, w.w.IntroShown())
	})
	sc.Then(`^the board is arriving by "([^"]*)"$`, func(name string) error {
		if _, on := w.w.Intro(); !on || w.w.IntroShown() != name {
			return fmt.Errorf("the board is arriving by %q: %v", w.w.IntroShown(), on)
		}
		return nil
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
		looked := w.w
		looked.Screen = board.OnTheBoard
		for _, b := range looked.BoardBanners() {
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
		if want := files.FileIn(w.home); got != want {
			return fmt.Errorf("the window asks the system to open %q, not %q", got, want)
		}
		return nil
	})
	sc.Then(`^a configuration file exists$`, func() error {
		if _, err := os.Stat(files.FileIn(w.home)); err != nil {
			return fmt.Errorf("no configuration file: %w", err)
		}
		return nil
	})
}
