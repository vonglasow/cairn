package domain

import (
	"fmt"
	"strings"
	"time"
)

// SpentChange is what a caller asked to happen to the time spent on a task:
// either what the tally becomes, or how far it moves.
//
// The same shape as LabelChange, on purpose. A caller who has learnt that
// `+urgent` joins a set has learnt that `+30m` joins a tally, and neither is a
// rule to look up.
type SpentChange struct {
	becomes time.Duration
	by      time.Duration
	moves   bool
}

// Moves says whether the change works from the time already recorded. A change
// that sets needs nothing of what came before; one that moves is meaningless
// without it.
func (c SpentChange) Moves() bool { return c.moves }

// ParseSpent reads what a caller gave. A duration given plainly says what the
// tally becomes; one carrying a sign says how far it moves. Nothing given
// clears it.
func ParseSpent(given string) (SpentChange, error) {
	if given == "" {
		return SpentChange{}, nil
	}
	sign, rest := split(given)
	if sign == 0 {
		total, err := ParseEstimate(given)
		if err != nil {
			return SpentChange{}, err
		}
		return SpentChange{becomes: total}, nil
	}

	// The sign is already read; a second one inside the duration is somebody
	// saying two things at once. Refusing it here also keeps ParseDuration from
	// ever handing back a negative, so there is no unreachable guard below to
	// carry — and no mutant on it that no test could ever kill.
	rest = strings.TrimSpace(rest)
	if strings.HasPrefix(rest, "+") || strings.HasPrefix(rest, "-") {
		return SpentChange{}, fmt.Errorf("%q: %w", given, ErrNotADuration)
	}
	by, err := time.ParseDuration(rest)
	if err != nil {
		return SpentChange{}, fmt.Errorf("%q: %w", given, ErrNotADuration)
	}
	if sign == parts {
		by = -by
	}
	return SpentChange{by: by, moves: true}, nil
}

// Apply lays the change over the time already recorded.
func (c SpentChange) Apply(current time.Duration) time.Duration {
	if !c.moves {
		return c.becomes
	}
	// A tally of work done has no negative side: taking off more than was ever
	// put on leaves nothing, rather than a number that would travel into every
	// sum that follows.
	return max(current+c.by, 0)
}
