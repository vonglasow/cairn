package domain

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
	"unicode"
)

const dayLayout = "2006-01-02"

// The reasons an attribute can be refused.
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

// State is the column a task sits in.
type State string

// States holds the columns in the order they are read, which is the order they are declared in below.
var States = []State{Backlog, Todo, Current, Blocked, Watching, Paused, Done}

// The columns.
const (
	Backlog  State = "backlog"
	Todo     State = "todo"
	Current  State = "current"
	Blocked  State = "blocked"
	Watching State = "watching"
	Paused   State = "paused"
	Done     State = "done"
)

// Priority is what a task carries.
type Priority string

// Priorities holds the four priorities, in the order they are declared above.
var Priorities = []Priority{Low, Normal, High, Critical}

// The four priorities.
const (
	Low      Priority = "low"
	Normal   Priority = "normal"
	High     Priority = "high"
	Critical Priority = "critical"
)

func naming[T ~string](of []T) string {
	words := make([]string, len(of))
	for i, one := range of {
		words[i] = string(one)
	}
	return strings.Join(words, ", ")
}

// ParseState reads a column name.
func ParseState(name string) (State, error) {
	if name == "" {
		return Backlog, nil
	}
	if slices.Contains(States, State(name)) {
		return State(name), nil
	}
	return "", fmt.Errorf("%q: %w — %s", name, ErrNotAState, naming(States))
}

// ParsePriority reads a priority name.
func ParsePriority(name string) (Priority, error) {
	if name == "" {
		return Normal, nil
	}
	if slices.Contains(Priorities, Priority(name)) {
		return Priority(name), nil
	}
	return "", fmt.Errorf("%q: %w — %s", name, ErrNotAPriority, naming(Priorities))
}

// ParseLabels turns what was given into the set a task carries: lower case, no duplicate, alphabetical order.
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

func dashed(word string) string {
	pieces := strings.FieldsFunc(word, func(r rune) bool {
		return r == '-' || r == '_' || r == '/' || unicode.IsSpace(r)
	})
	return strings.Join(pieces, "-")
}

func notInALabel(r rune) bool {
	return (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '-' && r != '.'
}

// ParseDeadline reads a day.
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

// ParseEstimate reads a duration.
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
	if parsed%time.Second != 0 {
		return 0, fmt.Errorf("%q: %w", amount, ErrFinerThanASecond)
	}
	return parsed, nil
}

// FormatDuration writes a duration the way it was meant to be read: 90 minutes comes back as 1h30m, not as 1h30m0s.
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

// FormatDay writes a deadline back as the day it was given.
func FormatDay(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(dayLayout)
}
