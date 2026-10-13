package domain

import (
	"errors"
	"fmt"
	"slices"
	"testing"
)

var errNotANumber = errors.New("a number is expected")

var errBothWays = errors.New("replaced or changed, not both")

func readNumber(raw, value string, _ bool) (int, error) {
	var n int
	if _, err := fmt.Sscan(value, &n); err != nil || n <= 0 {
		return 0, errNotANumber
	}
	return n, nil
}

func TestParseSetChange_isOneGrammarForEverySet(t *testing.T) {
	for _, tc := range []struct {
		what  string
		given []string
		want  []int
	}{
		{"what it becomes", []string{"4", "2"}, []int{4, 2}},
		{"added to", []string{"+5"}, []int{1, 2, 5}},
		{"taken from", []string{"-1"}, []int{2}},
		{"both in one piece", []string{"+5,-2"}, []int{1, 5}},
		{"emptied", []string{""}, nil},
	} {
		change, err := ParseSetChange(tc.given, readNumber, errBothWays)
		if err != nil {
			t.Errorf("%s: %v", tc.what, err)
			continue
		}
		if got := change.Apply([]int{1, 2}); !slices.Equal(got, tc.want) {
			t.Errorf("%s: Apply = %v, want %v", tc.what, got, tc.want)
		}
	}
}

func TestParseSetChange_refusesWhatItCannotRead(t *testing.T) {
	if _, err := ParseSetChange([]string{"+5", "3"}, readNumber, errBothWays); !errors.Is(err, errBothWays) {
		t.Errorf("both ways = %v, want errBothWays", err)
	}
	if _, err := ParseSetChange([]string{"+soon"}, readNumber, errBothWays); !errors.Is(err, errNotANumber) {
		t.Errorf("a piece it cannot read = %v, want the reader's refusal", err)
	}
	change, err := ParseSetChange([]string{"+5"}, readNumber, errBothWays)
	if err != nil || !change.Adjusts() {
		t.Errorf("a signed piece = %v, %v, want a change that adjusts", change, err)
	}
	change, err = ParseSetChange([]string{"5"}, readNumber, errBothWays)
	if err != nil || change.Adjusts() {
		t.Errorf("a plain piece = %v, %v, want a change that replaces", change, err)
	}
}
