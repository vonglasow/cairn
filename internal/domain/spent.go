package domain

import (
	"fmt"
	"strings"
	"time"
)

// SpentChange is what a caller asked to happen to the time spent on a task: either what the tally becomes, or how far it moves.
type SpentChange struct {
	becomes time.Duration
	by      time.Duration
	moves   bool
}

// ParseSpent reads what a caller gave.
func ParseSpent(given string) (SpentChange, error) {
	if given == "" {
		return SpentChange{}, nil
	}
	sign, rest := SplitSign(given)
	if sign == 0 {
		total, err := ParseEstimate(given)
		if err != nil {
			return SpentChange{}, err
		}
		return SpentChange{becomes: total}, nil
	}

	rest = strings.TrimSpace(rest)
	if strings.HasPrefix(rest, "+") || strings.HasPrefix(rest, "-") {
		return SpentChange{}, fmt.Errorf("%q: %w", given, ErrNotADuration)
	}
	by, err := time.ParseDuration(rest)
	if err != nil {
		return SpentChange{}, fmt.Errorf("%q: %w", given, ErrNotADuration)
	}
	if by%time.Second != 0 {
		return SpentChange{}, fmt.Errorf("%q: %w", given, ErrFinerThanASecond)
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
	return max(current+c.by, 0)
}
