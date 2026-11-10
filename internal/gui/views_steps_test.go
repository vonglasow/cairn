package gui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cucumber/godog"
)

func (w *world) viewsSteps(sc *godog.ScenarioContext) {
	sc.Then(`^the name proposed for the view is "([^"]*)"$`, func(want string) error {
		if w.w.Screen() != OnKeeping {
			return fmt.Errorf("nothing is being kept as a view: the screen is %v", w.w.Screen())
		}
		if got := w.w.Naming().Reads(); got != want {
			return fmt.Errorf("the name proposed is %q", got)
		}
		return nil
	})
	sc.Then(`^the filter menu lists the views (".*")$`, func(list string) error {
		var said []string
		for _, row := range w.w.ViewRows() {
			said = append(said, strconv.Quote(row.Name+" "+strconv.Itoa(row.Keeps)+" "+w.w.keyboard.Spelled(Pretty(row.Key))))
		}
		if got := strings.Join(said, ", "); got != list {
			return fmt.Errorf("the filter menu lists the views %s", got)
		}
		return nil
	})
	sc.Then(`^the keyboard is on the view "([^"]*)"$`, func(name string) error {
		for _, row := range w.w.ViewRows() {
			if row.On {
				if row.Name != name {
					return fmt.Errorf("the keyboard is on the view %q", row.Name)
				}
				return nil
			}
		}
		return fmt.Errorf("the keyboard is on no view: %+v", w.w.ViewRows())
	})
	sc.When(`^I click the view "([^"]*)"$`, func(name string) error {
		for _, row := range w.w.ViewRows() {
			if row.Name == name {
				w.clickAt(row.At)
				return nil
			}
		}
		return fmt.Errorf("the filter menu lists no view %q", name)
	})
	sc.Then(`^the pill of the view is (not )?marked changed$`, func(not string) error {
		for _, pill := range w.w.Pills() {
			if pill.Kind == PillView {
				if pill.Changed == (not != "") {
					return fmt.Errorf("the pill of the view is marked changed: %v", pill.Changed)
				}
				return nil
			}
		}
		return fmt.Errorf("no pill stands for a view: %+v", w.w.Pills())
	})
	sc.Then(`^the views panel lists (".*")$`, func(list string) error {
		if w.w.Screen() != OnViews {
			return fmt.Errorf("the views panel is not open: the screen is %v", w.w.Screen())
		}
		var said []string
		for _, row := range w.w.ViewsPanelRows() {
			said = append(said, strconv.Quote(row.Name))
		}
		if got := strings.Join(said, ", "); got != list {
			return fmt.Errorf("the views panel lists %s", got)
		}
		return nil
	})
	sc.Then(`^the views panel warns "([^"]*)"$`, func(want string) error {
		said, _, ok := w.w.ViewsWarning()
		if !ok || said != want {
			return fmt.Errorf("the views panel warns %q (%v)", said, ok)
		}
		return nil
	})
}
