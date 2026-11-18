package board

import (
	"testing"
)

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

func TestASnapshotSaysWhatAFieldsNumbersAreCalled(t *testing.T) {
	s := Snapshot{
		Cards: []Card{{Number: 4, Title: "Ship it"}},
		Goals: []Goal{{Number: 1, Title: "Ship cairn"}, {Number: 2, Title: "Docs", Parent: 1}},
	}
	for _, tc := range []struct {
		f    Field
		want string
	}{
		{Field{Arg: "objective", Value: "1 7"}, "#1 Ship cairn, #7 "},
		{Field{Arg: "objective", Value: "2"}, "#2 Ship cairn › Docs"},
		{Field{Arg: "title", Value: "x"}, "x"},
	} {
		if got := s.Named(tc.f); got != tc.want {
			t.Errorf("Named(%+v) = %q, want %q", tc.f, got, tc.want)
		}
	}
}

func TestNamed_namesTheParentAndWhatAnObjectiveWaitsOn(t *testing.T) {
	s := Snapshot{Goals: []Goal{{Number: 1, Title: "Home automation"}, {Number: 2, Title: "Firmware"}}}
	for _, c := range []struct{ arg, value, want string }{
		{"parent", "1", "#1 Home automation"},
		{"parent", "none", "none"},
		{"parent", "", ""},
		{"waits-on", "1 2", "#1 Home automation, #2 Firmware"},
	} {
		if got := s.Named(Field{Arg: c.arg, Value: c.value}); got != c.want {
			t.Errorf("%s %q reads %q, want %q", c.arg, c.value, got, c.want)
		}
	}
}
