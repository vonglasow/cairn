package engine

// What the engine writes down, asked of the engine: the history it hands back,
// what a creation and a change record, and what a period reaches.

import (
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"

	"cairn/internal/domain"
	"cairn/internal/trace"
)

// actsOf reads back what was written down about a task, failing the test if the
// engine will not say.
func actsOf(t *testing.T, e *Engine, number int) []trace.Act {
	t.Helper()
	acts, err := e.History(number)
	if err != nil {
		t.Fatalf("History(%d): %v", number, err)
	}
	return acts
}

// didOf reads what an act says was done, and about which attribute.
func didOf(a trace.Act) (did, field, was, became string) {
	var f recordedFacts
	a.Tell(&f)
	return f.did, f.field, f.was, f.became
}

// recordedFacts is an act's membrane for what these tests read off one.
type recordedFacts struct {
	task                    int
	did, field, was, became string
	at                      time.Time
}

func (f *recordedFacts) On(n int)              { f.task = n }
func (f *recordedFacts) Did(what string)       { f.did = what }
func (f *recordedFacts) Changing(field string) { f.field = field }
func (f *recordedFacts) NothingNamed()         { f.field = "" }
func (f *recordedFacts) Was(value string)      { f.was = value }
func (f *recordedFacts) WasNothing()           { f.was = "" }
func (f *recordedFacts) Became(value string)   { f.became = value }
func (f *recordedFacts) BecameNothing()        { f.became = "" }
func (f *recordedFacts) Through(string)        {}
func (f *recordedFacts) ByHand()               {}
func (f *recordedFacts) At(when time.Time)     { f.at = when }

// A creation writes one act per attribute the task carries, each with what it
// started as. One act carrying them all could not be read back apart, which is
// the whole reason it is one each.
func TestRecordCreated_writesOneActPerAttributeItStartedWith(t *testing.T) {
	e := openEngine(t)
	added, err := e.Add("Write the spec", Changes{Priority: of("high"), Description: of("the one")})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	started := map[string]string{}
	for _, a := range actsOf(t, e, factsOf(added).number) {
		did, field, was, became := didOf(a)
		if did != trace.Created {
			t.Errorf("a creation wrote an act saying %q", did)
		}
		if field == "" {
			continue
		}
		if was != "" {
			t.Errorf("%q says it was %q before the task existed", field, was)
		}
		started[field] = became
	}
	for field, want := range map[string]string{
		"title": "Write the spec", "column": "backlog",
		"priority": "high", "description": "the one",
	} {
		if got := started[field]; got != want {
			t.Errorf("it started with %s = %q, want %q", field, got, want)
		}
	}
}

// An attribute the task does not carry writes no act: a creation says what a
// task started with, not what it started without.
func TestRecordCreated_writesNothingForWhatItDoesNotCarry(t *testing.T) {
	e := openEngine(t)
	added, err := e.Add("Write the spec", Changes{})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	for _, a := range actsOf(t, e, factsOf(added).number) {
		_, field, _, _ := didOf(a)
		for _, absent := range []string{"deadline", "estimate", "labels", "parent", "objective"} {
			if field == absent {
				t.Errorf("it wrote down %q, which the task does not carry", absent)
			}
		}
	}
}

// A change writes what the attribute went from and to, and writes nothing for
// an attribute given the value it already held.
func TestRecordChanges_writesOnlyWhatActuallyMoved(t *testing.T) {
	e := openEngine(t)
	added, err := e.Add("Write the spec", Changes{Priority: of("high")})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	number := factsOf(added).number
	born := len(actsOf(t, e, number))

	if _, err := e.Edit(number, Changes{Priority: of("high"), State: of("todo")}); err != nil {
		t.Fatalf("Edit: %v", err)
	}
	changes := actsOf(t, e, number)[born:]
	if len(changes) != 1 {
		t.Fatalf("%d acts written, want just the one that moved", len(changes))
	}
	did, field, was, became := didOf(changes[0])
	if did != trace.Changed || field != "column" || was != "backlog" || became != "todo" {
		t.Errorf("it wrote %q %q: %q → %q, want the column going from backlog to todo",
			did, field, was, became)
	}
}

