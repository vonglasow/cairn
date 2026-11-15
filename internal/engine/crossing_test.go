package engine

import (
	"strings"
	"testing"

	"cairn/internal/engine/objective"
)

func TestOf_crossesFoundObjectives(t *testing.T) {
	e := openEngine(t)
	found := []objective.Objective{addObjective(t, e, "Ship cairn 1.0")}

	crossed, ok := Of(ObjectiveSearch, Answer{MatchedObjectives: found}).(ObjectivesListed)
	if !ok {
		t.Fatalf("Of = %T, want a listing of objectives", Of(ObjectiveSearch, Answer{MatchedObjectives: found}))
	}
	if len(crossed.Objectives) != 1 || crossed.Objectives[0].Title != "Ship cairn 1.0" {
		t.Errorf("crossed %+v, want the objective that was found", crossed.Objectives)
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
		err := TaskShow.check(tc.in)
		if err == nil {
			t.Errorf("%s: Check said nothing", tc.what)
			continue
		}
		if !strings.Contains(err.Error(), tc.says) {
			t.Errorf("%s: Check said %q, want it to say %q", tc.what, err, tc.says)
		}
	}
}
