package gui

// A piece of text somebody is typing into, and where the caret is in it.
//
// One mechanism for every text in this window — the fields of a form and the
// search line alike. A caret that behaved differently in two places would be
// two things to learn, and the second one is always the one that surprises.
//
// What is drawn is not here: where a caret falls among wrapped lines is a
// question about a box, and it is answered in parts.go with the rest of the
// geometry. This is only the text and the place in it.

import "strings"

// Written is a text and the caret in it. The caret is counted in characters,
// never in bytes: half of a letter is not a place.
type Written struct {
	Value string
	At    int
}

// Ending is a text with the caret after its last character, which is where a
// field opens — somebody who has just arrived is adding, not correcting.
func Ending(value string) Written {
	return Written{Value: value, At: len([]rune(value))}
}

// Types puts what was typed in at the caret and leaves the caret after it.
func (w Written) Types(text string) Written {
	if text == "" {
		return w
	}
	kept := []rune(w.Value)
	at := w.settled(w.At)
	w.Value = string(kept[:at]) + text + string(kept[at:])
	w.At = at + len([]rune(text))
	return w
}

// Back takes out the character before the caret, and takes the caret with it.
func (w Written) Back() Written {
	kept := []rune(w.Value)
	at := w.settled(w.At)
	if at == 0 {
		return w
	}
	w.Value = string(kept[:at-1]) + string(kept[at:])
	w.At = at - 1
	return w
}

// Left walks the caret one character back, and stops at the beginning.
func (w Written) Left() Written { return w.CaretAt(w.At - 1) }

// Right walks it one character on, and stops at the end.
func (w Written) Right() Written { return w.CaretAt(w.At + 1) }

// CaretAt puts the caret where it is asked for, or at the nearest end: a click
// past the last letter means the end, not a mistake.
func (w Written) CaretAt(at int) Written {
	w.At = w.settled(at)
	return w
}

// settled keeps a place inside the text.
func (w Written) settled(at int) int {
	return min(len([]rune(w.Value)), max(0, at))
}

// Reads is the text without the caret, which is what gets saved.
func (w Written) Reads() string { return w.Value }

// Empty says whether anything has been written at all.
func (w Written) Empty() bool { return strings.TrimSpace(w.Value) == "" }
