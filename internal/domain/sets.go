package domain

import (
	"slices"
	"strings"
)

// SetChange is what a caller asked to happen to a set — the labels a task carries, the tasks in its way, the objectives it serves: either what the set becomes, or what joins and leaves it.
type SetChange[T comparable] struct {
	becomes []T
	added   []T
	removed []T
	adjusts bool
}

// ParseSetChange reads what a caller gave.
func ParseSetChange[T comparable](given []string, read func(raw, value string, signed bool) (T, error), bothWays error) (SetChange[T], error) {
	var change SetChange[T]
	var plain []T
	for _, piece := range listed(given) {
		if piece == "" {
			continue
		}
		sign, value := SplitSign(strings.TrimSpace(piece))
		if sign == 0 {
			value = piece
		}
		member, err := read(strings.TrimSpace(piece), value, sign != 0)
		if err != nil {
			return SetChange[T]{}, err
		}
		switch sign {
		case joins:
			change.adjusts = true
			change.added = append(change.added, member)
		case parts:
			change.adjusts = true
			change.removed = append(change.removed, member)
		default:
			plain = append(plain, member)
		}
	}
	if change.adjusts && len(plain) > 0 {
		return SetChange[T]{}, bothWays
	}
	change.becomes = plain
	return change, nil
}

// Adjusts says whether the change works from the set already there.
func (c SetChange[T]) Adjusts() bool { return c.adjusts }

// Apply lays the change over the set as it stands and hands back what it ends up with, in no settled order: what a set's order is, and whether it may carry a member twice, is the set's own to say.
func (c SetChange[T]) Apply(current []T) []T {
	if !c.adjusts {
		return c.becomes
	}
	kept := make([]T, 0, len(current)+len(c.added))
	for _, one := range current {
		if !slices.Contains(c.removed, one) {
			kept = append(kept, one)
		}
	}
	return append(kept, c.added...)
}
