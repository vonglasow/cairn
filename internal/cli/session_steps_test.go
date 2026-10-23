package cli

import (
	"encoding/json"
	"fmt"

	"github.com/cucumber/godog"
)

type focusSession struct {
	Task int    `json:"task"`
	Up   string `json:"up"`
	Ask  string `json:"ask"`
}

func (w *world) sessionOf(n int) (focusSession, error) {
	var answer struct {
		Sessions []focusSession `json:"sessions"`
	}
	if err := json.Unmarshal([]byte(w.Out.Stdout), &answer); err != nil {
		return focusSession{}, fmt.Errorf("the focus answer: %w: %s", err, w.Out.Stdout)
	}
	for _, one := range answer.Sessions {
		if one.Task == n {
			return one, nil
		}
	}
	return focusSession{}, fmt.Errorf("no focus session of task %d in %s", n, w.Out.Stdout)
}

func (w *world) registerSession(sc *godog.ScenarioContext) {
	sc.Step(`^the focus session of task (\d+) has been up "([^"]*)"$`, func(n int, up string) error {
		one, err := w.sessionOf(n)
		if err != nil {
			return err
		}
		if one.Up != up {
			return fmt.Errorf("the focus session of task %d has been up %q, want %q", n, one.Up, up)
		}
		return nil
	})
	sc.Step(`^the focus session of task (\d+) asks "([^"]*)"$`, func(n int, ask string) error {
		one, err := w.sessionOf(n)
		if err != nil {
			return err
		}
		if one.Ask != ask {
			return fmt.Errorf("the focus session of task %d asks %q, want %q", n, one.Ask, ask)
		}
		return nil
	})
	sc.Step(`^the focus session of task (\d+) asks nothing$`, func(n int) error {
		one, err := w.sessionOf(n)
		if err != nil {
			return err
		}
		if one.Ask != "" {
			return fmt.Errorf("the focus session of task %d asks %q, want nothing", n, one.Ask)
		}
		return nil
	})
}
