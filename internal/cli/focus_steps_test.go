package cli

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cucumber/godog"

	"cairn/internal/engine"
	"cairn/test"
)

func (w *world) writeFocusLimits() error {
	said := "[focus]\n"
	if w.focusIdle != "" {
		said += fmt.Sprintf("idle = %q\n", w.focusIdle)
	}
	if w.focusDayEnds != "" {
		said += fmt.Sprintf("day_ends = %q\n", w.focusDayEnds)
	}
	return w.WriteConfig(said)
}

func (w *world) registerFocusLetGo(sc *godog.ScenarioContext) {
	sc.Step(`^it is (\d+) (minutes|hours) later$`, func(count int, unit string) error {
		later := time.Duration(count) * time.Minute
		if unit == "hours" {
			later = time.Duration(count) * time.Hour
		}
		w.Setenv(engine.NowVariable, time.Now().Add(later).Format(time.RFC3339))
		return nil
	})
	sc.Step(`^it says "([^"]*)"$`, func(want string) error {
		return test.WantContains(w.Out.Stdout, want, "the answer")
	})
	sc.Step(`^focus is let go after "([^"]*)" unused$`, func(idle string) error {
		w.focusIdle = idle
		return w.writeFocusLimits()
	})
	sc.Step(`^the working day ends at "([^"]*)"$`, func(at string) error {
		w.focusDayEnds = at
		return w.writeFocusLimits()
	})
	sc.Step(`^the focus of task (\d+) was let go through "([^"]*)" when it was taken$`, func(n int, through string) error {
		before := w.Out
		defer func() { w.Out = before }()
		w.Run(fmt.Sprintf("cairn trace show %d --json", n))
		var shown struct {
			Acts []struct {
				Did     string
				Through string
				At      time.Time
			}
		}
		if err := json.Unmarshal([]byte(w.Out.Stdout), &shown); err != nil {
			return fmt.Errorf("the trace of task %d: %w: %s", n, err, w.Out.Stdout)
		}
		var taken, released time.Time
		var by string
		for _, act := range shown.Acts {
			switch act.Did {
			case "focus taken":
				taken = act.At
			case "focus released":
				released, by = act.At, act.Through
			}
		}
		if released.IsZero() || taken.IsZero() {
			return fmt.Errorf("task %d was taken at %v and let go at %v", n, taken, released)
		}
		if by != through {
			return fmt.Errorf("the focus was let go through %q, want %q", by, through)
		}
		if released.Sub(taken).Abs() > time.Second {
			return fmt.Errorf("the focus was taken at %v and let go at %v", taken, released)
		}
		return nil
	})
}
