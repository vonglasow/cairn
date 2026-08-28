package engine

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"cairn/internal/domain"
	"cairn/internal/objective"
	"cairn/internal/task"
)

// addObjective adds an objective, failing the test if it cannot.
func addObjective(t *testing.T, e *Engine, title string) objective.Objective {
	t.Helper()
	o, err := e.AddObjective(title, ObjectiveChanges{})
	if err != nil {
		t.Fatalf("AddObjective(%q): %v", title, err)
	}
	return o
}

func TestEngine_putsATaskUnderAnObjective(t *testing.T) {
	e := openEngine(t)
	o := addObjective(t, e, "Ship cairn 1.0")

	added, err := e.Add("Write the spec", Changes{Objective: serves(factsOfObjective(o).number)})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if got := factsOf(added).objectives; len(got) != 1 || got[0] != factsOfObjective(o).number {
		t.Errorf("Objectives = %v, want just %d", got, factsOfObjective(o).number)
	}
}

func TestEngine_refusesAnObjectiveThatDoesNotExist(t *testing.T) {
	e := openEngine(t)
	if _, err := e.Add("Write the spec", Changes{Objective: serves(99)}); !errors.Is(err, ErrNoSuchObjective) {
		t.Errorf("err = %v, want ErrNoSuchObjective", err)
	}
}

func TestEngine_setsATaskLooseFromItsObjective(t *testing.T) {
	e := openEngine(t)
	o := addObjective(t, e, "Ship cairn 1.0")
	added, err := e.Add("Write the spec", Changes{Objective: serves(factsOfObjective(o).number)})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	loose, err := e.Edit(factsOf(added).number, Changes{Objective: serves()})
	if err != nil {
		t.Fatalf("Edit: %v", err)
	}
	if got := factsOf(loose).objectives; len(got) != 0 {
		t.Errorf("Objectives = %v, want none", got)
	}
}

// The three refusals below are one rule seen from three sides: a task never
// carries a parent and an objective at the same time.

func TestEngine_refusesAnObjectiveOnASubTask(t *testing.T) {
	e := openEngine(t)
	o := addObjective(t, e, "Ship cairn 1.0")
	parent := add(t, e, "Ship it")
	child, err := e.Add("Write the spec", Changes{Parent: of(factsOf(parent).number)})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	_, err = e.Edit(factsOf(child).number, Changes{Objective: serves(factsOfObjective(o).number)})
	if !errors.Is(err, ErrSubTaskServesNoObjective) {
		t.Errorf("err = %v, want ErrSubTaskServesNoObjective", err)
	}
}

func TestEngine_refusesAParentAndAnObjectiveAtOnce(t *testing.T) {
	e := openEngine(t)
	o := addObjective(t, e, "Ship cairn 1.0")
	parent := add(t, e, "Ship it")

	_, err := e.Add("Write the spec", Changes{
		Parent:    of(factsOf(parent).number),
		Objective: serves(factsOfObjective(o).number),
	})
	if !errors.Is(err, ErrSubTaskServesNoObjective) {
		t.Errorf("err = %v, want ErrSubTaskServesNoObjective", err)
	}
}

func TestEngine_refusesAParentOnATaskServingAnObjective(t *testing.T) {
	e := openEngine(t)
	o := addObjective(t, e, "Ship cairn 1.0")
	parent := add(t, e, "Ship it")
	servingTask, err := e.Add("Write the spec", Changes{Objective: serves(factsOfObjective(o).number)})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if _, err := e.Edit(factsOf(servingTask).number, Changes{Parent: of(factsOf(parent).number)}); !errors.Is(err, ErrSubTaskServesNoObjective) {
		t.Errorf("err = %v, want ErrSubTaskServesNoObjective", err)
	}
	// And nothing moved.
	still, err := e.Show(factsOf(servingTask).number)
	if err != nil {
		t.Fatalf("Show: %v", err)
	}
	stillFacts := factsOf(still)
	if len(stillFacts.objectives) != 1 || stillFacts.objectives[0] != factsOfObjective(o).number ||
		stillFacts.parent != task.NoParent {
		t.Errorf("%+v changed despite the refusal", stillFacts)
	}
}

