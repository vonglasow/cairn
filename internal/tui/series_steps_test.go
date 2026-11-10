package tui

import (
	"fmt"
	"strconv"

	"github.com/cucumber/godog"
)

func (w *world) registerSeries(sc *godog.ScenarioContext) {
	sc.Given(`^recurrence is (on|off) on the board$`, func(said string) error {
		w.recurs = said == "on"
		return nil
	})
	sc.Then(`^the open task is task (\d+)$`, func(n int) error {
		if w.b.screen != OnDetail || w.b.task != n {
			return fmt.Errorf("the screen %v is open on task %d, not on task %d", w.b.screen, w.b.task, n)
		}
		return w.mustSay(strconv.Itoa(n))
	})
}
