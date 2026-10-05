package domain

import (
	"slices"
	"strings"
)

// SetChange is what a caller asked to happen to a set — the labels a task
// carries, the tasks in its way, the objectives it serves: either what the set
// becomes, or what joins and leaves it.
//
// One grammar for every set, and deliberately: a plain member says what the
// set becomes, `+` adds to it, `-` takes away, and a comma names several in
// one. Somebody who knows one set knows them all. It was written once for the
// labels and once more for the numbers, and the two copies were already apart
// in what they did with a space after a comma.
type SetChange[T comparable] struct {
	becomes []T
	added   []T
	removed []T
	adjusts bool
}

// ParseSetChange reads what a caller gave. Each piece is read by read, which is
// handed the piece as written, its value with any sign taken off, and whether
// it carried one — and refuses a piece in the words of the set it belongs to.
// Mixing what the set becomes with how it changes is refused with bothWays
// rather than resolved: either would be a fine answer, and nothing is guessed.
func ParseSetChange[T comparable](given []string, read func(raw, value string, signed bool) (T, error), bothWays error) (SetChange[T], error) {
	var change SetChange[T]
	var plain []T
	for _, piece := range listed(given) {
		if piece == "" {
			continue
		}
		// The sign is looked for past any space before it, so "+3, -4" is two
		// changes; a piece with no sign is handed over as it was written, so
		// a member that is only spaces is the set's to refuse.
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

// Adjusts says whether the change works from the set already there. A change
// that replaces needs nothing of what came before; one that adjusts is
// meaningless without it.
func (c SetChange[T]) Adjusts() bool { return c.adjusts }

// Apply lays the change over the set as it stands and hands back what it ends
// up with, in no settled order: what a set's order is, and whether it may
// carry a member twice, is the set's own to say.
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