// Nothing written down means nothing was ever there: a task always carries at
// least its creation, and a deletion takes nothing away from a history — which
// is what lets the absence of acts be the signal.
func TestHistory_refusesATaskNothingWasEverWrittenAbout(t *testing.T) {
	e := openEngine(t)
	add(t, e, "The only one")
	if _, err := e.History(99); !errors.Is(err, ErrNoSuchTask) {
		t.Errorf("History(99) = %v, want %v", err, ErrNoSuchTask)
	}
}

// And a task since deleted still has one, which is the case the refusal above
// must not catch.
func TestHistory_stillAnswersForATaskThatIsGone(t *testing.T) {
	e := openEngine(t)
	number := factsOf(add(t, e, "Gone soon")).number
	if err := e.Delete(number); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	acts := actsOf(t, e, number)
	if did, _, _, _ := didOf(acts[len(acts)-1]); did != trace.Deleted {
		t.Errorf("the last act is %q, want %q", did, trace.Deleted)
	}
}

// A period reaches across tasks from a day onward, and leaves out what came
// before it.
func TestSince_reachesFromADayOnward(t *testing.T) {
	e := openEngine(t)
	add(t, e, "First")
	add(t, e, "Second")

	today, err := e.Since(startOfToday())
	if err != nil {
		t.Fatalf("Since: %v", err)
	}
	if len(today) == 0 {
		t.Error("a period covering today reaches nothing, and two tasks were made in it")
	}

	later, err := e.Since(time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("Since: %v", err)
	}
	if len(later) != 0 {
		t.Errorf("%d acts in a period nothing falls into, want none", len(later))
	}
}

// Left out, a period starts at the beginning of today: a listing of every act
// cairn ever recorded is not a question anybody asked.
func TestSinceWhen_startsAtTodayWhenNoDayIsGiven(t *testing.T) {
	from, err := sinceWhen(Input{Given: map[string]string{}})
	if err != nil {
		t.Fatalf("sinceWhen: %v", err)
	}
	if !from.Equal(startOfToday()) {
		t.Errorf("sinceWhen = %s, want the start of today, %s", from, startOfToday())
	}
}

func TestSinceWhen_readsTheDayItWasGiven(t *testing.T) {
	from, err := sinceWhen(Input{Given: map[string]string{"since": "2026-03-03"}})
	if err != nil {
		t.Fatalf("sinceWhen: %v", err)
	}
	if got := from.Format("2006-01-02"); got != "2026-03-03" {
		t.Errorf("sinceWhen = %s, want 2026-03-03", got)
	}
}

func TestSinceWhen_refusesWhatIsNotADay(t *testing.T) {
	if _, err := sinceWhen(Input{Given: map[string]string{"since": "pasunjour"}}); err == nil {
		t.Error("sinceWhen read \"pasunjour\" as a day")
	} else if !strings.Contains(err.Error(), "pasunjour") {
		t.Errorf("the reason does not quote what it could not read: %v", err)
	}
}

// startOfToday is what a period starts at when no day was named.
func startOfToday() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

// Every kind of move is written down, and says where it went. Asserted for all
// three because each takes its own path through the store, and a recording
// forgotten on one of them would be invisible on the others.
func TestMove_isWrittenDownWithWhereItWent(t *testing.T) {
	for _, tc := range []struct {
		what string
		at   Place
		want string
	}{
		{"to the first place", First, "first"},
		{"to the last place", Last, "last"},
		{"behind another", Behind, "behind task"},
	} {
		e := openEngine(t)
		first := factsOf(add(t, e, "First")).number
		second := factsOf(add(t, e, "Second")).number
		born := len(actsOf(t, e, second))

		if err := e.Move(second, tc.at, first); err != nil {
			t.Fatalf("%s: Move: %v", tc.what, err)
		}
		written := actsOf(t, e, second)[born:]
		if len(written) != 1 {
			t.Fatalf("%s: %d acts written, want 1", tc.what, len(written))
		}
		did, _, _, became := didOf(written[0])
		if did != trace.Moved || !strings.Contains(became, tc.want) {
			t.Errorf("%s: it wrote %q %q, want a move saying %q", tc.what, did, became, tc.want)
		}
	}
}

