package tui

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

// unsaid are the words the board never uses for an act: each act has one
// word, the engine's, on every surface — the window's are held to the same
// list in internal/gui.
var unsaid = regexp.MustCompile(`(?i)\badd\b|\btak(e|es|en|ing) up\b|\blet(ting)? go\b|off the board|onto the board|\bback(ed)? (the store )?up\b|\bputs? (one |it )?back\b|\bhistor(y|ies)\b`)

// The board says every act with the engine's word for it: every frame drawn
// here, its help and its bar included, is read for the words it must not say.
func TestTheBoardSaysEveryActWithTheEnginesWord(t *testing.T) {
	frames, err := filepath.Glob(filepath.Join("testdata", "*.txt"))
	if err != nil || len(frames) == 0 {
		t.Fatalf("no frame to read: %v", err)
	}
	for _, frame := range frames {
		drawn, err := os.ReadFile(frame)
		if err != nil {
			t.Fatal(err)
		}
		for _, word := range unsaid.FindAllString(string(drawn), -1) {
			t.Errorf("%s says %q", frame, word)
		}
	}
}
