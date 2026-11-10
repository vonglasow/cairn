package test

import (
	"fmt"

	"github.com/cucumber/godog"

	"cairn/internal/engine"
	"cairn/internal/task"
)

// Series reads back a series, as somebody else would see it.
func (g Givens) Series(number int) (engine.Series, error) {
	shown, err := g.ask(engine.SeriesShow, map[string]string{"number": fmt.Sprint(number)}, nil)
	if err != nil {
		return engine.Series{}, err
	}
	return *shown.SeriesShown, nil
}

func (g Givens) registerSeries(sc *godog.ScenarioContext) {
	sc.Given(`^I added a task titled "([^"]*)" coming back "([^"]*)"$`, func(title, every string) error {
		return g.addTask(title, map[string]string{"every": every})
	})
	sc.Given(`^I added a task titled "([^"]*)" coming back "([^"]*)" from "([^"]*)"$`, func(title, every, from string) error {
		return g.addTask(title, map[string]string{"every": every, "from": from})
	})
	sc.Given(`^I stopped series (\d+)$`, func(n int) error {
		_, err := g.ask(engine.SeriesStop, map[string]string{"number": fmt.Sprint(n)}, nil)
		return err
	})
	sc.Given(`^task (\d+) was skipped$`, func(n int) error {
		_, err := g.ask(engine.TaskSkip, map[string]string{"number": fmt.Sprint(n)}, nil)
		return err
	})
	sc.Then(`^series (\d+) is (active|stopped)$`, func(n int, state string) error {
		s, err := g.Series(n)
		if err != nil {
			return err
		}
		if s.State != state {
			return fmt.Errorf("series %d is %s, not %s", n, s.State, state)
		}
		return nil
	})
	sc.Then(`^series (\d+) is titled "([^"]*)"$`, func(n int, title string) error {
		s, err := g.Series(n)
		if err != nil {
			return err
		}
		if s.Title != title {
			return fmt.Errorf("series %d is titled %q, not %q", n, s.Title, title)
		}
		return nil
	})
	sc.Then(`^there is no series (\d+)$`, func(n int) error {
		if _, err := g.Series(n); err == nil {
			return fmt.Errorf("series %d is still there", n)
		}
		return nil
	})
	sc.Then(`^there is no task (\d+)$`, func(n int) error {
		if _, err := g.Task(n); err == nil {
			return fmt.Errorf("task %d is still there", n)
		}
		return nil
	})
	sc.Then(`^task (\d+) is archived$`, func(n int) error {
		found, err := g.Task(n)
		if err != nil {
			return err
		}
		if !task.OfTask(found).Archived {
			return fmt.Errorf("task %d is not archived", n)
		}
		return nil
	})
	sc.Then(`^task (\d+) is an occurrence of series (\d+)$`, func(n, series int) error {
		found, err := g.ask(engine.TaskShow, map[string]string{"number": fmt.Sprint(n)}, nil)
		if err != nil {
			return err
		}
		if found.Occurring[n].Series != series {
			return fmt.Errorf("task %d is an occurrence of ↻%d, not of ↻%d", n, found.Occurring[n].Series, series)
		}
		return nil
	})
}