// The refusal is not a dead end: set the task loose and hang it up in one go.
func TestEngine_lettingGoOfTheObjectiveMakesRoomForAParent(t *testing.T) {
	e := openEngine(t)
	o := addObjective(t, e, "Ship cairn 1.0")
	parent := add(t, e, "Ship it")
	servingTask, err := e.Add("Write the spec", Changes{Objective: serves(factsOfObjective(o).number)})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	moved, err := e.Edit(factsOf(servingTask).number, Changes{
		Objective: serves(),
		Parent:    of(factsOf(parent).number),
	})
	if err != nil {
		t.Fatalf("Edit: %v", err)
	}
	movedFacts := factsOf(moved)
	if movedFacts.parent != factsOf(parent).number || len(movedFacts.objectives) != 0 {
		t.Errorf("%+v, want it hanging under %d and serving nothing", movedFacts, factsOf(parent).number)
	}
}

// A sub-task that never had an objective is still a sub-task: hanging it
// somewhere else must not be mistaken for the refusal above.
func TestEngine_stillMovesASubTaskThatServesNothing(t *testing.T) {
	e := openEngine(t)
	first := add(t, e, "Ship it")
	second := add(t, e, "Ship it again")
	child, err := e.Add("Write the spec", Changes{Parent: of(factsOf(first).number)})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	moved, err := e.Edit(factsOf(child).number, Changes{Parent: of(factsOf(second).number)})
	if err != nil {
		t.Fatalf("Edit: %v", err)
	}
	if factsOf(moved).parent != factsOf(second).number {
		t.Errorf("Parent = %d, want %d", factsOf(moved).parent, factsOf(second).number)
	}
}

// done marks a task as done, failing the test if it cannot.
func done(t *testing.T, e *Engine, number int) {
	t.Helper()
	if _, err := e.Edit(number, Changes{State: of("done")}); err != nil {
		t.Fatalf("Edit(%d): %v", number, err)
	}
}

// serving adds a task under an objective, failing the test if it cannot.
func serving(t *testing.T, e *Engine, title string, objectiveNumber int) task.Task {
	t.Helper()
	added, err := e.Add(title, Changes{Objective: serves(objectiveNumber)})
	if err != nil {
		t.Fatalf("Add(%q): %v", title, err)
	}
	return added
}

// under adds a sub-task, failing the test if it cannot.
func under(t *testing.T, e *Engine, title string, parent int) task.Task {
	t.Helper()
	added, err := e.Add(title, Changes{Parent: of(parent)})
	if err != nil {
		t.Fatalf("Add(%q): %v", title, err)
	}
	return added
}

func TestEngine_countsTheTasksUnderAnObjective(t *testing.T) {
	e := openEngine(t)
	o := addObjective(t, e, "Ship cairn 1.0")
	first := serving(t, e, "Write the spec", factsOfObjective(o).number)
	serving(t, e, "Write the tests", factsOfObjective(o).number)
	done(t, e, factsOf(first).number)

	got, err := e.ObjectiveProgress(factsOfObjective(o).number)
	if err != nil {
		t.Fatalf("ObjectiveProgress: %v", err)
	}
	if got.Tasks != (domain.Tally{Total: 2, Done: 1}) {
		t.Errorf("Tasks = %+v, want 1 of 2", got.Tasks)
	}
	if got.SubTasks != (domain.Tally{}) {
		t.Errorf("SubTasks = %+v, want none", got.SubTasks)
	}
	if got.Percent() != 50 {
		t.Errorf("Percent() = %d, want 50", got.Percent())
	}
}

// A sub-task counts towards the objective its parent serves, and is counted
// apart from it: one task and one sub-task is never two tasks.
func TestEngine_countsSubTasksApartUnderAnObjective(t *testing.T) {
	e := openEngine(t)
	o := addObjective(t, e, "Ship cairn 1.0")
	parent := serving(t, e, "Write the spec", factsOfObjective(o).number)
	child := under(t, e, "Write the tests", factsOf(parent).number)
	done(t, e, factsOf(child).number)

	got, err := e.ObjectiveProgress(factsOfObjective(o).number)
	if err != nil {
		t.Fatalf("ObjectiveProgress: %v", err)
	}
	if got.Tasks != (domain.Tally{Total: 1}) {
		t.Errorf("Tasks = %+v, want 0 of 1", got.Tasks)
	}
	if got.SubTasks != (domain.Tally{Total: 1, Done: 1}) {
		t.Errorf("SubTasks = %+v, want 1 of 1", got.SubTasks)
	}
	if got.Percent() != 50 {
		t.Errorf("Percent() = %d, want 50: the sub-task is half the work", got.Percent())
	}
}

