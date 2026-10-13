package engine

import (
	"slices"
	"testing"
)

func TestACardSaysWhatIsInItsWay(t *testing.T) {
	for _, tc := range []struct {
		c    Card
		want string
	}{
		{Card{Blockers: []int{3, 5}, Because: "the client"}, "3 5"},
		{Card{Because: "the client"}, "the client"},
		{Card{}, ""},
	} {
		if got := tc.c.InTheWay(); got != tc.want {
			t.Errorf("InTheWay of %+v = %q, want %q", tc.c, got, tc.want)
		}
	}
}

func TestACardSaysWhatItCost(t *testing.T) {
	for _, tc := range []struct {
		c    Card
		want string
		over bool
	}{
		{Card{}, "", false},
		{Card{Spent: "1h"}, "1h", false},
		{Card{Estimate: "2h"}, "2h", false},
		{Card{Spent: "3h", Estimate: "2h", Overrun: true}, "3h/2h", true},
		{Card{Spent: "1h", Estimate: "2h"}, "1h/2h", false},
	} {
		got, over := tc.c.Cost()
		if got != tc.want || over != tc.over {
			t.Errorf("Cost of %+v = %q, %v, want %q, %v", tc.c, got, over, tc.want, tc.over)
		}
	}
}

func TestACardSaysHowFarItsPartsHaveGot(t *testing.T) {
	if got := (Card{SubDone: 2, SubTotal: 5}).PartsDone(); got != "2/5" {
		t.Errorf("PartsDone = %q, want 2/5", got)
	}
	if got := (Card{}).PartsDone(); got != "" {
		t.Errorf("PartsDone with no parts = %q, want nothing", got)
	}
}

func TestASnapshotSaysWhatACardServes(t *testing.T) {
	s := Snapshot{Goals: []Goal{{Number: 1, Title: "A"}, {Number: 2, Title: "B"}}}
	var got []int
	for _, g := range s.Served(Card{Objectives: []int{2, 9, 1}}) {
		got = append(got, g.Number)
	}
	if !slices.Equal(got, []int{2, 1}) {
		t.Errorf("Served = %v, want [2 1]", got)
	}
}

func TestASnapshotSaysWhatAFieldsNumbersAreCalled(t *testing.T) {
	s := Snapshot{
		Cards: []Card{{Number: 4, Title: "Ship it"}},
		Goals: []Goal{{Number: 1, Title: "Ship cairn"}},
	}
	for _, tc := range []struct {
		f    Field
		want string
	}{
		{Field{Arg: "objective", Value: "1 7"}, "#1 Ship cairn, #7 "},
		{Field{Arg: "parent", Value: "4"}, "#4 Ship it"},
		{Field{Arg: "parent", Value: "9"}, "#9 9"},
		{Field{Arg: "title", Value: "x"}, "x"},
	} {
		if got := s.Named(tc.f); got != tc.want {
			t.Errorf("Named(%+v) = %q, want %q", tc.f, got, tc.want)
		}
	}
}
