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
			if pill.Kind == PillCount {
				continue
			}
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
	sc.Then(`^the board says it keeps "([^"]*)"$`, func(want string) error {
		for _, pill := range w.w.Pills() {
			if pill.Kind == PillCount {
				if pill.Said != want {
					return fmt.Errorf("the board says it keeps %q", pill.Said)
				}
				return nil
			}
		}
		return fmt.Errorf("the board says nothing of what it keeps")
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
		w.w = w.w.Key("backspace")
		return nil
	})
	sc.Step(`^the board says it is filtered by "([^"]*)"$`, func(want string) error {
		if got := w.w.FilteredBy(); got != want {
			return fmt.Errorf("the board says it is filtered by %q", got)
		}
		return nil
	})
}
