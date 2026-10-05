package engine

// A set of task or objective numbers, changed the way a set of labels is. Two
// attributes are given this way — what blocks a task, and what it serves — so
// the reading is written once and the words each refuses with are handed in.

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"cairn/internal/domain"
)

// changedNumbers reads what a caller said about a set of numbers and hands back
// the set it ends up with.
//
// The same grammar as the labels, and deliberately: a plain number says what
// the set becomes, `+3` adds to it, `-3` takes away. Somebody who knows one
// knows all three, and domain.ParseSetChange is the one place the grammar is
// read.
//
// The three refusals are handed in rather than decided here, because what a
// caller reads has to name the thing they were naming: "a blocker is a task
// number" and "an objective number is expected" are the same rule said to two
// different questions.
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

// settled puts a set of numbers in the one order there is, with nothing twice:
// a task told the same number twice was told it once.
func settled(numbers []int) []int {
	slices.Sort(numbers)
	return slices.Compact(numbers)
}

// What a change to what a task serves can be turned down for. The words are
// the ones a caller reads, so they name objectives rather than numbers.
var (
	ErrObjectiveRequired    = errors.New("an objective number is expected")
	ErrObjectivesBothWays   = errors.New("the objectives are either replaced or changed, not both")
	ErrObjectiveSignAtBirth = errors.New("an objective is added or taken away only when changing a task")
)
