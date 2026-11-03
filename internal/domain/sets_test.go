package domain

import (
	"errors"
	"fmt"
	"slices"
	"testing"
)

var errNotANumber = errors.New("a number is expected")

var errBothWays = errors.New("replaced or changed, not both")

func readNumber(raw, value string) (int, error) {
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
		{"what it becomes", []string{"=4", "2"}, []int{2, 4}},
		{"what it becomes in one piece", []string{"=4,2"}, []int{2, 4}},
		{"what it becomes after an equals sign alone", []string{"=", "4", "2"}, []int{2, 4}},
		{"joined with no sign", []string{"5"}, []int{1, 2, 5}},
		{"joined with no sign and with a plus", []string{"5,+6"}, []int{1, 2, 5, 6}},
		{"added to", []string{"+5"}, []int{1, 2, 5}},
		{"taken from", []string{"-1"}, []int{2}},
		{"both in one piece", []string{"+5,-2"}, []int{1, 5}},
		{"emptied", []string{""}, nil},
		{"emptied by nothing at all", nil, nil},
		{"emptied by an equals sign alone", []string{"="}, nil},
	} {
		change, err := ParseSetChange(tc.given, readNumber, errBothWays)
		if err != nil {
			t.Errorf("%s: %v", tc.what, err)
			continue
		}
		got := change.Apply([]int{1, 2})
		slices.Sort(got)
		if !slices.Equal(got, tc.want) {
			t.Errorf("%s: Apply = %v, want %v", tc.what, got, tc.want)
		}
	}
}

func TestParseSetChange_refusesWhatItCannotRead(t *testing.T) {
	for _, given := range [][]string{{"+5", "=3"}, {"=3", "-5"}, {"=3,+5"}} {
		if _, err := ParseSetChange(given, readNumber, errBothWays); !errors.Is(err, errBothWays) {
			t.Errorf("both ways %v = %v, want errBothWays", given, err)
		}
	}
	if _, err := ParseSetChange([]string{"=soon"}, readNumber, errBothWays); !errors.Is(err, errNotANumber) {
		t.Errorf("a piece after an equals sign it cannot read = %v, want the reader's refusal", err)
	}
	if _, err := ParseSetChange([]string{"+soon"}, readNumber, errBothWays); !errors.Is(err, errNotANumber) {
		t.Errorf("a piece it cannot read = %v, want the reader's refusal", err)
	}
}

func TestParseSetChange_saysWhetherItWasSigned(t *testing.T) {
	for given, want := range map[string]bool{"+5": true, "-5": true, "5": false, "=5": false, "5,-6": true} {
		change, err := ParseSetChange([]string{given}, readNumber, errBothWays)
		if err != nil || change.Signed() != want {
			t.Errorf("%q: Signed = %v, %v; want %v", given, change.Signed(), err, want)
		}
	}
}
