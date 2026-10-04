package gui

import (
	"strings"
	"unicode"
)

// A piece of text somebody is typing into, and where the caret is in it.
//
// One mechanism for every text in this window — the fields of a form and the
// search line alike. A caret that behaved differently in two places would be
// two things to learn, and the second one is always the one that surprises.
//
// What is drawn is not here: where a caret falls among wrapped lines is a
// question about a box, and it is answered in parts.go with the rest of the
// geometry. This is only the text and the place in it.

// Written is a text and the caret in it. The caret is counted in characters,
// never in bytes: half of a letter is not a place.
type Written struct {
	Value string
	At    int
	// Held is how far the anchor of a held run lies from the caret, signed.
	// Nought is nothing held — which makes the zero value, and every literal
	// written before any of this existed, a text holding nothing.
	Held int
}

// Ending is a text with the caret after its last character, which is where a
// field opens — somebody who has just arrived is adding, not correcting.
func Ending(value string) Written {
	return Written{Value: value, At: len([]rune(value))}
}

// Holding says whether a run of text is held.
func (w Written) Holding() bool {
	from, to := w.Run()
	return from != to
}

// Run is the two ends of what is held, in the order they are read.
func (w Written) Run() (from, to int) {
	a, b := w.settled(w.At), w.settled(w.At+w.Held)
	return min(a, b), max(a, b)
}

// Holds is the text between them.
func (w Written) Holds() string {
	from, to := w.Run()
	return string([]rune(w.Value)[from:to])
}

// Selects moves the caret and leaves the anchor where it was, which is what
// holding a run down to somewhere means.
func (w Written) Selects(at int) Written {
	was := w.settled(w.At + w.Held)
	w.At = w.settled(at)
	w.Held = was - w.At
	return w
}

// All holds the whole of it.
func (w Written) All() Written {
	w.At, w.Held = len([]rune(w.Value)), -len([]rune(w.Value))
	return w
}

// Types puts what was typed in at the caret and leaves the caret after it.
func (w Written) Types(text string) Written {
	if text == "" {
		return w
	}
	// What is held is replaced, which is what every text field does and what
	// makes holding a run worth anything.
	if w.Holding() {
		w = w.cut()
	}
	kept := []rune(w.Value)
	at := w.settled(w.At)
	w.Value = string(kept[:at]) + text + string(kept[at:])
	w.At, w.Held = at+len([]rune(text)), 0
	return w
}

// Back takes out the character before the caret, and takes the caret with it.
func (w Written) Back() Written {
	// A held run goes in one stroke rather than one character at a time.
	if w.Holding() {
		return w.cut()
	}
	kept := []rune(w.Value)
	at := w.settled(w.At)
	if at == 0 {
		return w
	}
	w.Value = string(kept[:at-1]) + string(kept[at:])
	w.At, w.Held = at-1, 0
	return w
}

// cut takes the held run out and leaves the caret where it was.
func (w Written) cut() Written {
	from, to := w.Run()
	kept := []rune(w.Value)
	w.Value = string(kept[:from]) + string(kept[to:])
	w.At, w.Held = from, 0
	return w
}

// Left walks the caret one character back, and stops at the beginning.
func (w Written) Left() Written { return w.CaretAt(w.At - 1) }

// Right walks it one character on, and stops at the end.
func (w Written) Right() Written { return w.CaretAt(w.At + 1) }

// CaretAt puts the caret where it is asked for, or at the nearest end: a click
// past the last letter means the end, not a mistake.
func (w Written) CaretAt(at int) Written {
	w.At, w.Held = w.settled(at), 0
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

// Word holds the word around a character: the run of it between spaces, the
// way a double click holds one everywhere else.
func (w Written) Word(at int) Written { return w.around(at, unicode.IsSpace) }

// Paragraph holds the paragraph around a character: the run of it between line
// breaks, the way a triple click holds one everywhere else.
func (w Written) Paragraph(at int) Written {
	return w.around(at, func(r rune) bool { return r == '\n' })
}

// around holds the run around a character that stops at what ends it.
func (w Written) around(at int, ends func(rune) bool) Written {
	said := []rune(w.Value)
	at = min(max(0, at), len(said))
	from, to := at, at
	for from > 0 && !ends(said[from-1]) {
		from--
	}
	for to < len(said) && !ends(said[to]) {
		to++
	}
	return Written{Value: w.Value, At: to, Held: from - to}
}
