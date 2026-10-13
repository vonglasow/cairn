package trace

import (
	"testing"
	"time"
)

type everyFact struct {
	task                             int
	did, field, was, became, through string
	at                               time.Time
	nothingNamed, wasNothing         bool
	becameNothing, byHand            bool
}

func (f *everyFact) On(n int)               { f.task = n }
func (f *everyFact) Did(what string)        { f.did = what }
func (f *everyFact) Changing(field string)  { f.field = field }
func (f *everyFact) NothingNamed()          { f.nothingNamed = true }
func (f *everyFact) Was(value string)       { f.was = value }
func (f *everyFact) WasNothing()            { f.wasNothing = true }
func (f *everyFact) Became(value string)    { f.became = value }
func (f *everyFact) BecameNothing()         { f.becameNothing = true }
func (f *everyFact) Through(surface string) { f.through = surface }
func (f *everyFact) ByHand()                { f.byHand = true }
func (f *everyFact) At(when time.Time)      { f.at = when }

func told(a Act) everyFact {
	var f everyFact
	a.Tell(&f)
	return f
}

func built() Act {
	var a AsTold
	a.On(4)
	a.Did(Changed)
	a.Changing("priority")
	a.Was("normal")
	a.Became("high")
	a.Through("mcp")
	a.At(time.Date(2026, 3, 3, 9, 15, 0, 0, time.UTC))
	return a.Act()
}

func TestTell_saysEveryFactAnActCarries(t *testing.T) {
	f := told(built())
	for _, tc := range []struct {
		what      string
		got, want any
	}{
		{"the task", f.task, 4},
		{"what was done", f.did, Changed},
		{"the attribute", f.field, "priority"},
		{"what it was", f.was, "normal"},
		{"what it became", f.became, "high"},
		{"the surface", f.through, "mcp"},
	} {
		if tc.got != tc.want {
			t.Errorf("%s = %v, want %v", tc.what, tc.got, tc.want)
		}
	}
	if f.at.IsZero() {
		t.Error("the moment was not told")
	}
}

func TestTell_saysWhatIsAbsentRatherThanSayingNothing(t *testing.T) {
	var a AsTold
	a.On(1)
	a.Did(FocusTaken)
	a.NothingNamed()
	a.WasNothing()
	a.BecameNothing()
	a.ByHand()
	a.At(time.Date(2026, 3, 3, 9, 0, 0, 0, time.UTC))

	f := told(a.Act())
	for _, tc := range []struct {
		what string
		told bool
	}{
		{"no attribute named", f.nothingNamed},
		{"nothing before", f.wasNothing},
		{"nothing after", f.becameNothing},
		{"done at the terminal", f.byHand},
	} {
		if !tc.told {
			t.Errorf("%s was not told", tc.what)
		}
	}
}

func TestAsTold_losesNothingOnTheWayThrough(t *testing.T) {
	first := built()

	var again AsTold
	first.Tell(&again)
	second := again.Act()

	if told(first) != told(second) {
		t.Errorf("told twice differs:\n  %+v\n  %+v", told(first), told(second))
	}
}

func TestAsTold_losesNothingOnTheWayThroughWhenNothingIsThere(t *testing.T) {
	var bare AsTold
	bare.On(1)
	bare.Did(Deleted)
	bare.NothingNamed()
	bare.WasNothing()
	bare.BecameNothing()
	bare.ByHand()
	bare.At(time.Date(2026, 3, 3, 9, 0, 0, 0, time.UTC))
	first := bare.Act()

	var again AsTold
	first.Tell(&again)
	if told(first) != told(again.Act()) {
		t.Errorf("told twice differs:\n  %+v\n  %+v", told(first), told(again.Act()))
	}
}

func TestLastNamed_isTheLastActThatNamedTheField(t *testing.T) {
	acts := []Act{
		acted(1, Created, "column", "", "backlog", "", 9, 0),
		acted(1, Changed, "column", "backlog", "todo", "", 10, 0),
		acted(1, Changed, "priority", "normal", "high", "", 11, 0),
	}
	if got := LastNamed(acts, "column"); !got.Equal(acts[1].at) {
		t.Errorf("LastNamed = %v, want the 10:00 change", got)
	}
	if got := LastNamed(acts, "deadline"); !got.IsZero() {
		t.Errorf("LastNamed of a field never named = %v, want nothing", got)
	}
}
