package tui

import (
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

func (w *world) registerTree(sc *godog.ScenarioContext) {
	sc.Then(`^the objectives read (".*")$`, w.objectivesRead)
}

func (w *world) objectivesRead(list string) error {
	screen := w.screen()
	at := -1
	for _, name := range strings.Split(list, ",") {
		title := strings.Trim(strings.TrimSpace(name), `"`)
		found := strings.Index(screen, title)
		if found < 0 {
			return fmt.Errorf("the objectives never say %q:\n%s", title, screen)
		}
		if found < at {
			return fmt.Errorf("%q comes before what it should follow:\n%s", title, screen)
		}
		at = found
	}
	return nil
}
