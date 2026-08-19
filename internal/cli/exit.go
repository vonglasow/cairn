package cli

import (
	"errors"

	"cairn/internal/domain"
	"cairn/internal/engine"
)

// The codes cairn returns to the shell. They answer one question: what kind of
// answer is this? A script that only knew "it failed" would have to read the
// message to decide whether to fix its arguments or to retry later.
const (
	exitOK      = 0 // it worked
	exitRefused = 1 // the request was understood, and could not be satisfied
	exitUsage   = 2 // the command was not understood at all
	exitStore   = 3 // the store could not be used
)

// refusals are the reasons a request was understood and still could not be
// satisfied. Anything not named here was not understood in the first place —
// a flag that does not exist, a number that is not one — which is a different
// kind of answer and gets a different code.
var refusals = []error{
	domain.ErrTitleRequired,
	domain.ErrLabelRequired,
	domain.ErrLabelsBothWays,
	domain.ErrNotAState,
	domain.ErrNotAPriority,
	domain.ErrNotADay,
	domain.ErrNotADuration,
	domain.ErrNegativeAmount,
	engine.ErrNoSuchTask,
	engine.ErrNothingToChange,
	engine.ErrLabelSignAtCreation,
	errNotConfirmed,
}

// codeFor says which kind of answer an error is.
func codeFor(err error) int {
	if err == nil {
		return exitOK
	}
	if errors.Is(err, engine.ErrStore) {
		return exitStore
	}
	for _, refusal := range refusals {
		if errors.Is(err, refusal) {
			return exitRefused
		}
	}
	return exitUsage
}
