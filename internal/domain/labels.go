package domain

import (
	"errors"
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
// the set becomes, or what joins and leaves it. The grammar is SetChange's;
// what is a label, and the order the set is kept in, are the labels' own.
type LabelChange struct {
	set SetChange[string]
}

// Adjusts says whether the change works from the labels already there.
func (c LabelChange) Adjusts() bool { return c.set.Adjusts() }

// ParseLabelChange reads what a caller gave. A label given plainly says what
// the set becomes; one carrying a sign says how the set changes. Mixing the two
// is refused rather than resolved.
func ParseLabelChange(given []string) (LabelChange, error) {
	set, err := ParseSetChange(given, readLabel, ErrLabelsBothWays)
	if err != nil {
		return LabelChange{}, err
	}
	if !set.Adjusts() {
		if set.becomes, err = ParseLabels(set.becomes); err != nil {
			return LabelChange{}, err
		}
	}
	return LabelChange{set: set}, nil
}

// readLabel reads one label of a change. A label joining or leaving the set is
// made lower case here, where it is compared; one saying what the set becomes
// is left for ParseLabels, which holds every label to the same rule.
func readLabel(_, value string, signed bool) (string, error) {
	if !signed {
		return value, nil
	}
	return labelNamed(value)
}

// SplitSign takes the leading + or - off a value, if it carries one, and hands
// back the sign and the rest. Zero for the sign means the value carried none.
//
// Exported because more than one kind of set is changed this way — labels are
// named by word and blockers by number, and both are given, added to or taken
// from with the same two signs. The grammar is written here once so the two
// cannot come to disagree about what "+" means.
func SplitSign(value string) (rune, string) {
	if value == "" {
		return 0, value
	}
	first := rune(value[0])
	if first == joins || first == parts {
		return first, value[1:]
	}
	return 0, value
}

// Apply lays the change over the labels a task carries and hands back the set
// it ends up with: lower case, no duplicate, alphabetical order.
func (c LabelChange) Apply(current []string) ([]string, error) {
	if !c.set.Adjusts() {
		return c.set.becomes, nil
	}
	return ParseLabels(c.set.Apply(current))
}
