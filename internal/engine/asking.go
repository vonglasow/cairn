package engine

// The door for a surface that stays open.
//
// Do opens the store, runs one operation and closes it, which is exactly right
// for a surface that lives for one invocation: a command line, an answer to an
// agent. A board on screen cannot pay an open and a close on every keystroke,
// and does not have to — it is holding the engine already, so all that is left
// of Do is the operation itself.

import "cairn/internal/store"

// Ask runs an operation against the store this engine is open on.
//
// It refuses first, for the reason Do refuses first: a caller that named an
// argument the operation does not take asked a different question, and telling
// it so beats answering the question it did not ask.
func (e *Engine) Ask(op Operation, in Input) (Answer, error) {
	if err := op.Check(in); err != nil {
		return Answer{}, err
	}
	return e.run(op, in)
}

// run hands an operation everything it may reach: the store through this
// engine, and what sits beside the store, found from where the store was
// opened. One place builds that, so a surface holding an engine and a surface
// opening one for a single call reach the same things.
//
// An operation that writes runs as one transaction: what it changed and what
// it wrote down about it are kept together or not at all. The ones that change
// the shape of the store are left out — they set a backup aside first, which
// copies the file, and the store keeps each change in a transaction of its
// own (see store.make).
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

// on is what an operation run against this engine is handed.
func (e *Engine) on() On {
	return On{Engine: e, Backups: BackupsAt(e.path), Path: e.path, Notes: e.notes, Home: e.home}
}
