package gui

// How a key is written for the person in front of the window. The window was
// written on a Mac and said ⌘ everywhere, so on Windows the help told somebody
// to press a key their keyboard does not have. What arrives was never the
// problem — Typed reads ⌘ and Ctrl alike — only the words were.

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

// Spelled writes a key, or a line naming keys, for this keyboard. The help and
// the bar are written once, the Mac's way, and spelled here: one place for the
// difference rather than two spellings of every line.
func (k Keyboard) Spelled(said string) string {
	if k.mac {
		return said
	}
	return strings.NewReplacer("⌘", "Ctrl+", "⇧", "Shift+").Replace(said)
}

// OnSystem is the window written for another system than the one it runs on:
// what a test asks for, so that what it reads does not depend on the machine.
func (w Window) OnSystem(system string) Window {
	w.keyboard = KeyboardOf(system)
	return w
}

// Said is an offer as the bar writes it: the key beside the word, spelled for
// this keyboard. An offer with no key is written alone rather than with a gap
// in front of it.
func (w Window) Said(offer string) string {
	key := Shortcut(offer)
	if key == "" {
		return offer
	}
	return w.keyboard.Spelled(Pretty(key)) + "  " + offer
}

// thisKeyboard is the keyboard of the system the window runs on.
func thisKeyboard() Keyboard { return KeyboardOf(runtime.GOOS) }
