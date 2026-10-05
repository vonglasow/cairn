package cli

import "cairn/internal/engine"

// The codes cairn returns to the shell. They answer one question: what kind of
// answer is this? A script that only knew "it failed" would have to read the
// message to decide whether to fix its arguments or to retry later.
const (
	exitOK      = 0 // it worked
	exitRefused = 1 // the request was understood and could not be satisfied, or failed
	exitUsage   = 2 // the command was not understood at all
	exitStore   = 3 // the store could not be used
)

// codes turns what kind of answer a failure is — said by the engine, for every
// surface — into what a shell reads. A refusal and a failure nobody foresaw
// share a code: neither is fixed by rewording the command.
var codes = map[string]int{
	engine.KindRefused: exitRefused,
	engine.KindFailed:  exitRefused,
	engine.KindUsage:   exitUsage,
	engine.KindStore:   exitStore,
}

// codeFor says which kind of answer an error is.
func codeFor(err error) int {
	if err == nil {
		return exitOK
	}
	return codes[engine.RefusalOf(err).Kind]
}
