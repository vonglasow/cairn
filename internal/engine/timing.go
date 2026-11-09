package engine

import (
	"cairn/internal/config"
	"cairn/internal/store"
	"cairn/internal/trace"
)

func (e *Engine) timeCounted() bool {
	if e.home.Dir == "" {
		return true
	}
	read, err := e.configured()
	return err != nil || read.TimeMode() != config.TimeOff
}

func (e *Engine) timerKept() error {
	counted := e.timeCounted()
	held, err := e.places.Focus()
	if err != nil {
		return err
	}
	var owed []int
	for _, number := range held {
		acts, err := e.history.ActsOn(number)
		if err != nil {
			return err
		}
		if runs, known := trace.TimerRuns(acts, number); known && runs != counted {
			owed = append(owed, number)
		}
	}
	if len(owed) == 0 {
		return nil
	}
	did := trace.TimerStopped
	if counted {
		did = trace.TimerStarted
	}
	return e.store.Within(func(kept *store.Store) error {
		in := e.keptBy(kept)
		for _, number := range owed {
			if err := in.record(number, did, "", "", ""); err != nil {
				return err
			}
		}
		return nil
	})
}
