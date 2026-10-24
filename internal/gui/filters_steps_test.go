package gui

import (
	"fmt"

	"github.com/cucumber/godog"
)

func (w *world) filtersSteps(sc *godog.ScenarioContext) {
	opening := func() {
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
		for range len(w.w.Filters()[at].Values) {
			if w.w.Filters()[at].Value() == value {
				return nil
			}
			w.w = w.w.Key("right")
		}
		return fmt.Errorf("the filter %q never reaches %q", name, value)
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
		l := w.w.Layout()
		said, last := FilteredSaid(l), TabAt(l, len(w.w.Shown())-1)
		if said.X < last.X+last.W || said.Y < l.Tabs.Y || said.Y+said.H > l.Tabs.Y+l.Tabs.H || said.W <= 0 {
			return fmt.Errorf("what filters the board is written at %+v, not in the tabs' band right of the last tab %+v", said, last)
		}
		return nil
	})
}
