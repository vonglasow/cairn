package store

// What the store keeps of what happened. Written down through Record and read
// back through the two ways of asking, so what a test checks is what came off
// the disk and not what it handed over.

import (
	"testing"
	"time"

	"cairn/internal/trace"
)

// happened builds an act to write down.
//
// Written out rather than through trace.TellOrNot, which is what the code under
// test uses. Measured: sharing it made a mutation of it invisible — the fixture
// and the code moved together, so breaking one broke both and every assertion
// still held.
func happened(task int, did, field, was, became, through string, at time.Time) trace.Act {
	var a trace.AsTold
	a.On(task)
	a.Did(did)
	if field == "" {
		a.NothingNamed()
	} else {
		a.Changing(field)
	}
	if was == "" {
		a.WasNothing()
	} else {
		a.Was(was)
	}
	if became == "" {
		a.BecameNothing()
	} else {
		a.Became(became)
	}
	if through == "" {
		a.ByHand()
	} else {
		a.Through(through)
	}
	a.At(at)
	return a.Act()
}

// actFacts is an act's membrane for what a test reads back off one.
type actFacts struct {
	task                             int
	did, field, was, became, through string
	at                               time.Time
}

func (f *actFacts) On(n int)               { f.task = n }
func (f *actFacts) Did(what string)        { f.did = what }
func (f *actFacts) Changing(field string)  { f.field = field }
func (f *actFacts) NothingNamed()          { f.field = "" }
func (f *actFacts) Was(value string)       { f.was = value }
func (f *actFacts) WasNothing()            { f.was = "" }
func (f *actFacts) Became(value string)    { f.became = value }
func (f *actFacts) BecameNothing()         { f.became = "" }
func (f *actFacts) Through(surface string) { f.through = surface }
func (f *actFacts) ByHand()                { f.through = "" }
func (f *actFacts) At(when time.Time)      { f.at = when }

// factsOfAct reads back what an act says.
func factsOfAct(a trace.Act) actFacts {
	var f actFacts
	a.Tell(&f)
	return f
}

// record writes an act down, failing the test if it cannot.
func record(t *testing.T, s *Store, a trace.Act) {
	t.Helper()
	if err := s.Record(a); err != nil {
		t.Fatalf("Record: %v", err)
	}
}

// aMoment is a moment these tests can name.
func aMoment(hour int) time.Time {
	return time.Date(2026, 3, 3, hour, 0, 0, 0, time.UTC)
}

// Every fact survives the round trip through the disk, the moment included —
// which is the one that could have been lost, since it goes down as text.
func TestRecord_keepsEveryFactAnActCarries(t *testing.T) {
	s := openStore(t)
	written := happened(4, trace.Changed, "priority", "normal", "high", "mcp", aMoment(9))
	record(t, s, written)

	acts, err := s.ActsOn(4)
	if err != nil {
		t.Fatalf("ActsOn: %v", err)
	}
	if len(acts) != 1 {
		t.Fatalf("%d acts, want 1", len(acts))
	}
	got, want := factsOfAct(acts[0]), factsOfAct(written)
	if !got.at.Equal(want.at) {
		t.Errorf("the moment came back as %s, want %s", got.at, want.at)
	}
	got.at, want.at = time.Time{}, time.Time{}
	if got != want {
		t.Errorf("what came back differs:\n  %+v\n  %+v", got, want)
	}
}

// An act carrying nothing but its own absences survives too: absent must not
// come back as something.
func TestRecord_keepsWhatWasAbsentAbsent(t *testing.T) {
	s := openStore(t)
	record(t, s, happened(1, trace.FocusTaken, "", "", "", "", aMoment(9)))

	acts, err := s.ActsOn(1)
	if err != nil {
		t.Fatalf("ActsOn: %v", err)
	}
	f := factsOfAct(acts[0])
	if f.field != "" || f.was != "" || f.became != "" || f.through != "" {
		t.Errorf("something absent came back as something: %+v", f)
	}
}

