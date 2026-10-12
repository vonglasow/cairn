package gui

import (
	"strings"
	"testing"
)

// A key is written the way the keyboard in front of the person writes it: ⌘ and
// ⇧ on a Mac, Ctrl+ and Shift+ everywhere else. What arrives is the same on
// all of them — Typed reads ⌘ and Ctrl alike — so only the words change.
func TestAKeyIsWrittenTheWayItsSystemWritesIt(t *testing.T) {
	for _, c := range []struct{ system, said, want string }{
		{"darwin", "⌘N", "⌘N"},
		{"darwin", "⇧A", "⇧A"},
		{"windows", "⌘N", "Ctrl+N"},
		{"linux", "⌘⏎ or ⌘S in a form", "Ctrl+⏎ or Ctrl+S in a form"},
		{"windows", "⇧←→ in a field", "Shift+←→ in a field"},
		{"windows", "esc", "esc"},
		{"windows", "⌥←→ a word", "Alt+←→ a word"},
		{"linux", "⌦ ⌘⌫ ⌥⌫", "Delete Ctrl+Backspace Alt+Backspace"},
		{"darwin", "⌦ ⌘⌫ ⌥⌫", "⌦ ⌘⌫ ⌥⌫"},
	} {
		if got := KeyboardOf(c.system).Spelled(c.said); got != c.want {
			t.Errorf("on %s, %q is written %q, want %q", c.system, c.said, got, c.want)
		}
	}
}

// The bar's offers and every line of the help are written for the system the
// window runs on.
func TestTheWindowWritesItsKeysForItsSystem(t *testing.T) {
	w := New(newFake(), 1200, 800).OnSystem("windows")
	for _, line := range w.Helping() {
		if strings.ContainsAny(line.Key+line.Means, "⌘⇧⌥⌫⌦") {
			t.Errorf("the help says %q — %q on Windows", line.Key, line.Means)
		}
	}
	if got := w.Said("New"); got != "Ctrl+N  New" {
		t.Errorf("the bar writes %q, want the key a Windows keyboard has", got)
	}
	if got := New(newFake(), 1200, 800).OnSystem("darwin").Said("New"); got != "⌘N  New" {
		t.Errorf("on a Mac the bar writes %q", got)
	}
}
