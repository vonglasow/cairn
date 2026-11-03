package domain

import (
	"slices"
	"strings"
)

// SetChange is what a caller asked to happen to a set — the labels a task carries, the tasks in its way, the objectives it serves: either what the set becomes, or what joins and leaves it.
type SetChange[T comparable] struct {
	becomes  []T
	added    []T
	removed  []T
	replaces bool
	signed   bool
}

// ParseSetChange reads what a caller gave: a value joins the set, a value signed with - leaves it, and an = before the values says what the set becomes instead; nothing at all empties it.
func ParseSetChange[T comparable](given []string, read func(raw, value string) (T, error), bothWays error) (SetChange[T], error) {
	var change SetChange[T]
	var plain []T
	named := false
	for _, piece := range listed(given) {
		if piece == "" {
			continue
		}
		named = true
		raw := strings.TrimSpace(piece)
		value, replacing := strings.CutPrefix(raw, replaces)
		if replacing {
			change.replaces = true
			if strings.TrimSpace(value) == "" {
				continue
			}
		}
		sign, rest := SplitSign(value)
		if sign == 0 && !replacing {
			rest = piece
		}
		member, err := read(raw, rest)
		if err != nil {
			return SetChange[T]{}, err
		}
		switch sign {
		case joins:
			change.signed = true
			change.added = append(change.added, member)
		case parts:
			change.signed = true
			change.removed = append(change.removed, member)
		default:
			plain = append(plain, member)
		}
	}
	if change.replaces && change.signed {
		return SetChange[T]{}, bothWays
	}
	if change.replaces || !named {
		change.replaces = true
		change.becomes = plain
		return change, nil
	}
	change.added = append(change.added, plain...)
	return change, nil
}

// Signed says whether a value was given with a + or a -, which only a set already there can be read against.
func (c SetChange[T]) Signed() bool { return c.signed }

// Apply lays the change over the set as it stands and hands back what it ends up with, in no settled order: what a set's order is, and whether it may carry a member twice, is the set's own to say.
func (c SetChange[T]) Apply(current []T) []T {
	if c.replaces {
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
