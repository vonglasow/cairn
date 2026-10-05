// Package domain holds the entities and what makes them valid. It knows
// nobody: an entity that knew how it is displayed would drag a terminal into
// every test that touches it, and one that knew how it is stored would drag a
// database in.
package domain

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

// ErrTitleRequired is returned when a title carries no text at all. It is a
// sentinel so a surface can recognise it with errors.Is rather than by
// matching a message that will be reworded one day.
//
// It carries no subject of its own: what needs a title is named by whoever
// wraps it, so "a task needs a title" and "an objective needs a title" are one
// rule with two wordings rather than two sentinels a surface has to know both
// of.
var ErrTitleRequired = errors.New("needs a title")

// ErrTitleNotOneLine is returned when a title carries something that is not
// text. A line break turns one row of a listing into two, the second carrying
// no number and belonging to nothing; a tab walks through the tabwriter that
// lines a card up and shifts columns it has no business in. Trimming it away
// quietly would mean keeping a title other than the one that was typed.
var ErrTitleNotOneLine = errors.New("a title is one line, with nothing but text on it")

// breaksALine says whether a rune would end the line it is written on, or
// otherwise fail to be text. IsControl covers C0 and C1 and stops there, so
// the two Unicode separators are named: U+2028 breaks a line and U+2029 a
// paragraph, and both split a listing row in two exactly like a newline.
func breaksALine(r rune) bool {
	return unicode.IsControl(r) || r == '\u2028' || r == '\u2029' || hides(r)
}

// hides says whether a rune turns the text around it, or sits in it unseen:
// the bidirectional overrides and isolates, the zero-width spaces and joiners,
// and the byte-order mark. None of them is a control character, and every one
// makes text show something other than what it holds — a title drawn
// backwards, a word with an invisible break in it.
func hides(r rune) bool {
	return r >= '\u202A' && r <= '\u202E' || r >= '\u2066' && r <= '\u2069' ||
		r >= '\u200B' && r <= '\u200D' || r == '\uFEFF'
}

// OneLine says whether a title is text and nothing else. It is shared by a
// task and an objective, which are shown the same way and break the same way
// — exported because a task, built outside this package, still has to be
// held to the same rule.
func OneLine(title string) error {
	// ContainsFunc rather than IndexFunc >= 0: the index is not wanted, and a
	// comparison on it is a boundary no test can be written against once the
	// first character is covered.
	if strings.ContainsFunc(title, breaksALine) {
		return fmt.Errorf("%q: %w", title, ErrTitleNotOneLine)
	}
	return nil
}

// ErrNotText is returned when what is written about a thing carries a
// character a terminal reads as an instruction rather than as text. A
// description is printed straight to the screen by more than one surface, and
// an escape among it hides text, draws a link that lies about where it goes,
// or writes to the clipboard — and an agent can be talked into writing one.
var ErrNotText = errors.New("only text, line breaks and tabs can be written here")

// instructs says whether a rune, among prose, would be read as an instruction
// rather than shown: every control character but the line break and the tab.
func instructs(r rune) bool {
	return unicode.IsControl(r) && r != '\n' && r != '\t' || hides(r)
}

// Shown is text as it is safe to put on a screen: every rune a terminal would
// obey, or that would turn the text around, replaced by the one that says
// something was there. What is written is refused such runes; what was kept
// before that refusal existed — or comes back with an older backup — is
// shown this way instead of obeyed.
func Shown(kept string) string {
	return strings.Map(func(r rune) rune {
		if instructs(r) {
			return '\uFFFD'
		}
		return r
	}, kept)
}

// Prose is what is written about a thing as it is kept: several lines and tabs
// allowed, a line ending written the Windows way kept as a line break, and
// anything a terminal would read as an instruction refused. Nothing is trimmed:
// what somebody wrote is what is kept, or it is refused.
func Prose(raw string) (string, error) {
	said := strings.ReplaceAll(raw, "\r\n", "\n")
	if strings.ContainsFunc(said, instructs) {
		return "", fmt.Errorf("%q: %w", raw, ErrNotText)
	}
	return said, nil
}

// Absent is what an attribute a thing does not carry looks like. A blank would
// read as a value nobody typed rather than as an absence.
//
// It sits here beside FormatDay and FormatDuration, and for the same reason:
// what an absence looks like is a fact about a carried attribute, not about the
// thing carrying it. A task and an objective both show one, and neither may
// import the other, so this is the only place both can read it from.
const Absent = "—"

// OrNothing is a value as it is shown, or Absent when there is none.
func OrNothing(value string) string {
	if value == "" {
		return Absent
	}
	return value
}
