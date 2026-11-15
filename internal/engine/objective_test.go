package engine

import (
	"cairn/internal/engine/item"
	"errors"
	"slices"
	"strings"
	"testing"

	"cairn/internal/engine/objective"
)

func TestObjective_anEditToABlankTitleSaysWhatWasTyped(t *testing.T) {
	path := storeAt(t)
	if _, err := Do(ObjectiveNew, path, given("title", "Ship cairn")); err != nil {
		t.Fatal(err)
	}
	_, err := Do(ObjectiveEdit, path, given("number", "1", "title", "   "))
	if !errors.Is(err, item.ErrTitleRequired) || !strings.HasSuffix(err.Error(), `, got "   "`) {
		t.Errorf("err = %v, want the title typed said", err)
	}
}

func TestObjective_searchFindsOnlyWhatCarriesIt(t *testing.T) {
	path := storeAt(t)
	for _, title := range []string{"Ship cairn", "Hire someone"} {
		if _, err := Do(ObjectiveNew, path, given("title", title)); err != nil {
			t.Fatal(err)
		}
	}
	found, err := Do(ObjectiveSearch, path, Input{Listed: map[string][]string{"named": {"ship"}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(found.MatchedObjectives) != 1 || objective.OfObjective(found.MatchedObjectives[0]).Title != "Ship cairn" {
		t.Errorf("found %v, want Ship cairn alone", found.MatchedObjectives)
	}
}

func TestObjectiveShown_withSubObjectivesSaysTheTasksApart(t *testing.T) {
	e := openEngine(t)
	addObjective(t, e, "Ship")
	if _, err := e.Ask(ObjectiveNew, given("title", "Write", "parent", "1")); err != nil {
		t.Fatal(err)
	}
	serving(t, e, "Draft", 1)
	serving(t, e, "Chapter", 2)
	a, err := e.Ask(ObjectiveShow, given("number", "1"))
	if err != nil {
		t.Fatal(err)
	}
	if said := Of(ObjectiveShow, a).(ObjectiveRow); !slices.Equal(said.Children, []int{2}) || !slices.Equal(said.Tasks, []int{1, 2}) {
		t.Errorf("it answers children %v and tasks %v, want 2 and 1, 2", said.Children, said.Tasks)
	}
}

func TestObjectiveDelete_deletesEveryTaskItIsToldTo(t *testing.T) {
	e := openEngine(t)
	addObjective(t, e, "Ship")
	serving(t, e, "one", 1)
	serving(t, e, "two", 1)
	if err := objectiveDeleted(e, map[string]string{"number": "1", "tasks": "delete"}); err != nil {
		t.Fatal(err)
	}
	if left, _ := e.store.All(); len(left) != 0 {
		t.Errorf("%d tasks are left", len(left))
	}
}