// Taking a task up writes it down, and leaves whatever else is being worked on
// alone.
//
// It used to release the one before and name what took over, so that a cut
// could be told from a stop. Nothing is cut any more: several tasks are worked
// on at once, one per agent session, and their stretches overlap on purpose.
// What closes a stretch is letting go of that task, and nothing else.
func TestPutInFocus_writesTheTakeAndLeavesTheRestHeld(t *testing.T) {
	e := openEngine(t)
	first := factsOf(add(t, e, "First")).number
	second := factsOf(add(t, e, "Second")).number
	bornFirst := len(actsOf(t, e, first))
	bornSecond := len(actsOf(t, e, second))

	if _, err := e.PutInFocus(first); err != nil {
		t.Fatalf("PutInFocus: %v", err)
	}
	if _, err := e.PutInFocus(second); err != nil {
		t.Fatalf("PutInFocus: %v", err)
	}

	onFirst := actsOf(t, e, first)[bornFirst:]
	if len(onFirst) != 1 {
		t.Fatalf("%d acts on the task taken up first, want only its take", len(onFirst))
	}
	if did, _, _, _ := didOf(onFirst[0]); did != trace.FocusTaken {
		t.Errorf("the act is %q, want %q", did, trace.FocusTaken)
	}
	// And letting it go writes the release, which is what closes its stretch.
	if err := e.LetGoOf(first); err != nil {
		t.Fatalf("LetGoOf: %v", err)
	}
	released := actsOf(t, e, first)[bornFirst:]
	if len(released) != 2 {
		t.Fatalf("%d acts once it was let go, want a take and a release", len(released))
	}
	// The release names nothing: nothing took over. It used to name the task
	// that had, so a cut could be told from a stop — and with several held at
	// once there is no cut to tell.
	did, _, _, became := didOf(released[1])
	if did != trace.FocusReleased {
		t.Errorf("the second act is %q, want %q", did, trace.FocusReleased)
	}
	if became != "" {
		t.Errorf("the release says %q took over, and nothing did", became)
	}
	if written := actsOf(t, e, second)[bornSecond:]; len(written) != 1 {
		t.Errorf("%d acts on the task taken up second, want just its take", len(written))
	}
}

// Taking the focus where it already is writes nothing: a stretch closed and
// reopened on the same task at the same moment is a switch nobody made.
func TestPutInFocus_writesNothingWhenItIsAlreadyThere(t *testing.T) {
	e := openEngine(t)
	only := factsOf(add(t, e, "The only one")).number
	if _, err := e.PutInFocus(only); err != nil {
		t.Fatalf("PutInFocus: %v", err)
	}
	held := len(actsOf(t, e, only))
	if _, err := e.PutInFocus(only); err != nil {
		t.Fatalf("PutInFocus again: %v", err)
	}
	if got := len(actsOf(t, e, only)); got != held {
		t.Errorf("%d acts, want the %d that were already there", got, held)
	}
}

