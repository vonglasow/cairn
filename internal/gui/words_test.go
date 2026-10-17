package gui

import (
	"cairn/internal/engine"
	"regexp"
	"testing"
)

var unsaid = regexp.MustCompile(`(?i)\badd\b|\btak(e|es|en|ing) up\b|\blet(ting)? go\b|off the board|onto the board|\bback(ed)? (the store )?up\b|\bputs? (one |it )?back\b|\bhistor(y|ies)\b`)

func TestTheWindowSaysEveryActWithTheEnginesWord(t *testing.T) {
	w := holding(newFake(), engine.Card{Number: 1, Title: "Write the spec", State: "todo"})
	var said []string
	for _, line := range w.Choose("Help").Helping() {
		said = append(said, line.Key+" "+line.Means)
	}
	for _, group := range MenuBar(engine.Pages{}) {
		said = append(said, group.Name)
		said = append(said, group.Items...)
	}
	said = append(said, AddingAPart)
	for _, line := range said {
		if word := unsaid.FindString(line); word != "" {
			t.Errorf("the window says %q in %q", word, line)
		}
	}
}
