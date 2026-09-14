package tui

import "fmt"

// unknown says a command was not understood, naming what was typed and what is
// accepted instead. A prompt that swallowed a typo would leave somebody
// convinced they had run something.
func unknown(line string) error {
	return fmt.Errorf("%q: not a command — try backup, migration, trace or q", line)
}
