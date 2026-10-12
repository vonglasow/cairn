package gui

import (
	"regexp"
	"testing"

	"cairn/internal/reading"
)

// unsaid are the words the window never uses for an act: each act has one
// word, the engine's, on every surface — the terminal board's are held to the
// same list in internal/tui.
var unsaid = regexp.MustCompile(`(?i)\badd\b`)

// The window says every act with the engine's word for it, in its help and in
// its menu bar alike.
func TestTheWindowSaysEveryActWithTheEnginesWord(t *testing.T) {
	w := holding(newFake(), reading.Card{Number: 1, Title: "Write the spec", State: "todo"})
	var said []string
	for _, line := range w.Choose("Help").Helping() {
		said = append(said, line.Key+" "+line.Means)
	}
	for _, group := range MenuBar() {
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
