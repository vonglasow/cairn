package engine

import (
	"regexp"
	"testing"
)

// unsaid are the words cairn never uses for an act it has a word for: the
// command's own. The boards are held to the same list in internal/tui and
// internal/gui.
var unsaid = regexp.MustCompile(`(?i)\btak(e|es|en|ing) up\b|\blet(ting)? go\b`)

// What every command says of itself — what it is, what it does, what each of
// its arguments means — uses the commands' words for the acts it names.
func TestEveryCommandSaysItsActsWithTheirWords(t *testing.T) {
	for _, op := range All() {
		said := []string{op.Short, op.Does}
		for _, arg := range op.Takes {
			said = append(said, arg.Of, arg.Means)
		}
		for _, line := range said {
			if word := unsaid.FindString(line); word != "" {
				t.Errorf("%s says %q in %q", op.Line(), word, line)
			}
		}
	}
}
