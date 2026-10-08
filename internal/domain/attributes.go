package domain

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
	"unicode"
)

// dayLayout is how a deadline is written and read back: a day, never an
// instant. A deadline with a time of day would be a different promise.
const dayLayout = "2006-01-02"

// The reasons an attribute can be refused. Sentinels, so a surface recognises
// them with errors.Is rather than by matching a message.
var (
	ErrLabelRequired    = errors.New("a label needs a name")
	ErrLabelNotOneLine  = errors.New("a label is one word or a few, on one line")
	ErrLabelHasNoFamily = errors.New("a label has no family: write two labels, one each side of the colon")
	ErrLabelHolds       = errors.New("a label holds only a to z, 0 to 9, dashes and dots")
	ErrNotAState        = errors.New("not one of the columns")
	ErrNotAPriority     = errors.New("not one of the priorities")
	ErrNotADay          = errors.New("not a date")
	ErrNotADuration     = errors.New("not a duration")
	ErrNegativeAmount   = errors.New("a duration cannot run backwards")
	ErrFinerThanASecond = errors.New("a duration is kept to the second, not finer")
)

// State is the column a task sits in. There are seven, and nothing else is one.
type State string

// States holds the columns in the order they are read, which is the order they
// are declared in below. It is what ParseState consults, so the two cannot say
// different things.
//
// The order is a reading order and it is the whole reason this is a slice and
// not a set: what you can act on comes first, then what is under way without
// you, then what you stopped on purpose, then what is history. See
// docs/ordering.md.
var States = []State{Backlog, Todo, Current, Blocked, Watching, Paused, Done}

// The columns. A task with no column stated starts in the backlog.
//
// Blocked, Watching and Paused all mean "not moving", and they are three words
// rather than one because what unblocks them differs: a blocked task waits on
// somebody else and needs chasing, a watched one was handed over and needs
// checking, a paused one waits on nothing but your own decision to pick it up.
// One word for the three would have made the list shorter and the answer to
// "what do I do about this?" unknowable.
const (
	Backlog  State = "backlog"
	Todo     State = "todo"
	Current  State = "current"
	Blocked  State = "blocked"
	Watching State = "watching"
	Paused   State = "paused"
	Done     State = "done"
)

// Priority is what a task carries. There are four, and nothing else is one.
type Priority string

// Priorities holds the four priorities, in the order they are declared above.
// It is what ParsePriority consults, so the two cannot say different things.
var Priorities = []Priority{Low, Normal, High, Critical}

// The four priorities. A task with no priority stated carries normal.
const (
	Low      Priority = "low"
	Normal   Priority = "normal"
	High     Priority = "high"
	Critical Priority = "critical"
)

// naming lists a closed set in a refusal. A refusal that says only what was
// wrong leaves the next attempt to guesswork, and all of these sets are short
// enough to put in the sentence that refuses.
func naming[T ~string](of []T) string {
	words := make([]string, len(of))
	for i, one := range of {
		words[i] = string(one)
	}
	return strings.Join(words, ", ")
}

// ParseState reads a column name. Nothing given returns the backlog, which is
// where a task starts and where clearing the column puts it back.
func ParseState(name string) (State, error) {
	if name == "" {
		return Backlog, nil
	}
	if slices.Contains(States, State(name)) {
		return State(name), nil
	}
	return "", fmt.Errorf("%q: %w — %s", name, ErrNotAState, naming(States))
}

// ParsePriority reads a priority name. Nothing given returns normal, which is
// what a task carries by default and what clearing it returns to.
func ParsePriority(name string) (Priority, error) {
	if name == "" {
		return Normal, nil
	}
	if slices.Contains(Priorities, Priority(name)) {
		return Priority(name), nil
	}
	return "", fmt.Errorf("%q: %w — %s", name, ErrNotAPriority, naming(Priorities))
}

// ParseLabels turns what was given into the set a task carries: lower case, no
// duplicate, alphabetical order. An empty string is how a caller says "none";
// a string of spaces is a label somebody meant to name and did not.
func ParseLabels(given []string) ([]string, error) {
	given = listed(given)
	kept := make([]string, 0, len(given))
	for _, label := range given {
		if label == "" {
			continue
		}
		name, err := labelNamed(label)
		if err != nil {
			return nil, err
		}
		kept = append(kept, name)
	}
	slices.Sort(kept)
	return slices.Compact(kept), nil
}