// Sub-tasks of a task serving another objective must not be swept in.
func TestEngine_countsOnlyWhatServesTheObjectiveAsked(t *testing.T) {
	e := openEngine(t)
	o := addObjective(t, e, "Ship cairn 1.0")
	other := addObjective(t, e, "Learn Go properly")
	serving(t, e, "Write the spec", factsOfObjective(o).number)
	elsewhere := serving(t, e, "Read the book", factsOfObjective(other).number)
	under(t, e, "Take notes", factsOf(elsewhere).number)
	loose := add(t, e, "Nothing to do with it")
	under(t, e, "Nor this", factsOf(loose).number)

	got, err := e.ObjectiveProgress(factsOfObjective(o).number)
	if err != nil {
		t.Fatalf("ObjectiveProgress: %v", err)
	}
	if got.Total() != 1 {
		t.Errorf("%+v counts %d, want only the one task under it", got, got.Total())
	}
}

func TestEngine_countsNothingUnderAnEmptyObjective(t *testing.T) {
	e := openEngine(t)
	o := addObjective(t, e, "Ship cairn 1.0")
	add(t, e, "Nothing to do with it")

	got, err := e.ObjectiveProgress(factsOfObjective(o).number)
	if err != nil {
		t.Fatalf("ObjectiveProgress: %v", err)
	}
	if got.Any() {
		t.Errorf("Any() = true for %+v, want false", got)
	}
}

func TestEngine_saysWhenTheObjectiveAskedAboutDoesNotExist(t *testing.T) {
	if _, err := openEngine(t).ObjectiveProgress(99); !errors.Is(err, ErrNoSuchObjective) {
		t.Errorf("err = %v, want ErrNoSuchObjective", err)
	}
}

func TestEngine_countsTheSubTasksOfATask(t *testing.T) {
	e := openEngine(t)
	parent := add(t, e, "Ship it")
	first := under(t, e, "Write the spec", factsOf(parent).number)
	under(t, e, "Write the tests", factsOf(parent).number)
	done(t, e, factsOf(first).number)

	got, err := e.TaskProgress(factsOf(parent).number)
	if err != nil {
		t.Fatalf("TaskProgress: %v", err)
	}
	if got.SubTasks != (domain.Tally{Total: 2, Done: 1}) {
		t.Errorf("SubTasks = %+v, want 1 of 2", got.SubTasks)
	}
	// A task never counts itself: two sub-tasks is one of two, not one of three.
	if got.Tasks != (domain.Tally{}) {
		t.Errorf("Tasks = %+v, want none: a task does not count itself", got.Tasks)
	}
}

func TestEngine_countsNothingUnderATaskWithNoSubTasks(t *testing.T) {
	e := openEngine(t)
	alone := add(t, e, "Ship it")
	if got, err := e.TaskProgress(factsOf(alone).number); err != nil {
		t.Fatalf("TaskProgress: %v", err)
	} else if got.Any() {
		t.Errorf("Any() = true for %+v, want false", got)
	}
}

func TestObjectiveList_listsEveryObjective(t *testing.T) {
	path := storeAt(t)
	if _, err := Do(ObjectiveAdd, path, Input{Given: map[string]string{"title": "one"}}); err != nil {
		t.Fatalf("Do(ObjectiveAdd): %v", err)
	}
	if _, err := Do(ObjectiveAdd, path, Input{Given: map[string]string{"title": "two"}}); err != nil {
		t.Fatalf("Do(ObjectiveAdd): %v", err)
	}

	answer, err := Do(ObjectiveList, path, Input{})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if len(answer.Objectives) != 2 {
		t.Fatalf("len(answer.Objectives) = %d, want 2", len(answer.Objectives))
	}
}

func TestObjectiveSearch_listsWhatCarriesEveryNamedThing(t *testing.T) {
	path := storeAt(t)
	if _, err := Do(ObjectiveAdd, path, Input{Given: map[string]string{"title": "ship the release"}}); err != nil {
		t.Fatalf("Do(ObjectiveAdd): %v", err)
	}
	if _, err := Do(ObjectiveAdd, path, Input{Given: map[string]string{"title": "plan the offsite"}}); err != nil {
		t.Fatalf("Do(ObjectiveAdd): %v", err)
	}

	answer, err := Do(ObjectiveSearch, path, Input{Listed: map[string][]string{"named": {"release"}}})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if len(answer.MatchedObjectives) != 1 || factsOfObjective(answer.MatchedObjectives[0]).title != "ship the release" {
		t.Fatalf("answer.MatchedObjectives = %v, want one objective titled %q", answer.MatchedObjectives, "ship the release")
	}
}

