package engine

import (
	"regexp"
	"testing"
)

var unsaid = regexp.MustCompile(`(?i)\btak(e|es|en|ing) up\b|\blet(ting)? go\b|off the board|onto the board|\bback(ed)? (the store )?up\b|\bputs? (one |it )?back\b|\bhistor(y|ies)\b`)

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
