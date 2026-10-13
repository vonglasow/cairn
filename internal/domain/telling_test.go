package domain

import (
	"slices"
	"testing"
	"time"
)

func TestTellOrNot_saysAValueOrThatThereIsNone(t *testing.T) {
	var said []string
	TellOrNot("spec", func(s string) { said = append(said, "told "+s) }, func() { said = append(said, "none") })
	TellOrNot("", func(s string) { said = append(said, "told "+s) }, func() { said = append(said, "none") })
	TellOrNot(time.Duration(0), func(time.Duration) { said = append(said, "told a length") }, func() { said = append(said, "no length") })
	if want := []string{"told spec", "none", "no length"}; !slices.Equal(said, want) {
		t.Errorf("said %v, want %v", said, want)
	}
}

func TestTellAll_tellsEachMemberThenClosesTheSet(t *testing.T) {
	var said []string
	each := func(s string) { said = append(said, s) }
	none := func() { said = append(said, "none") }
	closed := func() { said = append(said, "closed") }
	TellAll([]string{"a", "b"}, each, none, closed)
	TellAll(nil, each, none, closed)
	if want := []string{"a", "b", "closed", "none", "closed"}; !slices.Equal(said, want) {
		t.Errorf("said %v, want %v", said, want)
	}
}
