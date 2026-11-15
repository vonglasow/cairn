// Package label holds a label: its name, its colour, its scope.
package label

import (
	"cairn/internal/engine/change"
	"errors"
	"slices"
)

// ErrLabelsBothWays is returned when a change says both what the set of labels becomes and how it changes.
var ErrLabelsBothWays = errors.New("the labels are either replaced or changed, not both")

// Change is what a caller asked to happen to a task's labels: either what the set becomes, or what joins and leaves it.
type Change struct {
	set change.SetChange[string]
}

// ParseLabelChange reads what a caller gave.
func ParseLabelChange(given []string) (Change, error) {
	set, err := change.ParseSetChange(given, readLabel, ErrLabelsBothWays)
	if err != nil {
		return Change{}, err
	}
	return Change{set: set}, nil
}

// Signed says whether a label was given with a + or a -.
func (c Change) Signed() bool { return c.set.Signed() }

func readLabel(_, value string) (string, error) { return labelNamed(value) }

// Apply lays the change over the labels a task carries and hands back the set it ends up with: no duplicate, alphabetical order.
func (c Change) Apply(current []string) ([]string, error) {
	kept := slices.Clone(c.set.Apply(current))
	slices.Sort(kept)
	return slices.Compact(kept), nil
}
