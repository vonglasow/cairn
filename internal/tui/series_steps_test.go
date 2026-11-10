package tui

import (
	"fmt"
	"strconv"

	"github.com/cucumber/godog"

	"cairn/internal/engine"
	"cairn/internal/task"
)

func (w *world) registerSeries(sc *godog.ScenarioContext) {
	sc.Given(`^I added a task titled "([^"]*)" coming back "([^"]*)"$`, func(title, every string) error {
		_, err := w.other.Ask(engine.TaskNew, engine.Input{Given: map[string]string{"title": title, "every": every}})
		return err
	})
	sc.Given(`^I stopped series (\d+)$`, func(n int) error {
		_, err := w.other.Ask(engine.SeriesStop, engine.Numbered(n))
		return err
	})
	sc.Given(`^task (\d+) was skipped$`, func(n int) error {
		_, err := w.other.Ask(engine.TaskSkip, engine.Numbered(n))
		return err
	})
	sc.Given(`^recurrence is (on|off) on the board$`, func(said string) error {
		w.recurs = said == "on"
		return nil
	})
	sc.Then(`^series (\d+) is (active|stopped)$`, func(n int, state string) error {
		s, err := w.series(n)
		if err != nil {
			return err
		}
		if s.State != state {
			return fmt.Errorf("series %d is %s, not %s", n, s.State, state)
		}
		return nil
	})
	sc.Then(`^series (\d+) is titled "([^"]*)"$`, func(n int, title string) error {
		s, err := w.series(n)
		if err != nil {
			return err
		}
		if s.Title != title {
			return fmt.Errorf("series %d is titled %q, not %q", n, s.Title, title)
		}
		return nil
	})
	sc.Then(`^there is no series (\d+)$`, func(n int) error {
		if _, err := w.series(n); err == nil {
			return fmt.Errorf("series %d is still there", n)
		}
		return nil
	})
	sc.Then(`^there is no task (\d+)$`, func(n int) error {
		if _, err := w.givens().Task(n); err == nil {
			return fmt.Errorf("task %d is still there", n)
		}
		return nil
	})
	sc.Then(`^task (\d+) is archived$`, func(n int) error {
		found, err := w.givens().Task(n)
		if err != nil {
			return err
		}
		if !task.OfTask(found).Archived {
			return fmt.Errorf("task %d is not archived", n)
		}
		return nil
	})
	sc.Then(`^task (\d+) is an occurrence of series (\d+)$`, func(n, series int) error {
		found, err := w.other.Ask(engine.TaskShow, engine.Numbered(n))
		if err != nil {
			return err
		}
		if found.Occurring[n].Series != series {
			return fmt.Errorf("task %d is an occurrence of ↻%d, not of ↻%d", n, found.Occurring[n].Series, series)
		}
		return nil
	})
	sc.Then(`^the open task is task (\d+)$`, func(n int) error {
		if w.b.screen != OnDetail || w.b.task != n {
			return fmt.Errorf("the screen %v is open on task %d, not on task %d", w.b.screen, w.b.task, n)
		}
		return w.mustSay(strconv.Itoa(n))
	})
}

func (w *world) series(n int) (engine.Series, error) {
	shown, err := w.other.Ask(engine.SeriesShow, engine.Numbered(n))
	if err != nil {
		return engine.Series{}, err
	}
	return *shown.SeriesShown, nil
}
