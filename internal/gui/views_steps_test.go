package gui

import (
	"cairn/internal/board"
	"fmt"
	"strconv"
	"strings"

	"github.com/cucumber/godog"
)

func (w *world) viewsSteps(sc *godog.ScenarioContext) {
	sc.Then(`^the name proposed for the view is "([^"]*)"$`, func(want string) error {
		if w.w.Screen != board.OnKeeping {
			return fmt.Errorf("nothing is being kept as a view: the screen is %v", w.w.Screen)
		}
		if got := w.w.Naming().Reads(); got != want {
			return fmt.Errorf("the name proposed is %q", got)
		}
		return nil
	})
	sc.Then(`^the views under the Tasks tab list (".*")$`, func(list string) error {
		var said []string
		for _, line := range w.w.ViewingLines() {
			said = append(said, strconv.Quote(line.Name+" "+line.Count+" "+w.w.keyboard.Spelled(Pretty(line.Key))))
		}
		if got := strings.Join(said, ", "); got != list {
			return fmt.Errorf("the views under the Tasks tab list %s", got)
		}
		return nil
	})
	sc.When(`^I click the view "([^"]*)"$`, func(name string) error {
		for _, line := range w.w.ViewingLines() {
			if line.Name == name {
				w.clickAt(line.At)
				return nil
			}
		}
		return fmt.Errorf("the views under the Tasks tab list no %q", name)
	})
	sc.When(`^I click "([^"]*)" under the views$`, func(said string) error {
		for _, act := range w.w.ViewingActs() {
			if act.Said == said {
				w.clickAt(act.At)
				return nil
			}
		}
		return fmt.Errorf("the foot of the views offers no %q", said)
	})
	sc.Then(`^the tab "Tasks" reads "([^"]*)"$`, func(want string) error {
		if got := w.w.TabSaid(board.Tasks); got != want {
			return fmt.Errorf("the tab Tasks reads %q", got)
		}
		return nil
	})
	sc.Then(`^the tab "Tasks" (does not mark|marks) the view changed$`, func(marks string) error {
		if got := w.w.TabChanged(board.Tasks); got != (marks == "marks") {
			return fmt.Errorf("the tab Tasks marks the view changed: %v", got)
		}
		return nil
	})
	sc.Then(`^the views panel lists (".*")$`, func(list string) error {
		if w.w.Screen != board.OnViews {
			return fmt.Errorf("the views panel is not open: the screen is %v", w.w.Screen)
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
