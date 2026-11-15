package change

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// ChangedNumbers reads a change to a set of numbers and applies it to the current ones.
func ChangedNumbers(given []string, current []int, atCreation bool,
	notANumber, bothWays, signAtBirth error,
) ([]int, error) {
	c, err := ParseSetChange(given, func(raw, value string) (int, error) {
		number, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil || number <= 0 {
			return 0, fmt.Errorf("%q: %w", raw, notANumber)
		}
		return number, nil
	}, bothWays)
	if err != nil {
		return nil, err
	}
	if c.Signed() && atCreation {
		return nil, signAtBirth
	}
	return settled(c.Apply(current)), nil
}

func settled(numbers []int) []int {
	slices.Sort(numbers)
	return slices.Compact(numbers)
}
