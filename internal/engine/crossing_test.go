package engine

// What an answer crosses as, for the fields whose own package never asks.
//
// Every one of these is exercised by a scenario, through a surface — and that
// credits the surface, not this package. Coverage is measured per package, so
// code called from elsewhere and never from home reads as untested however well
// it is proven. These close that gap where it is worth closing: the dispatch
// that decides which shape an answer takes.

import (
	"strings"
	"testing"
	"time"

	"cairn/internal/domain"
	"cairn/internal/objective"
	"cairn/internal/task"
	"cairn/internal/trace"
)

// anAct builds an act, for the two fields that carry them.
func anAct(task int, did string) trace.Act {
	var a trace.AsTold
	a.On(task)
	a.Did(did)
	a.NothingNamed()
	a.WasNothing()
	a.BecameNothing()
	a.ByHand()
	a.At(time.Date(2026, 3, 3, 9, 0, 0, 0, time.UTC))
	return a.Act()
}

// A history and a period cross as the same shape, since a machine reading
// either does not care which question was asked.
func TestOf_crossesAHistoryAndAPeriodAsActs(t *testing.T) {
	for _, tc := range []struct {
		what string
		in   Answer
	}{
		{"a task's history", Answer{Acts: []trace.Act{anAct(1, trace.Created)}}},
		{"a period", Answer{Since: []trace.Act{anAct(1, trace.FocusTaken)}}},
	} {
		crossed, ok := Of(tc.in).(Acts)
		if !ok {
			t.Errorf("%s: Of = %T, want a listing of acts", tc.what, Of(tc.in))
			continue
		}
		if len(crossed.Acts) != 1 || crossed.Acts[0].Task != 1 {
			t.Errorf("%s: crossed %+v, want one act on task 1", tc.what, crossed.Acts)
		}
	}
}

// A history renders as a card a person reads; a period renders as its acts and
// then what they come to.
func TestRenderText_rendersAHistory(t *testing.T) {
	got := RenderText(TraceShow, Answer{Acts: []trace.Act{anAct(1, trace.Created)}})
	if !strings.Contains(got, trace.Created) {
		t.Errorf("the history does not carry what was done:\n%s", got)
	}
}

// What an objective search found crosses as objectives, and reads as a listing:
// its own field beside Objectives, so nobody has to ask which one they got.
func TestOf_crossesFoundObjectives(t *testing.T) {
	e := openEngine(t)
	found := []objective.Objective{addObjective(t, e, "Ship cairn 1.0")}

	crossed, ok := Of(Answer{MatchedObjectives: found}).(Objectives)
	if !ok {
		t.Fatalf("Of = %T, want a listing of objectives", Of(Answer{MatchedObjectives: found}))
	}
	if len(crossed.Objectives) != 1 || crossed.Objectives[0].Title != "Ship cairn 1.0" {
		t.Errorf("crossed %+v, want the objective that was found", crossed.Objectives)
	}
	if text := RenderText(ObjectiveSearch, Answer{MatchedObjectives: found}); !strings.Contains(text, "Ship cairn 1.0") {
		t.Errorf("the listing does not carry it:\n%s", text)
	}
}

// An objective crossing with tasks under it names their numbers, which is the
// loop nothing else in this package walks.
func TestOfObjective_namesTheTasksUnderIt(t *testing.T) {
	e := openEngine(t)
	served := addObjective(t, e, "Ship cairn 1.0")
	under, err := e.Add("Write the spec", Changes{Objective: serves(factsOfObjective(served).number)})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	crossed := OfObjective(served, []task.Task{under}, domain.Progress{})
	if len(crossed.Tasks) != 1 || crossed.Tasks[0] != factsOf(under).number {
		t.Errorf("it names %v, want the one task under it", crossed.Tasks)
	}
	if text := RenderText(ObjectiveShow, Answer{Objective: &served, Under: []task.Task{under}}); !strings.Contains(text, "Write the spec") {
		t.Errorf("the card does not name what serves it:\n%s", text)
	}
}

// Check refuses before anything is opened, and each of its three refusals is
// reached from a surface's scenarios rather than from here.
func TestCheck_refusesWhatNoOperationCanMakeSenseOf(t *testing.T) {
	for _, tc := range []struct {
		what string
		in   Input
		says string
	}{
		{
			"an argument named that the operation does not take",
			Input{Given: map[string]string{"number": "1", "nope": "x"}},
			"takes no argument by that name",
		},
		{
			"a list argument named that the operation does not take",
			Input{Given: map[string]string{"number": "1"}, Listed: map[string][]string{"nope": {"x"}}},
			"takes no argument by that name",
		},
		{
			"a required argument left out",
			Input{Given: map[string]string{}},
			"was not given one",
		},
	} {
		err := TaskShow.Check(tc.in)
		if err == nil {
			t.Errorf("%s: Check said nothing", tc.what)
			continue
		}
		if !strings.Contains(err.Error(), tc.says) {
			t.Errorf("%s: Check said %q, want it to say %q", tc.what, err, tc.says)
		}
	}
}

// Retitling an objective goes through a path of its own, since a title is the
// one thing an objective cannot be left without.
func TestEditObjective_retitlesIt(t *testing.T) {
	e := openEngine(t)
	added := addObjective(t, e, "Ship cairn 1.0")
	edited, err := e.EditObjective(factsOfObjective(added).number,
		ObjectiveChanges{Title: of("Ship cairn 1.1")})
	if err != nil {
		t.Fatalf("EditObjective: %v", err)
	}
	if got := factsOfObjective(edited).title; got != "Ship cairn 1.1" {
		t.Errorf("the title is %q, want %q", got, "Ship cairn 1.1")
	}
}
