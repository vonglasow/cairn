package store

import (
	"testing"
	"time"
)

func TestFocus_saysWhenEachWasLastUsedAndIsTouched(t *testing.T) {
	s := openStore(t)
	first := factsOf(keep(t, s, "First")).Number
	second := factsOf(keep(t, s, "Second")).Number
	taken := time.Date(2026, 10, 21, 9, 30, 0, 0, time.UTC)
	for _, n := range []int{first, second} {
		if err := s.AddFocus(n, taken); err != nil {
			t.Fatal(err)
		}
	}
	used := taken.Add(3 * time.Hour)
	if err := s.TouchFocus(used); err != nil {
		t.Fatal(err)
	}
	if err := s.DropFocus(first); err != nil {
		t.Fatal(err)
	}
	if err := s.AddFocus(first, taken); err != nil {
		t.Fatal(err)
	}
	held, err := s.FocusSeen()
	if err != nil {
		t.Fatal(err)
	}
	if len(held) != 2 || held[0].Task != second || !held[0].Seen.Equal(used) || held[1].Task != first || !held[1].Seen.Equal(taken) {
		t.Errorf("what is in focus, and when it was last used = %+v", held)
	}
}
