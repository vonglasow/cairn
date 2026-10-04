package engine

// The door for a surface that stays open.
//
// Do opens the store, runs one operation and closes it, which is exactly right
// for a surface that lives for one invocation: a command line, an answer to an
// agent. A board on screen cannot pay an open and a close on every keystroke,
// and does not have to — it is holding the engine already, so all that is left
// of Do is the operation itself.

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
func (e *Engine) run(op Operation, in Input) (Answer, error) {
	return op.Run(On{Engine: e, Backups: BackupsAt(e.path), Path: e.path, Notes: e.notes}, in)
}