// labelNamed is one label as it is kept, and the one way it is written, so
// that one word is one label however it was typed (docs/labels.md): lower
// case, its accents taken off, spaces, underscores and slashes made dashes, a
// run of dashes made one and none kept at either end. What is left holds
// a to z, 0 to 9, - and . — anything else is refused, a colon with a word of
// its own. One rule for a label named in a set and one joining or leaving it.
func labelNamed(raw string) (string, error) {
	name := strings.ToLower(strings.TrimSpace(raw))
	if strings.ContainsFunc(name, breaksALine) {
		return "", fmt.Errorf("%q: %w", raw, ErrLabelNotOneLine)
	}
	name = dashed(unaccented(name))
	if name == "" {
		return "", ErrLabelRequired
	}
	if strings.ContainsRune(name, ':') {
		return "", fmt.Errorf("%q: %w", raw, ErrLabelHasNoFamily)
	}
	if strings.ContainsFunc(name, notInALabel) {
		return "", fmt.Errorf("%q: %w", raw, ErrLabelHolds)
	}
	return name, nil
}

// unaccented takes the accents off a word: é is e, and ç is c. The letters
// are named rather than decomposed: the domain leans on the standard library
// alone, and the migration that wrote the labels before the rule knew the
// same letters and no others — a word written one way here and another there
// would be two labels again (docs/labels.md).
var unaccented = strings.NewReplacer(
	"à", "a", "á", "a", "â", "a", "ã", "a", "ä", "a", "å", "a",
	"ç", "c",
	"è", "e", "é", "e", "ê", "e", "ë", "e",
	"ì", "i", "í", "i", "î", "i", "ï", "i",
	"ñ", "n",
	"ò", "o", "ó", "o", "ô", "o", "õ", "o", "ö", "o",
	"ù", "u", "ú", "u", "û", "u", "ü", "u",
	"ý", "y", "ÿ", "y",
).Replace

// dashed makes every space, underscore and slash a dash, a run of dashes one,
// and keeps none at either end.
func dashed(word string) string {
	pieces := strings.FieldsFunc(word, func(r rune) bool {
		return r == '-' || r == '_' || r == '/' || unicode.IsSpace(r)
	})
	return strings.Join(pieces, "-")
}

// notInALabel says whether a character has no place in a label once written.
func notInALabel(r rune) bool {
	return (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '-' && r != '.'
}

// ParseDeadline reads a day. Nothing given is no deadline, which is also how a
// deadline is taken away.
func ParseDeadline(day string) (time.Time, error) {
	if day == "" {
		return time.Time{}, nil
	}
	parsed, err := time.Parse(dayLayout, day)
	if err != nil {
		return time.Time{}, fmt.Errorf("%q: %w — a day written %s", day, ErrNotADay, dayLayout)
	}
	return parsed, nil
}

// ParseEstimate reads a duration. Nothing given is no estimate; a negative one
// is refused rather than stored, since no work takes less than no time.
func ParseEstimate(amount string) (time.Duration, error) {
	if amount == "" {
		return 0, nil
	}
	parsed, err := time.ParseDuration(amount)
	if err != nil {
		return 0, fmt.Errorf("%q: %w — an amount and a unit, as in 2h or 1h30m", amount, ErrNotADuration)
	}
	if parsed < 0 {
		return 0, fmt.Errorf("%q: %w", amount, ErrNegativeAmount)
	}
	// A duration is written back to the second and no finer, so anything with
	// a remainder cannot come back as it was given. Taking it and dropping the
	// remainder would be worse than refusing: 500ms was accepted, stored, and
	// read back as no estimate at all.
	if parsed%time.Second != 0 {
		return 0, fmt.Errorf("%q: %w", amount, ErrFinerThanASecond)
	}
	return parsed, nil
}

// FormatDuration writes a duration the way it was meant to be read: 90 minutes
// comes back as 1h30m, not as 1h30m0s. A duration of nothing writes nothing.
func FormatDuration(d time.Duration) string {
	if d == 0 {
		return ""
	}
	var out strings.Builder
	if h := int(d.Hours()); h > 0 {
		fmt.Fprintf(&out, "%dh", h)
		d -= time.Duration(h) * time.Hour
	}
	if m := int(d.Minutes()); m > 0 {
		fmt.Fprintf(&out, "%dm", m)
		d -= time.Duration(m) * time.Minute
	}
	if s := int(d.Seconds()); s > 0 {
		fmt.Fprintf(&out, "%ds", s)
	}
	return out.String()
}

// FormatDay writes a deadline back as the day it was given. No deadline writes
// nothing.
func FormatDay(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(dayLayout)
}