// Letting go writes a release naming nothing — a stop, not a cut. Letting go of
// nothing writes nothing at all, since a release with no stretch open would
// fault every duration around it.
func TestLetGo_writesAReleaseOnlyWhenSomethingWasHeld(t *testing.T) {
	e := openEngine(t)
	only := factsOf(add(t, e, "The only one")).number
	born := len(actsOf(t, e, only))

	if err := e.LetGo(); err != nil {
		t.Fatalf("LetGo with nothing held: %v", err)
	}
	if got := len(actsOf(t, e, only)); got != born {
		t.Errorf("%d acts after letting go of nothing, want the %d already there", got, born)
	}

	if _, err := e.PutInFocus(only); err != nil {
		t.Fatalf("PutInFocus: %v", err)
	}
	if err := e.LetGo(); err != nil {
		t.Fatalf("LetGo: %v", err)
	}
	written := actsOf(t, e, only)[born:]
	did, _, _, became := didOf(written[len(written)-1])
	if did != trace.FocusReleased {
		t.Errorf("the last act is %q, want %q", did, trace.FocusReleased)
	}
	if became != "" {
		t.Errorf("a stop names %q as taking over, and nothing did", became)
	}
}

// A task being worked on that goes away closes its stretch first, then records
// the deletion. Deleting some other task closes nothing.
func TestDelete_closesTheStretchOnlyOnTheTaskThatWasHeld(t *testing.T) {
	e := openEngine(t)
	held := factsOf(add(t, e, "Being worked on")).number
	other := factsOf(add(t, e, "Something else")).number
	if _, err := e.PutInFocus(held); err != nil {
		t.Fatalf("PutInFocus: %v", err)
	}

	bornOther := len(actsOf(t, e, other))
	if err := e.Delete(other); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	onOther := actsOf(t, e, other)[bornOther:]
	if len(onOther) != 1 {
		t.Fatalf("%d acts on the other task, want just its deletion", len(onOther))
	}
	if did, _, _, _ := didOf(onOther[0]); did != trace.Deleted {
		t.Errorf("the act is %q, want %q", did, trace.Deleted)
	}

	bornHeld := len(actsOf(t, e, held))
	if err := e.Delete(held); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	onHeld := actsOf(t, e, held)[bornHeld:]
	if len(onHeld) != 2 {
		t.Fatalf("%d acts on the task held, want a release and a deletion", len(onHeld))
	}
	if did, _, _, _ := didOf(onHeld[0]); did != trace.FocusReleased {
		t.Errorf("the first act is %q, want the stretch closed first", did)
	}
}

// Two attributes changing in one edit write two acts, not one: a recording that
// stopped after the first would look right on every edit that changes one thing.
func TestRecordChanges_writesEveryAttributeThatMoved(t *testing.T) {
	e := openEngine(t)
	number := factsOf(add(t, e, "Write the spec")).number
	born := len(actsOf(t, e, number))

	if _, err := e.Edit(number, Changes{State: of("todo"), Priority: of("high")}); err != nil {
		t.Fatalf("Edit: %v", err)
	}
	written := actsOf(t, e, number)[born:]
	if len(written) != 2 {
		t.Fatalf("%d acts written, want one per attribute that moved", len(written))
	}
	moved := map[string]bool{}
	for _, a := range written {
		_, field, _, _ := didOf(a)
		moved[field] = true
	}
	for _, want := range []string{"column", "priority"} {
		if !moved[want] {
			t.Errorf("%q moved and was not written down", want)
		}
	}
}

// A length of time is written down as a person reads one, and an absent one is
// not written down at all.
func TestRecordCreated_writesALengthOfTimeItStartedWith(t *testing.T) {
	e := openEngine(t)
	added, err := e.Add("Write the spec", Changes{Estimate: of("90m")})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	started := map[string]string{}
	for _, a := range actsOf(t, e, factsOf(added).number) {
		_, field, _, became := didOf(a)
		started[field] = became
	}
	if got := started["estimate"]; got != "1h30m" {
		t.Errorf("it started with estimate = %q, want %q", got, "1h30m")
	}
	if _, written := started["time spent"]; written {
		t.Error("it wrote down a time spent, and nothing had been spent")
	}
}

