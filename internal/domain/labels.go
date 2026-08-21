package domain

import (
	"errors"
	"slices"
	"strings"
)

// ErrLabelsBothWays is returned when a change says both what the set of labels
// becomes and how it changes. Either would be a fine answer; together they ask
// the command to guess which one wins, and nothing here is guessed.
var ErrLabelsBothWays = errors.New("the labels are either replaced or changed, not both")

// The signs that turn a label into a change rather than a value.
const (
	joins = '+'
	parts = '-'
)

// listed splits what was given on commas, so `--label "spec,team"` names two
// labels and not one. It runs before anything else looks at a label, which is
// what makes each piece count as a --label of its own: a signed piece beside a
// plain one then falls foul of the rule that was already there, rather than
// needing a rule of its own.
//
// An empty place — a doubled comma, a trailing one — is dropped rather than
// refused, the same way an empty --label already is.
func listed(given []string) []string {
	pieces := make([]string, 0, len(given))
	for _, one := range given {
		for _, piece := range strings.Split(one, ",") {
			if strings.TrimSpace(piece) == "" && piece != one {
				continue
			}
			pieces = append(pieces, piece)
		}
	}
	return pieces
}

// LabelChange is what a caller asked to happen to a task's labels: either what
// the set becomes, or what joins and leaves it.
type LabelChange struct {
	becomes []string
	added   []string
	removed []string
	adjusts bool
}

// Adjusts says whether the change works from the labels already there. A
// change that replaces needs nothing of what came before; one that adjusts is
// meaningless without it.
func (c LabelChange) Adjusts() bool { return c.adjusts }

// ParseLabelChange reads what a caller gave. A label given plainly says what
// the set becomes; one carrying a sign says how the set changes. Mixing the two
// is refused rather than resolved.
func ParseLabelChange(given []string) (LabelChange, error) {
	var change LabelChange
	given = listed(given)
	plain := make([]string, 0, len(given))

	for _, label := range given {
		if label == "" {
			continue
		}
		sign, name := split(label)
		if sign == 0 {
			plain = append(plain, name)
			continue
		}
		if strings.TrimSpace(name) == "" {
			return LabelChange{}, ErrLabelRequired
		}
		change.adjusts = true
		if sign == joins {
			change.added = append(change.added, strings.ToLower(strings.TrimSpace(name)))
			continue
		}
		change.removed = append(change.removed, strings.ToLower(strings.TrimSpace(name)))
	}

	if change.adjusts && len(plain) > 0 {
		return LabelChange{}, ErrLabelsBothWays
	}
	if change.adjusts {
		return change, nil
	}

	becomes, err := ParseLabels(plain)
	if err != nil {
		return LabelChange{}, err
	}
	change.becomes = becomes
	return change, nil
}

// split takes the sign off a label, if it carries one.
func split(label string) (rune, string) {
	first := rune(label[0])
	if first == joins || first == parts {
		return first, label[1:]
	}
	return 0, label
}

// Apply lays the change over the labels a task carries and hands back the set
// it ends up with: lower case, no duplicate, alphabetical order.
func (c LabelChange) Apply(current []string) ([]string, error) {
	if !c.adjusts {
		return c.becomes, nil
	}
	kept := make([]string, 0, len(current)+len(c.added))
	for _, label := range current {
		if slices.Contains(c.removed, label) {
			continue
		}
		kept = append(kept, label)
	}
	kept = append(kept, c.added...)
	return ParseLabels(kept)
}
