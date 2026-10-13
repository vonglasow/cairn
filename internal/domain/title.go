// Package domain holds the entities and what makes them valid.
package domain

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

// ErrTitleRequired is returned when a title carries no text at all.
var ErrTitleRequired = errors.New("needs a title")

// ErrTitleNotOneLine is returned when a title carries something that is not text.
var ErrTitleNotOneLine = errors.New("a title is one line, with nothing but text on it")

func breaksALine(r rune) bool {
	return unicode.IsControl(r) || r == '\u2028' || r == '\u2029' || hides(r)
}

func hides(r rune) bool {
	return unicode.Is(unicode.Cf, r)
}

// OneLine says whether a title is text and nothing else.
func OneLine(title string) error {
	if strings.ContainsFunc(title, breaksALine) {
		return fmt.Errorf("%q: %w", title, ErrTitleNotOneLine)
	}
	return nil
}

// ErrNotText is returned when what is written about a thing carries a character a terminal reads as an instruction rather than as text.
var ErrNotText = errors.New("only text, line breaks and tabs can be written here")

func instructs(r rune) bool {
	return unicode.IsControl(r) && r != '\n' && r != '\t' || hides(r)
}

// Shown is text as it is safe to put on a screen: every rune a terminal would obey, or that would turn the text around, replaced by the one that says something was there.
func Shown(kept string) string {
	return strings.Map(func(r rune) rune {
		if instructs(r) {
			return '\uFFFD'
		}
		return r
	}, kept)
}

// Prose is what is written about a thing as it is kept: several lines and tabs allowed, a line ending written the Windows way kept as a line break, and anything a terminal would read as an instruction refused.
func Prose(raw string) (string, error) {
	said := strings.ReplaceAll(raw, "\r\n", "\n")
	if strings.ContainsFunc(said, instructs) {
		return "", fmt.Errorf("%q: %w", raw, ErrNotText)
	}
	return said, nil
}

// Absent is what an attribute a thing does not carry looks like.
const Absent = "—"

// OrNothing is a value as it is shown, or Absent when there is none.
func OrNothing(value string) string {
	if value == "" {
		return Absent
	}
	return value
}
