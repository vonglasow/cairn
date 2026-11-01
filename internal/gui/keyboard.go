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

// Said is an offer as a button writes it: the word, then the key beside it, spelled for this keyboard.
func (w Window) Said(offer string) string {
	if key := w.OfferKey(offer); key != "" {
		return offer + "  " + key
	}
	return offer
}

// OfferKey is the key an offer answers to, spelled for this keyboard, and nothing when it has none.
func (w Window) OfferKey(offer string) string { return w.spelled(w.shortcut(offer)) }

// BarKey is the key the board's bar says an offer answers to, whatever panel stands over the board.
func (w Window) BarKey(offer string) string { return w.spelled(Shortcut(offer)) }

func (w Window) spelled(key string) string {
	if key == "" {
		return ""
	}
	return w.keyboard.Spelled(Pretty(key))
}

func thisKeyboard() Keyboard { return KeyboardOf(runtime.GOOS) }
