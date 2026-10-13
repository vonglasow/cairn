package engine

import (
	"errors"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"cairn/internal/domain"
	"cairn/internal/trace"
)

func actsOf(t *testing.T, e *Engine, number int) []trace.Act {
	t.Helper()
	acts, err := e.historyOf(number)
	if err != nil {
		t.Fatalf("History(%d): %v", number, err)
	}
	return acts
}

func didOf(a trace.Act) (did, field, was, became string) {
	var f recordedFacts
	a.Tell(&f)
	return f.did, f.field, f.was, f.became
}

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

func TestRecordCreated_writesOneActPerAttributeItStartedWith(t *testing.T) {
	e := openEngine(t)
	added, err := e.add("Write the spec", Changes{Priority: of("high"), Description: of("the one")})
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

func TestRecordCreated_writesNothingForWhatItDoesNotCarry(t *testing.T) {
	e := openEngine(t)
	added, err := e.add("Write the spec", Changes{})
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

func TestRecordChanges_writesOnlyWhatActuallyMoved(t *testing.T) {
	e := openEngine(t)
	added, err := e.add("Write the spec", Changes{Priority: of("high")})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	number := factsOf(added).number
	born := len(actsOf(t, e, number))

	if _, err := e.edit(number, Changes{Priority: of("high"), State: of("todo")}); err != nil {
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

func TestHistory_refusesATaskNothingWasEverWrittenAbout(t *testing.T) {
	e := openEngine(t)
	add(t, e, "The only one")
	if _, err := e.historyOf(99); !errors.Is(err, ErrNoSuchTask) {
		t.Errorf("History(99) = %v, want %v", err, ErrNoSuchTask)
	}
}

func TestHistory_stillAnswersForATaskThatIsGone(t *testing.T) {
	e := openEngine(t)
	number := factsOf(add(t, e, "Gone soon")).number
	if err := e.delete(number); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	acts := actsOf(t, e, number)
	if did, _, _, _ := didOf(acts[len(acts)-1]); did != trace.Deleted {
		t.Errorf("the last act is %q, want %q", did, trace.Deleted)
	}
}

func TestSince_reachesFromADayOnward(t *testing.T) {
	e := openEngine(t)
	add(t, e, "First")
	add(t, e, "Second")

	today, err := e.since(startOfToday())
	if err != nil {
		t.Fatalf("Since: %v", err)
	}
	if len(today) == 0 {
		t.Error("a period covering today reaches nothing, and two tasks were made in it")
	}

	later, err := e.since(time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("Since: %v", err)
	}
	if len(later) != 0 {
		t.Errorf("%d acts in a period nothing falls into, want none", len(later))
	}
}

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

func startOfToday() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

func TestMove_isWrittenDownWithWhereItWent(t *testing.T) {
	for _, tc := range []struct {
		what   string
		at     Place
		moving int
		want   string
	}{
		{"to the first place", First, 2, "first"},
		{"to the last place", Last, 0, "last"},
		{"behind another", Behind, 2, "behind task"},
	} {
		e := openEngine(t)
		first, second, third := three(t, e)
		moving := []int{first, second, third}[tc.moving]
		born := len(actsOf(t, e, moving))

		if err := e.move(moving, tc.at, first); err != nil {
			t.Fatalf("%s: Move: %v", tc.what, err)
		}
		written := actsOf(t, e, moving)[born:]
		if len(written) != 1 {
			t.Fatalf("%s: %d acts written, want 1", tc.what, len(written))
		}
		did, _, _, became := didOf(written[0])
		if did != trace.Moved || !strings.Contains(became, tc.want) {
			t.Errorf("%s: it wrote %q %q, want a move saying %q", tc.what, did, became, tc.want)
		}
	}
}

func TestPutInFocus_writesTheTakeAndLeavesTheRestHeld(t *testing.T) {
	e := openEngine(t)
	first := factsOf(add(t, e, "First")).number
	second := factsOf(add(t, e, "Second")).number
	bornFirst := len(actsOf(t, e, first))
	bornSecond := len(actsOf(t, e, second))

	if _, err := e.putInFocus(first); err != nil {
		t.Fatalf("PutInFocus: %v", err)
	}
	if _, err := e.putInFocus(second); err != nil {
		t.Fatalf("PutInFocus: %v", err)
	}

	onFirst := actsOf(t, e, first)[bornFirst:]
	if len(onFirst) != 1 {
		t.Fatalf("%d acts on the task taken up first, want only its take", len(onFirst))
	}
	if did, _, _, _ := didOf(onFirst[0]); did != trace.FocusTaken {
		t.Errorf("the act is %q, want %q", did, trace.FocusTaken)
	}
	if err := e.letGoOf(first); err != nil {
		t.Fatalf("LetGoOf: %v", err)
	}
	released := actsOf(t, e, first)[bornFirst:]
	if len(released) != 2 {
		t.Fatalf("%d acts once it was let go, want a take and a release", len(released))
	}
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

func TestPutInFocus_writesNothingWhenItIsAlreadyThere(t *testing.T) {
	e := openEngine(t)
	only := factsOf(add(t, e, "The only one")).number
	if _, err := e.putInFocus(only); err != nil {
		t.Fatalf("PutInFocus: %v", err)
	}
	held := len(actsOf(t, e, only))
	if _, err := e.putInFocus(only); err != nil {
		t.Fatalf("PutInFocus again: %v", err)
	}
	if got := len(actsOf(t, e, only)); got != held {
		t.Errorf("%d acts, want the %d that were already there", got, held)
	}
}

func TestLetGo_writesAReleaseOnlyWhenSomethingWasHeld(t *testing.T) {
	e := openEngine(t)
	only := factsOf(add(t, e, "The only one")).number
	born := len(actsOf(t, e, only))

	if err := e.letGo(); err != nil {
		t.Fatalf("LetGo with nothing held: %v", err)
	}
	if got := len(actsOf(t, e, only)); got != born {
		t.Errorf("%d acts after letting go of nothing, want the %d already there", got, born)
	}

	if _, err := e.putInFocus(only); err != nil {
		t.Fatalf("PutInFocus: %v", err)
	}
	if err := e.letGo(); err != nil {
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

func TestDelete_closesTheStretchOnlyOnTheTaskThatWasHeld(t *testing.T) {
	e := openEngine(t)
	held := factsOf(add(t, e, "Being worked on")).number
	other := factsOf(add(t, e, "Something else")).number
	if _, err := e.putInFocus(held); err != nil {
		t.Fatalf("PutInFocus: %v", err)
	}

	bornOther := len(actsOf(t, e, other))
	if err := e.delete(other); err != nil {
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
	if err := e.delete(held); err != nil {
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

func TestRecordChanges_writesEveryAttributeThatMoved(t *testing.T) {
	e := openEngine(t)
	number := factsOf(add(t, e, "Write the spec")).number
	born := len(actsOf(t, e, number))

	if _, err := e.edit(number, Changes{State: of("todo"), Priority: of("high")}); err != nil {
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

func TestRecordCreated_writesALengthOfTimeItStartedWith(t *testing.T) {
	e := openEngine(t)
	added, err := e.add("Write the spec", Changes{Estimate: of("90m")})
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

func TestRenderText_rendersAPeriodAsItsActs(t *testing.T) {
	e := openEngine(t)
	first := factsOf(add(t, e, "First")).number
	second := factsOf(add(t, e, "Second")).number
	for _, one := range []int{first, second} {
		if _, err := e.putInFocus(one); err != nil {
			t.Fatalf("PutInFocus: %v", err)
		}
	}
	acts, err := e.since(startOfToday())
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

func TestDeletingWritesDownWhatWentWithIt(t *testing.T) {
	e := openEngine(t)
	spec := add(t, e, "Write the spec")
	number := factsOf(spec).number
	note := "the one that matters"
	high := string(domain.High)
	if _, err := e.edit(number, Changes{Description: &note, Priority: &high}); err != nil {
		t.Fatal(err)
	}
	if err := e.delete(number); err != nil {
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

func TestArchiving_isWrittenDown(t *testing.T) {
	e := openEngine(t)
	n := factsOf(add(t, e, "Ship it")).number
	if err := e.archive(n); err != nil {
		t.Fatal(err)
	}
	if err := e.unarchive(n); err != nil {
		t.Fatal(err)
	}
	acts, err := e.historyOf(n)
	if err != nil {
		t.Fatal(err)
	}
	var said []string
	for _, a := range acts {
		if did, field, was, became := didOf(a); did == trace.Changed && field == "archived" {
			said = append(said, was+"→"+became)
		}
	}
	if want := []string{"→yes", "yes→"}; !slices.Equal(said, want) {
		t.Errorf("the history says %v about archiving, want %v", said, want)
	}
}

func TestShow_saysSinceWhenATaskHasStoodInItsColumn(t *testing.T) {
	path := storeAt(t, "Write the spec")
	made, err := Do(TaskShow, path, given("number", "1"))
	if err != nil {
		t.Fatal(err)
	}
	if made.StoodSince.IsZero() {
		t.Fatal("a task just made says nothing of since when it has stood where it is")
	}
	if _, err := Do(TaskEdit, path, given("number", "1", "priority", "high")); err != nil {
		t.Fatal(err)
	}
	if _, err := Do(TaskEdit, path, given("number", "1", "state", "current")); err != nil {
		t.Fatal(err)
	}
	moved, err := Do(TaskShow, path, given("number", "1"))
	if err != nil {
		t.Fatal(err)
	}
	acts, err := Do(TraceShow, path, given("number", "1"))
	if err != nil {
		t.Fatal(err)
	}
	var last time.Time
	for _, a := range acts.Acts {
		var f momentOf
		a.Tell(&f)
		if f.field == "column" {
			last = f.at
		}
	}
	if !moved.StoodSince.Equal(last) {
		t.Errorf("StoodSince = %v, want %v, the moment the column last changed", moved.StoodSince, last)
	}
}

type momentOf struct {
	actFacts
	at time.Time
}

func (m *momentOf) At(t time.Time) { m.at = t }