// Oldest first, and by the order they were written rather than by their moment:
// two acts of one command land in the same second, and what a history means is
// the order they happened in.
func TestActsOn_handsThemBackOldestFirst(t *testing.T) {
	s := openStore(t)
	for _, did := range []string{trace.Created, trace.FocusTaken, trace.FocusReleased} {
		record(t, s, happened(1, did, "", "", "", "", aMoment(9)))
	}
	acts, err := s.ActsOn(1)
	if err != nil {
		t.Fatalf("ActsOn: %v", err)
	}
	for i, want := range []string{trace.Created, trace.FocusTaken, trace.FocusReleased} {
		if got := factsOfAct(acts[i]).did; got != want {
			t.Errorf("act %d is %q, want %q", i+1, got, want)
		}
	}
}

// One task's history is that task's, and says nothing about another's.
func TestActsOn_keepsOneTasksHistoryToItself(t *testing.T) {
	s := openStore(t)
	record(t, s, happened(1, trace.Created, "", "", "", "", aMoment(9)))
	record(t, s, happened(2, trace.Created, "", "", "", "", aMoment(9)))

	acts, err := s.ActsOn(1)
	if err != nil {
		t.Fatalf("ActsOn: %v", err)
	}
	if len(acts) != 1 || factsOfAct(acts[0]).task != 1 {
		t.Errorf("task 1's history carries %d acts, and the first is on task %d",
			len(acts), factsOfAct(acts[0]).task)
	}
}

// A task nothing was written down about answers with nothing rather than
// failing: what that means is the engine's to decide, not the store's.
func TestActsOn_answersNothingForATaskItKnowsNothingAbout(t *testing.T) {
	acts, err := openStore(t).ActsOn(99)
	if err != nil {
		t.Fatalf("ActsOn: %v", err)
	}
	if len(acts) != 0 {
		t.Errorf("%d acts, want none", len(acts))
	}
}

// A period reaches across tasks, from a moment onward, and leaves out what came
// before it.
func TestActsSince_reachesAcrossTasksFromAMomentOnward(t *testing.T) {
	s := openStore(t)
	record(t, s, happened(1, trace.Created, "", "", "", "", aMoment(8)))
	record(t, s, happened(1, trace.FocusTaken, "", "", "", "", aMoment(10)))
	record(t, s, happened(2, trace.Created, "", "", "", "", aMoment(11)))

	acts, err := s.ActsSince(aMoment(9))
	if err != nil {
		t.Fatalf("ActsSince: %v", err)
	}
	if len(acts) != 2 {
		t.Fatalf("%d acts, want the 2 from 9 o'clock onward", len(acts))
	}
	if got := factsOfAct(acts[0]).did; got != trace.FocusTaken {
		t.Errorf("the first act is %q, want %q", got, trace.FocusTaken)
	}
	if got := factsOfAct(acts[1]).task; got != 2 {
		t.Errorf("the second act is on task %d, want 2 — a period reaches across tasks", got)
	}
}

// A period nothing falls into answers with nothing.
func TestActsSince_answersNothingForAPeriodNothingFallsInto(t *testing.T) {
	s := openStore(t)
	record(t, s, happened(1, trace.Created, "", "", "", "", aMoment(9)))

	acts, err := s.ActsSince(time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("ActsSince: %v", err)
	}
	if len(acts) != 0 {
		t.Errorf("%d acts, want none", len(acts))
	}
}

// A history outlives what it describes: the acts of a deleted task are still
// there, which is the whole reason they are not a child of the task.
func TestActsOn_outlivesTheTaskItDescribes(t *testing.T) {
	s := openStore(t)
	kept := keep(t, s, "Gone soon")
	number := factsOf(kept).number
	record(t, s, happened(number, trace.Created, "", "", "", "", aMoment(9)))

	if err := s.Delete(number); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	acts, err := s.ActsOn(number)
	if err != nil {
		t.Fatalf("ActsOn: %v", err)
	}
	if len(acts) != 1 {
		t.Errorf("%d acts left after the task went, want 1", len(acts))
	}
}

// A moment the store cannot read is a store that has been written to by
// something else, and it says so rather than handing back a zero moment nobody
// would question.
func TestActsOn_refusesAMomentItCannotRead(t *testing.T) {
	s := openStore(t)
	if _, err := s.db.Exec(
		`INSERT INTO trace (task_number, did, field, was, became, through, at)
		 VALUES (1, 'created', '', '', '', '', 'not a moment')`); err != nil {
		t.Fatalf("writing a moment by hand: %v", err)
	}
	if _, err := s.ActsOn(1); err == nil {
		t.Error("ActsOn read a moment that is not one, and said nothing")
	}
}
