package tui

import (
	"fmt"
	"slices"
	"strings"

	"github.com/cucumber/godog"
)

func (w *world) registerTree(sc *godog.ScenarioContext) {
	sc.Then(`^the objectives read (".*")$`, w.objectivesRead)
	sc.Then(`^objective (\d+) sits under objective (\d+)$`, func(n, parent int) error {
		o, err := w.givens().Objective(n)
		if err != nil {
			return err
		}
		if got := o.Carried().Parent; got != parent {
			return fmt.Errorf("objective %d sits under %d, not under %d", n, got, parent)
		}
		return nil
	})
	sc.Then(`^objective (\d+) stands on its own$`, func(n int) error {
		o, err := w.givens().Objective(n)
		if err != nil {
			return err
		}
		if got := o.Carried().Parent; got != 0 {
			return fmt.Errorf("objective %d sits under %d", n, got)
		}
		return nil
	})
	sc.Then(`^objective (\d+) now waits on objective (\d+)$`, func(n, on int) error {
		o, err := w.givens().Objective(n)
		if err != nil {
			return err
		}
		if !slices.Contains(o.Carried().WaitsOn, on) {
			return fmt.Errorf("objective %d waits on %v, not on %d", n, o.Carried().WaitsOn, on)
		}
		return nil
	})
	sc.Then(`^objective (\d+) is gone$`, func(n int) error {
		if _, err := w.givens().Objective(n); err == nil {
			return fmt.Errorf("objective %d is still there", n)
		}
		return nil
	})
	sc.Then(`^objective (\d+) is still there$`, func(n int) error {
		_, err := w.givens().Objective(n)
		return err
	})
	sc.Then(`^task (\d+) serves nothing$`, func(n int) error {
		t, err := w.givens().Task(n)
		if err != nil {
			return err
		}
		if got := t.Carried().Objectives; len(got) > 0 {
			return fmt.Errorf("task %d serves %v", n, got)
		}
		return nil
	})
	sc.Then(`^task (\d+) serves objective (\d+)$`, func(n, objective int) error {
		t, err := w.givens().Task(n)
		if err != nil {
			return err
		}
		if !slices.Contains(t.Carried().Objectives, objective) {
			return fmt.Errorf("task %d serves %v, not %d", n, t.Carried().Objectives, objective)
		}
		return nil
	})
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
