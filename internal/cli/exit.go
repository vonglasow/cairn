package cli

import "cairn/internal/engine"

const (
	exitOK      = 0
	exitRefused = 1
	exitUsage   = 2
	exitStore   = 3
)

var codes = map[string]int{
	engine.KindRefused: exitRefused,
	engine.KindFailed:  exitRefused,
	engine.KindUsage:   exitUsage,
	engine.KindStore:   exitStore,
}

func codeFor(err error) int {
	if err == nil {
		return exitOK
	}
	return codes[engine.RefusalOf(err).Kind]
}
