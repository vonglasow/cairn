package trace

// What an act tells, and that nothing is lost telling it and being told it
// again. Nothing here reaches an act's fields: a test that did would be the one
// place allowed to, and the membrane would stop being the only way in.

import (
	"testing"
	"time"
)

// everyFact is an act's membrane for a test: it keeps every fact it is told and
// counts what it was told about absence, so a test can check both that a fact
// arrived and that its absence did.
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

// told hands back everything an act says.
func told(a Act) everyFact {
	var f everyFact
	a.Tell(&f)
	return f
}

// built makes an act carrying everything, so a test can check every fact
// survives the round trip.
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

// Absence is told, and told explicitly. An act that said nothing about an
// attribute it does not have would leave whoever is listening to guess whether
// they were told or missed it.
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

// The membrane both ways round: what an act is told is what it tells, so a
// store reading a row it wrote gets back the act it wrote.
func TestAsTold_losesNothingOnTheWayThrough(t *testing.T) {
	first := built()

	var again AsTold
	first.Tell(&again)
	second := again.Act()

	if told(first) != told(second) {
		t.Errorf("told twice differs:\n  %+v\n  %+v", told(first), told(second))
	}
}

// And the same for an act carrying nothing but its own absences.
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

// TellOrNot is what everyone who builds an act uses for the four facts that can
// be absent. Tested on plain recorders rather than through a fixture that uses
// it: a fixture calling the function under test moves with it, so breaking one
// breaks both and every assertion still holds — measured, in this repository,
// on this very helper.
func TestTellOrNot_tellsTheValueOrTellsThereIsNone(t *testing.T) {
	for _, tc := range []struct {
		value        string
		wantTold     string
		wantedAbsent bool
	}{
		{"high", "high", false},
		{"", "", true},
	} {
		told, absent := "", false
		TellOrNot(tc.value, func(v string) { told = v }, func() { absent = true })
		if told != tc.wantTold {
			t.Errorf("TellOrNot(%q) told %q, want %q", tc.value, told, tc.wantTold)
		}
		if absent != tc.wantedAbsent {
			t.Errorf("TellOrNot(%q) told absence = %v, want %v", tc.value, absent, tc.wantedAbsent)
		}
	}
}
