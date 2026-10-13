package domain

import (
	"errors"
	"slices"
	"strings"
)

// ErrLabelsBothWays is returned when a change says both what the set of labels becomes and how it changes.
var ErrLabelsBothWays = errors.New("the labels are either replaced or changed, not both")

const (
	joins = '+'
	parts = '-'
)

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

// LabelChange is what a caller asked to happen to a task's labels: either what the set becomes, or what joins and leaves it.
type LabelChange struct {
	set SetChange[string]
}

// Adjusts says whether the change works from the labels already there.
func (c LabelChange) Adjusts() bool { return c.set.Adjusts() }

// ParseLabelChange reads what a caller gave.
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

func readLabel(_, value string, signed bool) (string, error) {
	if !signed {
		return value, nil
	}
	return labelNamed(value)
}

// SplitSign takes the leading + or - off a value, if it carries one, and hands back the sign and the rest.
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

// Apply lays the change over the labels a task carries and hands back the set it ends up with: no duplicate, alphabetical order.
func (c LabelChange) Apply(current []string) ([]string, error) {
	if !c.set.Adjusts() {
		return c.set.becomes, nil
	}
	kept := c.set.Apply(current)
	slices.Sort(kept)
	return slices.Compact(kept), nil
}
