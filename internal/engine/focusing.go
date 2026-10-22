package engine

import (
	"fmt"
	"time"

	"cairn/internal/config"
	"cairn/internal/trace"
)

const throughFocus = "focus"

// NowVariable is the variable that names the moment cairn takes for now when it checks a focus and marks one used, for a script or a scenario to stand later.
const NowVariable = "CAIRN_NOW"

func (e *Engine) clock() time.Time {
	if named, ok := e.home.lookup()(NowVariable); ok {
		if moment, err := time.Parse(time.RFC3339, named); err == nil {
			return moment
		}
	}
	return time.Now()
}

type focusLimits struct {
	idle      time.Duration
	hour      int
	minute    int
	hasDayEnd bool
}

func readFocusLimits(idle, dayEnds string) (focusLimits, error) {
	var limits focusLimits
	if idle != "" {
		held, err := time.ParseDuration(idle)
		if err != nil || held <= 0 {
			return focusLimits{}, fmt.Errorf("focus.idle %q: %w", idle, config.ErrFocus)
		}
		limits.idle = held
	}
	if dayEnds != "" {
		at, err := time.Parse("15:04", dayEnds)
		if err != nil {
			return focusLimits{}, fmt.Errorf("focus.day_ends %q: %w", dayEnds, config.ErrFocus)
		}
		limits.hour, limits.minute = at.Hour(), at.Minute()
		limits.hasDayEnd = true
	}
	return limits, nil
}

func (l focusLimits) passed(seen, now time.Time) bool {
	if l.idle > 0 && now.Sub(seen) > l.idle {
		return true
	}
	if !l.hasDayEnd {
		return false
	}
	end := time.Date(seen.Year(), seen.Month(), seen.Day(), l.hour, l.minute, 0, 0, seen.Location())
	if !end.After(seen) {
		end = end.AddDate(0, 0, 1)
	}
	return end.Before(now)
}

func (e *Engine) focusKept() error {
	if e.home.Dir == "" {
		return nil
	}
	read, err := config.Read(e.home.Dir, e.home.lookup())
	if err != nil {
		return err
	}
	limits, err := readFocusLimits(read.Focus())
	if err != nil {
		return err
	}
	if err := e.focusLetGoPast(limits); err != nil {
		return storeProblem(err)
	}
	return nil
}

func (e *Engine) focusLetGoPast(limits focusLimits) error {
	held, err := e.places.FocusSeen()
	if err != nil {
		return err
	}
	now := e.clock()
	acting := e.acting
	e.acting.Through = throughFocus
	defer func() { e.acting = acting }()
	kept := false
	for _, one := range held {
		if one.Seen.IsZero() || !limits.passed(one.Seen, now) {
			kept = true
			continue
		}
		if err := e.places.DropFocus(one.Task); err != nil {
			return err
		}
		if err := e.recordAt(one.Seen, one.Task, trace.FocusReleased, "", "", ""); err != nil {
			return err
		}
	}
	if !kept {
		return nil
	}
	return e.places.TouchFocus(now)
}
