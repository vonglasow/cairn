package engine

import (
	"strings"
	"testing"
	"time"

	"cairn/internal/domain"
	"cairn/internal/objective"
	"cairn/internal/task"
	"cairn/internal/trace"
)

func anAct(task int, did string) trace.Act {
	var a trace.AsTold
	a.On(task)
	a.Did(did)
	a.At(time.Date(2026, 3, 3, 9, 0, 0, 0, time.UTC))
	return a.Act()
}

func TestOf_crossesAHistoryAndAPeriodAsActs(t *testing.T) {
	for _, tc := range []struct {
		what string
		op   Operation
		in   Answer
	}{
		{"a task's history", TraceShow, Answer{Acts: []trace.Act{anAct(1, trace.Created)}}},
		{"a period", TraceList, Answer{Since: []trace.Act{anAct(1, trace.FocusTaken)}}},
	} {
		crossed, ok := Of(tc.op, tc.in).(Acts)
		if !ok {
			t.Errorf("%s: Of = %T, want a listing of acts", tc.what, Of(tc.op, tc.in))
			continue
		}
		if len(crossed.Acts) != 1 || crossed.Acts[0].Task != 1 {
			t.Errorf("%s: crossed %+v, want one act on task 1", tc.what, crossed.Acts)
		}
	}
}

func TestRenderText_rendersAHistory(t *testing.T) {
	got := RenderText(TraceShow, Answer{Acts: []trace.Act{anAct(1, trace.Created)}})
	if !strings.Contains(got, trace.Created) {
		t.Errorf("the history does not carry what was done:\n%s", got)
	}
}

func TestOf_crossesFoundObjectives(t *testing.T) {
	e := openEngine(t)
	found := []objective.Objective{addObjective(t, e, "Ship cairn 1.0")}

	crossed, ok := Of(ObjectiveSearch, Answer{MatchedObjectives: found}).(Objectives)
	if !ok {
		t.Fatalf("Of = %T, want a listing of objectives", Of(ObjectiveSearch, Answer{MatchedObjectives: found}))
	}
	if len(crossed.Objectives) != 1 || crossed.Objectives[0].Title != "Ship cairn 1.0" {
		t.Errorf("crossed %+v, want the objective that was found", crossed.Objectives)
	}
	if text := RenderText(ObjectiveSearch, Answer{MatchedObjectives: found}); !strings.Contains(text, "Ship cairn 1.0") {
		t.Errorf("the listing does not carry it:\n%s", text)
	}
}

func TestOfObjective_namesTheTasksUnderIt(t *testing.T) {
	e := openEngine(t)
	served := addObjective(t, e, "Ship cairn 1.0")
	under, err := e.add("Write the spec", Changes{Objective: serves(factsOfObjective(served).Number)})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	crossed := OfObjective(served, []task.Task{under}, domain.Progress{})
	if len(crossed.Tasks) != 1 || crossed.Tasks[0] != factsOf(under).Number {
		t.Errorf("it names %v, want the one task under it", crossed.Tasks)
	}
	if text := RenderText(ObjectiveShow, Answer{Objective: &served, Under: []task.Task{under}}); !strings.Contains(text, "Write the spec") {
		t.Errorf("the card does not name what serves it:\n%s", text)
	}
}

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

func TestEditObjective_retitlesIt(t *testing.T) {
	e := openEngine(t)
	added := addObjective(t, e, "Ship cairn 1.0")
	edited, err := e.editObjective(factsOfObjective(added).Number,
		ObjectiveChanges{Title: of("Ship cairn 1.1")})
	if err != nil {
		t.Fatalf("EditObjective: %v", err)
	}
	if got := factsOfObjective(edited).Title; got != "Ship cairn 1.1" {
		t.Errorf("the title is %q, want %q", got, "Ship cairn 1.1")
	}
}
