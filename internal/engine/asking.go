package engine

import (
	"errors"
	"time"

	"cairn/internal/store"
)

// Ask runs an operation against the store this engine is open on.
func (e *Engine) Ask(op Operation, in Input) (answer Answer, err error) {
	defer func(from time.Time) { noteDone(op, in, from, answer, err) }(time.Now())
	if err := op.admitted(in); err != nil {
		return Answer{}, err
	}
	defer e.taking()()
	if err := e.following(); err != nil {
		return Answer{}, Told(err)
	}
	if e.refused != nil && op.Needs != NoStore {
		return Answer{}, Told(e.refused)
	}
	if op.Needs != NoStore {
		if err := e.store.Current(); err != nil && (op.Needs == AStore || errors.Is(err, ErrNewerStore)) {
			return Answer{}, Told(err)
		}
	}
	if op.Needs == AStore {
		if err := e.sweepIfOwed(); err != nil {
			return Answer{}, Told(err)
		}
		if err := e.timerKept(); err != nil {
			return Answer{}, Told(storeProblem(err))
		}
		if err := e.seriesDue(); err != nil {
			return Answer{}, Told(storeProblem(err))
		}
	}
	return e.run(op, in)
}

func (e *Engine) run(op Operation, in Input) (Answer, error) {
	answer, err := e.acted(op, in)
	if err != nil || e.timeShown() {
		return answer, err
	}
	return answer.untimed(), nil
}

func (e *Engine) acted(op Operation, in Input) (Answer, error) {
	if op.Level == Read || op.Needs != AStore {
		return op.Run(e.on(), in)
	}
	var answer Answer
	err := e.store.Within(func(kept *store.Store) error {
		held := e.keptBy(kept)
		before, err := held.openSeries()
		if err != nil {
			return err
		}
		if answer, err = op.Run(held.on(), in); err != nil {
			return err
		}
		made, err := held.madeAfter(before)
		answer.Followed = append(answer.Followed, made...)
		return err
	})
	return answer, err
}

func (e *Engine) on() On {
	return On{Engine: e, Backups: BackupsAt(e.path), Path: e.path, Notes: e.notes, Home: e.home}
}
