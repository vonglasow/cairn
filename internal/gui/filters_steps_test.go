package gui

import (
	"fmt"
	"slices"
	"strings"

	"github.com/cucumber/godog"
)

func (w *world) filtersSteps(sc *godog.ScenarioContext) {
	opening := func() {
		if w.w.Screen() != OnFiltering {
			w.w = w.w.Key("shift+f")
		}
		if _, open := w.w.FilterList(); open {
			w.w = w.w.Key("esc")
		}
		if w.w.Screen() != OnFiltering {
			w.w = w.w.Key("shift+f")
		}
	}
	sc.Step(`^I set the filter "([^"]*)" to "([^"]*)"$`, func(name, value string) error {
		opening()
		at := -1
		for i, f := range w.w.Filters() {
			if f.Name == name {
				at = i
			}
		}
		if at < 0 {
			return fmt.Errorf("no filter called %q among %v", name, w.w.Filters())
		}
		for w.w.FilterOn() < at {
			w.w = w.w.Key("down")
		}
		for w.w.FilterOn() > at {
			w.w = w.w.Key("up")
		}
		w.w = w.w.Key("enter")
		want := slices.Index(w.w.Filters()[at].Values, value)
		if want < 0 {
			want = slices.Index(w.w.Filters()[at].Shown, value)
		}
		if want < 0 {
			return fmt.Errorf("the filter %q offers no %q: %v", name, value, w.w.Filters()[at].Values)
		}
		for w.w.FilterPick() < want {
			w.w = w.w.Key("down")
		}
		for w.w.FilterPick() > want {
			w.w = w.w.Key("up")
		}
		w.w = w.w.Key("enter")
		if _, open := w.w.FilterList(); open {
			w.w = w.w.Key("esc")
		}
		return nil
	})
	sc.Then(`^the pills after the tabs read "([^"]*)"$`, func(want string) error {
		l := w.w.Layout()
		last := TabAt(l, len(w.w.Shown())-1)
		var said []string
		for _, pill := range w.w.Pills() {
			if pill.Box.X < last.X+last.W || pill.Box.Y < l.Tabs.Y || pill.Box.Y+pill.Box.H > l.Tabs.Y+l.Tabs.H {
				return fmt.Errorf("the pill %q stands at %+v, not in the tabs' band right of the last tab %+v", pill.Said, pill.Box, last)
			}
			said = append(said, pill.Said)
		}
		if got := strings.Join(said, ", "); got != want {
			return fmt.Errorf("the pills after the tabs read %q", got)
		}
		return nil
	})
	sc.Then(`^the pills end against the menu$`, func() error {
		pills := w.w.Pills()
		if len(pills) == 0 {
			return fmt.Errorf("no pill is drawn")
		}
		for i := 1; i < len(pills); i++ {
			if pills[i].Box.X != pills[i-1].Box.X+pills[i-1].Box.W+pillGap {
				return fmt.Errorf("the pill %q at %+v does not follow %q at %+v", pills[i].Said, pills[i].Box, pills[i-1].Said, pills[i-1].Box)
			}
		}
		last := pills[len(pills)-1].Box
		if menu := MenuButton(w.w.Layout()); last.X+last.W != menu.X-between {
			return fmt.Errorf("the pills end at %d, not against the menu at %d", last.X+last.W, menu.X)
		}
		return nil
	})
	sc.Then(`^the filter "([^"]*)" lists "([^"]*)"$`, func(name, want string) error {
		f := w.w.Filters()[w.w.FilterOn()]
		values, open := w.w.FilterList()
		if !open || f.Name != name {
			return fmt.Errorf("the filter %q lists nothing", name)
		}
		var got []string
		for i, value := range values {
			if value.H > 0 {
				got = append(got, f.Shown[i])
			}
		}
		if strings.Join(got, ", ") != want {
			return fmt.Errorf("the filter %q lists %q", name, strings.Join(got, ", "))
		}
		return nil
	})
	sc.Then(`^the objectives shown read "([^"]*)"$`, func(want string) error {
		var got []string
		for _, g := range w.w.Showing().Goals {
			got = append(got, g.Title)
		}
		if strings.Join(got, ", ") != want {
			return fmt.Errorf("the objectives shown read %q", strings.Join(got, ", "))
		}
		return nil
	})
	sc.Then(`^the pill "([^"]*)" is (not )?drawn faint$`, func(said, not string) error {
		for _, pill := range w.w.Pills() {
			if pill.Said == said {
				if pill.Faint == (not != "") {
					return fmt.Errorf("the pill %q is faint: %v", said, pill.Faint)
				}
				return nil
			}
		}
		return fmt.Errorf("no pill reads %q", said)
	})
	sc.Then(`^the tab "([^"]*)" counts "([^"]*)"$`, func(name, want string) error {
		for _, page := range w.w.Shown() {
			if Titled(Pages[page]) == name {
				if got := w.w.TabCount(page); got != want {
					return fmt.Errorf("the tab %q counts %q", name, got)
				}
				return nil
			}
		}
		return fmt.Errorf("no tab is called %q", name)
	})
	var noted Box
	pillAt := func(said string) (Box, error) {
		for _, pill := range w.w.Pills() {
			if pill.Said == said {
				return pill.Box, nil
			}
		}
		return Box{}, fmt.Errorf("no pill reads %q", said)
	}
	sc.When(`^I note where the pill "([^"]*)" stands$`, func(said string) error {
		at, err := pillAt(said)
		noted = at
		return err
	})
	sc.Then(`^the pill "([^"]*)" stands where it stood$`, func(said string) error {
		at, err := pillAt(said)
		if err == nil && at != noted {
			return fmt.Errorf("the pill %q stands at %+v, not at %+v", said, at, noted)
		}
		return err
	})
	sc.When(`^I click the pill "([^"]*)"$`, func(said string) error {
		for _, pill := range w.w.Pills() {
			if pill.Said == said {
				w.w = w.w.clicked(Point{X: pill.Box.X + pillPad + 2, Y: pill.Box.Y + pill.Box.H/2})
				return nil
			}
		}
		return fmt.Errorf("no pill reads %q", said)
	})
	sc.Then(`^the filter "([^"]*)" says "([^"]*)" keeps (\d+)$`, func(name, value string, n int) error {
		f := w.w.Filters()[w.w.FilterOn()]
		if f.Name != name {
			return fmt.Errorf("the filter open is %q", f.Name)
		}
		counts := w.w.FilterCounts()
		at := slices.Index(f.Shown, value)
		if at < 0 || at >= len(counts) || counts[at] != n {
			return fmt.Errorf("the filter %q counts %v for %v", name, counts, f.Shown)
		}
		return nil
	})
	sc.Step(`^the window shows the filters$`, func() error {
		if w.w.Screen() != OnFiltering {
			return fmt.Errorf("the window is on %v", w.w.Screen())
		}
		return nil
	})
	sc.Step(`^I clear every filter$`, func() error {
		opening()
		w.w = w.w.Key("alt+backspace")
		return nil
	})
	sc.Step(`^the board says it is filtered by "([^"]*)"$`, func(want string) error {
		if got := w.w.FilteredBy(); got != want {
			return fmt.Errorf("the board says it is filtered by %q", got)
		}
		return nil
	})
}
