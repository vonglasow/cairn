package engine

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"cairn/internal/domain"
	"cairn/internal/objective"
	"cairn/internal/task"
)

func TestObjective_aTaskServingNoObjectiveThatExistsIsRefused(t *testing.T) {
	e := openEngine(t)
	if _, err := e.add("Write the spec", Changes{Objective: serves(7)}); err == nil {
		t.Error("a task serving an objective that does not exist was added")
	}
	addObjective(t, e, "Ship cairn")
	if _, err := e.add("Ship it", Changes{Objective: serves(1, 7)}); err == nil {
		t.Error("a task serving one objective that exists and one that does not was added")
	}
}

func TestObjective_progressCountsWhatServesIt(t *testing.T) {
	e := openEngine(t)
	addObjective(t, e, "Ship cairn")
	serving(t, e, "Write the spec", 1)
	got, err := e.objectiveProgress(1)
	if err != nil {
		t.Fatal(err)
	}
	if got.Tasks.Total != 1 {
		t.Errorf("progress = %+v, want one task counted", got)
	}
}

func TestObjective_stepsAreTalliedPastATaskWithNone(t *testing.T) {
	e := openEngine(t)
	add(t, e, "No steps")
	steps := []task.Step{{Title: "one", Done: true}, {Title: "two"}}
	if _, err := e.add("Two steps", Changes{Steps: &steps}); err != nil {
		t.Fatal(err)
	}
	tasks, _, err := e.alongEvery(nil, false)
	if err != nil {
		t.Fatal(err)
	}
	if got := tasks[2].Steps; got.Total != 2 || got.Done != 1 {
		t.Errorf("task 2 tallied %+v, want 1 of 2", got)
	}
	p, err := e.taskProgress(2)
	if err != nil {
		t.Fatal(err)
	}
	if p.Steps.Total != 2 || p.Steps.Done != 1 {
		t.Errorf("task 2 progress %+v, want 1 of 2", p.Steps)
	}
}

func TestObjective_newAndListAnswerTheObjectives(t *testing.T) {
	path := storeAt(t)
	made, err := Do(ObjectiveNew, path, given("title", "Ship cairn"))
	if err != nil {
		t.Fatal(err)
	}
	if made.Objective == nil {
		t.Fatal("a new objective answers none")
	}
	listed, err := Do(ObjectiveList, path, given())
	if err != nil {
		t.Fatal(err)
	}
	if len(listed.Objectives) != 1 {
		t.Errorf("listed %d objectives, want 1", len(listed.Objectives))
	}
}

func TestObjective_anEditToABlankTitleSaysWhatWasTyped(t *testing.T) {
	path := storeAt(t)
	if _, err := Do(ObjectiveNew, path, given("title", "Ship cairn")); err != nil {
		t.Fatal(err)
	}
	_, err := Do(ObjectiveEdit, path, given("number", "1", "title", "   "))
	if !errors.Is(err, domain.ErrTitleRequired) || !strings.HasSuffix(err.Error(), `, got "   "`) {
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
	said := RenderText(ObjectiveShow, a)
	for _, want := range []string{"Write", "Draft", "Chapter"} {
		if !strings.Contains(said, want) {
			t.Errorf("%q is not in\n%s", want, said)
		}
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

func TestObjectiveDelete_carryingTasksIsNotNamedByAView(t *testing.T) {
	e := openEngine(t)
	addObjective(t, e, "Ship")
	serving(t, e, "one", 1)
	err := objectiveDeleted(e, map[string]string{"number": "1"})
	if !errors.Is(err, ErrObjectiveCarriesTasks) || errors.Is(err, ErrObjectiveNamedByAView) {
		t.Errorf("refused as %v", err)
	}
}

func TestObjectiveDelete_aSubObjectiveCarryingTasksAsksForThem(t *testing.T) {
	e := openEngine(t)
	addObjective(t, e, "Ship")
	if _, err := e.Ask(ObjectiveNew, given("title", "Write", "parent", "1")); err != nil {
		t.Fatal(err)
	}
	serving(t, e, "one", 2)
	if err := objectiveDeleted(e, map[string]string{"number": "1"}); !errors.Is(err, ErrObjectiveCarriesTasks) {
		t.Errorf("refused as %v", err)
	}
}

func TestObjectiveEdit_aTitleChangedMovesNothing(t *testing.T) {
	e := openEngine(t)
	for _, title := range []string{"a", "b", "c"} {
		addObjective(t, e, title)
	}
	if _, err := e.Ask(ObjectiveEdit, given("number", "1", "title", "A")); err != nil {
		t.Fatal(err)
	}
	listed, err := e.listObjectives()
	if err != nil {
		t.Fatal(err)
	}
	var got []int
	for _, o := range listed {
		got = append(got, o.Carried().Number)
	}
	if !slices.Equal(got, []int{1, 2, 3}) {
		t.Errorf("the objectives stand %v", got)
	}
}

func TestObjectiveDelete_aSubObjectiveWithNoTaskAsksOnlyForItself(t *testing.T) {
	e := openEngine(t)
	addObjective(t, e, "Ship")
	if _, err := e.Ask(ObjectiveNew, given("title", "Write", "parent", "1")); err != nil {
		t.Fatal(err)
	}
	if err := objectiveDeleted(e, map[string]string{"number": "1"}); !errors.Is(err, ErrObjectiveHoldsSubObjectives) || errors.Is(err, ErrObjectiveCarriesTasks) {
		t.Errorf("refused as %v", err)
	}
}

func TestObjectiveEdit_movedUnderAParentAnswersWhatWasAdded(t *testing.T) {
	e := openEngine(t)
	addObjective(t, e, "Ship")
	addObjective(t, e, "Write")
	a, err := e.Ask(ObjectiveEdit, Input{Given: map[string]string{"number": "2", "parent": "1"}, Listed: map[string][]string{"label": {"x"}}})
	if err != nil {
		t.Fatal(err)
	}
	if a.Moved == nil || len(a.Moved.Added.Labels) != 1 {
		t.Errorf("moved %+v", a.Moved)
	}
}