func TestObjectiveShow_showsAnObjectiveItsTasksAndProgress(t *testing.T) {
	path := storeAt(t)
	added, err := Do(ObjectiveAdd, path, Input{Given: map[string]string{"title": "ship the release"}})
	if err != nil {
		t.Fatalf("Do(ObjectiveAdd): %v", err)
	}
	number := strconv.Itoa(factsOfObjective(*added.Objective).number)
	if _, err := Do(TaskAdd, path, Input{Given: map[string]string{"title": "task"}, Listed: map[string][]string{"objective": {number}}}); err != nil {
		t.Fatalf("Do(TaskAdd): %v", err)
	}

	answer, err := Do(ObjectiveShow, path, Input{Given: map[string]string{"number": number}})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if answer.Objective == nil || factsOfObjective(*answer.Objective).title != "ship the release" {
		t.Fatalf("answer.Objective = %v, want the objective just added", answer.Objective)
	}
	if len(answer.Under) != 1 || factsOf(answer.Under[0]).title != "task" {
		t.Fatalf("answer.Under = %v, want one task titled %q", answer.Under, "task")
	}
	if answer.Progress.Tasks.Total != 1 {
		t.Fatalf("answer.Progress.Tasks.Total = %d, want 1", answer.Progress.Tasks.Total)
	}
}

func TestObjectiveEdit_editsAnObjectiveAndRereads(t *testing.T) {
	path := storeAt(t)
	added, err := Do(ObjectiveAdd, path, Input{Given: map[string]string{"title": "ship the release"}})
	if err != nil {
		t.Fatalf("Do(ObjectiveAdd): %v", err)
	}
	number := strconv.Itoa(factsOfObjective(*added.Objective).number)
	if _, err := Do(TaskAdd, path, Input{Given: map[string]string{"title": "task"}, Listed: map[string][]string{"objective": {number}}}); err != nil {
		t.Fatalf("Do(TaskAdd): %v", err)
	}

	answer, err := Do(ObjectiveEdit, path, Input{Given: map[string]string{"number": number, "desc": "get it out the door"}})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if answer.Objective == nil || factsOfObjective(*answer.Objective).description != "get it out the door" {
		t.Fatalf("answer.Objective = %v, want the description changed", answer.Objective)
	}
	if len(answer.Under) != 1 {
		t.Fatalf("answer.Under = %v, want the objective re-read the way objective_show does", answer.Under)
	}

	if _, err := Do(ObjectiveEdit, path, Input{Given: map[string]string{"number": number}}); err == nil ||
		!strings.Contains(err.Error(), "--title") {
		t.Fatalf("err = %v, want a refusal naming the flags to change", err)
	}
}

func TestObjectiveDelete_deletesAnObjective(t *testing.T) {
	path := storeAt(t)
	added, err := Do(ObjectiveAdd, path, Input{Given: map[string]string{"title": "gone soon"}})
	if err != nil {
		t.Fatalf("Do(ObjectiveAdd): %v", err)
	}

	answer, err := Do(ObjectiveDelete, path, Input{Given: map[string]string{"number": strconv.Itoa(factsOfObjective(*added.Objective).number)}})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if answer.Deleted == nil || *answer.Deleted != factsOfObjective(*added.Objective).number {
		t.Fatalf("answer.Deleted = %v, want %d", answer.Deleted, factsOfObjective(*added.Objective).number)
	}

	list, err := Do(ObjectiveList, path, Input{})
	if err != nil {
		t.Fatalf("Do(ObjectiveList): %v", err)
	}
	if len(list.Objectives) != 0 {
		t.Fatalf("len(list.Objectives) = %d, want 0: the objective was not deleted", len(list.Objectives))
	}
}

func TestObjectiveAdd_addsAnObjective(t *testing.T) {
	path := storeAt(t)
	in := Input{Given: map[string]string{"title": "ship the release"}}

	answer, err := Do(ObjectiveAdd, path, in)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if answer.Objective == nil || factsOfObjective(*answer.Objective).title != "ship the release" {
		t.Fatalf("answer.Objective = %v, want one titled %q", answer.Objective, "ship the release")
	}
}

// serves is Changes.Objective for a single objective, which is what most of
// these tests name. A task serves a set, so even one is given as one.
func serves(numbers ...int) *[]string {
	named := make([]string, 0, len(numbers))
	for _, one := range numbers {
		named = append(named, fmt.Sprint(one))
	}
	return &named
}
