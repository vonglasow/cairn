package gui

import (
	"strings"
	"testing"

	"gioui.org/io/key"
	"gioui.org/io/pointer"

	"cairn/internal/reading"
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

// AltGr is Ctrl and Alt together on Windows, and writes a letter: ś, ź, €.
// It arrives as text, and the key that came with it is not a shortcut. Alone,
// ⌥ with a letter writes one on a Mac too. Neither saves, cuts or archives.
func TestALetterWrittenWithAltIsNoShortcut(t *testing.T) {
	f := newFake()
	w := holding(f, reading.Card{Number: 1, Title: "Write the spec", State: "todo"})
	chosen := w.clicked(middleOf(w, 1, 0))
	if got := Typed(chosen, "X", key.ModCtrl|key.ModAlt); f.what("task", "archive") != 0 || got.Screen() != OnTheBoard {
		t.Errorf("AltGr+X archived, or left the window on %v", got.Screen())
	}
	if got := Typed(chosen, "X", key.ModAlt); f.what("task", "archive") != 0 || got.Screen() != OnTheBoard {
		t.Errorf("⌥X archived, or left the window on %v", got.Screen())
	}
	// ⌥ with what moves the caret, or takes back, is still read.
	if got := Typed(chosen, string(key.NameDeleteBackward), key.ModAlt); got.Screen() != OnTheBoard {
		t.Errorf("⌥⌫ on the board left the window on %v", got.Screen())
	}
}

// A right press while the cairn is being built ends it, as any press does:
// the menu it opens would otherwise be drawn under the stones.
func TestARightPressEndsTheCairn(t *testing.T) {
	f := newFake()
	w := holding(f, reading.Card{Number: 1, Title: "Write the spec", State: "todo"}).Born(moment)
	w, _, _ = frames(t, w, pointer.Event{Kind: pointer.Press, Position: at(w, 1, 0), Buttons: pointer.ButtonSecondary})
	if _, on := w.Intro(); on {
		t.Error("a right press left the cairn building")
	}
}