// The two operations are run here, not only the engine methods behind them: a
// Run body is where an argument becomes a call, and the lines that check what
// came back are exercised by nothing else.
func TestTraceShow_answersWithTheHistoryOrRefuses(t *testing.T) {
	e := openEngine(t)
	number := factsOf(add(t, e, "Write the spec")).number

	answer, err := TraceShow.Run(On{Engine: e}, Input{
		Given: map[string]string{"number": strconv.Itoa(number)},
	})
	if err != nil {
		t.Fatalf("TraceShow.Run: %v", err)
	}
	if len(answer.Acts) == 0 {
		t.Error("it answered with no acts, and the task has at least its creation")
	}

	if _, err := TraceShow.Run(On{Engine: e}, Input{
		Given: map[string]string{"number": "99"},
	}); !errors.Is(err, ErrNoSuchTask) {
		t.Errorf("TraceShow.Run on a task never written about = %v, want %v", err, ErrNoSuchTask)
	}
}

func TestTraceList_answersWithThePeriodOrRefusesTheDay(t *testing.T) {
	e := openEngine(t)
	add(t, e, "Write the spec")

	answer, err := TraceList.Run(On{Engine: e}, Input{Given: map[string]string{}})
	if err != nil {
		t.Fatalf("TraceList.Run: %v", err)
	}
	if len(answer.Since) == 0 {
		t.Error("today reaches nothing, and a task was made in it")
	}

	empty, err := TraceList.Run(On{Engine: e}, Input{
		Given: map[string]string{"since": "2099-01-01"},
	})
	if err != nil {
		t.Fatalf("TraceList.Run: %v", err)
	}
	if len(empty.Since) != 0 {
		t.Errorf("%d acts in a period nothing falls into, want none", len(empty.Since))
	}

	if _, err := TraceList.Run(On{Engine: e}, Input{
		Given: map[string]string{"since": "pasunjour"},
	}); err == nil {
		t.Error("TraceList.Run read \"pasunjour\" as a day")
	}
}

// A period is rendered as its acts and nothing else: how long the work took is
// the timesheet's to say, worked out day by day, and a second sum under the
// acts counted a focus left open for days as days of work.
func TestRenderText_rendersAPeriodAsItsActs(t *testing.T) {
	e := openEngine(t)
	first := factsOf(add(t, e, "First")).number
	second := factsOf(add(t, e, "Second")).number
	for _, one := range []int{first, second} {
		if _, err := e.PutInFocus(one); err != nil {
			t.Fatalf("PutInFocus: %v", err)
		}
	}
	acts, err := e.Since(startOfToday())
	if err != nil {
		t.Fatalf("Since: %v", err)
	}

	got := RenderText(TraceList, Answer{Since: acts})
	if !strings.Contains(got, "focus taken") {
		t.Errorf("the period does not carry its acts:\n%s", got)
	}
	if strings.Contains(got, "stretches of work") {
		t.Errorf("the period still adds itself up:\n%s", got)
	}
}

// The trace recorded that a task was deleted and not what the task was: no
// title, no description, nothing. Reading and undoing are two different asks and
// cairn answers only the first — but a trace that keeps no content makes even a
// manual restore impossible, which is a different thing from not having undo.
func TestDeletingWritesDownWhatWentWithIt(t *testing.T) {
	e := openEngine(t)
	spec := add(t, e, "Write the spec")
	number := factsOf(spec).number
	note := "the one that matters"
	high := string(domain.High)
	if _, err := e.Edit(number, Changes{Description: &note, Priority: &high}); err != nil {
		t.Fatal(err)
	}
	if err := e.Delete(number); err != nil {
		t.Fatal(err)
	}
	kept := ""
	for _, a := range actsOf(t, e, number) {
		if did, _, was, _ := didOf(a); did == trace.Deleted {
			kept = was
		}
	}
	for _, want := range []string{"Write the spec", note, high} {
		if !strings.Contains(kept, want) {
			t.Errorf("the trace kept %q, which does not carry %q", kept, want)
		}
	}
}
