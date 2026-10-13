package engine

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"cairn/internal/domain"
)

func changedNumbers(given []string, current []int, atCreation bool,
	notANumber, bothWays, signAtBirth error,
) ([]int, error) {
	change, err := domain.ParseSetChange(given, func(raw, value string, _ bool) (int, error) {
		number, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil || number <= 0 {
			return 0, fmt.Errorf("%q: %w", raw, notANumber)
		}
		return number, nil
	}, bothWays)
	if err != nil {
		return nil, err
	}
	if change.Adjusts() && atCreation {
		return nil, signAtBirth
	}
	return settled(change.Apply(current)), nil
}

func settled(numbers []int) []int {
	slices.Sort(numbers)
	return slices.Compact(numbers)
}

// What a change to what a task serves can be turned down for.
var (
	ErrObjectiveRequired    = errors.New("an objective number is expected")
	ErrObjectivesBothWays   = errors.New("the objectives are either replaced or changed, not both")
	ErrObjectiveSignAtBirth = errors.New("an objective is added or taken away only when changing a task")
)
