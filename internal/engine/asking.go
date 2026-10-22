package engine

import "cairn/internal/store"

// Ask runs an operation against the store this engine is open on.
func (e *Engine) Ask(op Operation, in Input) (Answer, error) {
	if err := op.admitted(in); err != nil {
		return Answer{}, err
	}
	defer e.taking()()
	if err := e.following(); err != nil {
		return Answer{}, Told(err)
	}
	if op.Level != Read && op.Needs == AStore {
		if err := e.focusKept(); err != nil {
			return Answer{}, err
		}
	}
	return e.run(op, in)
}

func (e *Engine) run(op Operation, in Input) (Answer, error) {
	if op.Level == Read || op.Needs != AStore {
		return op.Run(e.on(), in)
	}
	var answer Answer
	err := e.store.Within(func(kept *store.Store) error {
		var err error
		answer, err = op.Run(e.keptBy(kept).on(), in)
		return err
	})
	return answer, err
}

func (e *Engine) on() On {
	return On{Engine: e, Backups: BackupsAt(e.path), Path: e.path, Notes: e.notes, Home: e.home}
}
