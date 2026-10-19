package gui

import (
	"strings"
	"unicode"
)

// Written is a text and the caret in it.
type Written struct {
	Value string
	At    int
	Held  int
}

// Ending is a text with the caret after its last character, which is where a field opens — somebody who has just arrived is adding, not correcting.
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

// Selects moves the caret and leaves the anchor where it was, which is what holding a run down to somewhere means.
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
	from, to := w.Run()
	return w.Replaces(from, to, text)
}

// Replaces puts a text in place of the run between two places and leaves the caret after it, which is how a system composing a letter takes back the accent it wrote first.
func (w Written) Replaces(from, to int, text string) Written {
	from, to = w.settled(min(from, to)), w.settled(max(from, to))
	kept := []rune(w.Value)
	w.Value = string(kept[:from]) + text + string(kept[to:])
	w.At, w.Held = from+len([]rune(text)), 0
	return w
}

// Back takes out the character before the caret, and takes the caret with it.
func (w Written) Back() Written {
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

// CaretAt puts the caret where it is asked for, or at the nearest end: a click past the last letter means the end, not a mistake.
func (w Written) CaretAt(at int) Written {
	w.At, w.Held = w.settled(at), 0
	return w
}

func (w Written) settled(at int) int {
	return min(len([]rune(w.Value)), max(0, at))
}

// Reads is the text without the caret, which is what gets saved.
func (w Written) Reads() string { return w.Value }

// Empty says whether anything has been written at all.
func (w Written) Empty() bool { return strings.TrimSpace(w.Value) == "" }

// Word holds the word around a character: the run of it between spaces, the way a double click holds one everywhere else.
func (w Written) Word(at int) Written { return w.around(at, unicode.IsSpace) }

// Paragraph holds the paragraph around a character: the run of it between line breaks, the way a triple click holds one everywhere else.
func (w Written) Paragraph(at int) Written {
	return w.around(at, func(r rune) bool { return r == '\n' })
}

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

// LineStart puts the caret at the start of the line it is on: just after the break before it, or at the very start.
func (w Written) LineStart() Written {
	line := w.Paragraph(w.At)
	return w.CaretAt(line.At + line.Held)
}

// LineEnd puts the caret at the end of the line it is on: just before the break after it, or at the very end.
func (w Written) LineEnd() Written { return w.CaretAt(w.Paragraph(w.At).At) }

// WordBack puts the caret at the start of the word it is in, or of the one before when it is already there.
func (w Written) WordBack() Written {
	said := []rune(w.Value)
	at := w.settled(w.At)
	for at > 0 && unicode.IsSpace(said[at-1]) {
		at--
	}
	for at > 0 && !unicode.IsSpace(said[at-1]) {
		at--
	}
	return w.CaretAt(at)
}

// WordOn puts the caret at the end of the word it is in, or of the one after when it is already there.
func (w Written) WordOn() Written {
	said := []rune(w.Value)
	at := w.settled(w.At)
	for at < len(said) && unicode.IsSpace(said[at]) {
		at++
	}
	for at < len(said) && !unicode.IsSpace(said[at]) {
		at++
	}
	return w.CaretAt(at)
}

// Forward takes out the character after the caret, and leaves the caret where it is.
func (w Written) Forward() Written {
	if w.Holding() {
		return w.cut()
	}
	return w.Selects(w.At + 1).cut()
}

// BackTo takes out everything between a place before the caret and the caret: back to the start of the word, or of the line.
func (w Written) BackTo(at int) Written {
	if w.Holding() {
		return w.cut()
	}
	return w.Selects(at).cut()
}
