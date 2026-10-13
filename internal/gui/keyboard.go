package gui

import (
	"runtime"
	"strings"
)

// Keyboard is the keyboard a window's keys are written for.
type Keyboard struct {
	mac bool
}

// KeyboardOf is the keyboard of a system, named the way Go names it.
func KeyboardOf(system string) Keyboard { return Keyboard{mac: system == "darwin"} }

// Spelled writes a key, or a line naming keys, for this keyboard.
func (k Keyboard) Spelled(said string) string {
	if k.mac {
		return said
	}
	return strings.NewReplacer("⌘", "Ctrl+", "⇧", "Shift+", "⌥", "Alt+", "⌫", "Backspace", "⌦", "Delete").Replace(said)
}

// Said is an offer as the bar writes it: the key beside the word, spelled for this keyboard.
func (w Window) Said(offer string) string {
	key := Shortcut(offer)
	if key == "" {
		return offer
	}
	return w.keyboard.Spelled(Pretty(key)) + "  " + offer
}

func thisKeyboard() Keyboard { return KeyboardOf(runtime.GOOS) }
