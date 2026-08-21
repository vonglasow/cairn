package engine

import (
	"errors"
	"testing"

	"cairn/internal/domain"
)

func addObjective(t *testing.T, e *Engine, title string) domain.Objective {
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

	task, err := e.Add("Write the spec", Changes{Objective: of(o.Number)})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if task.Objective != o.Number {
		t.Errorf("Objective = %d, want %d", task.Objective, o.Number)
	}
}

func TestEngine_refusesAnObjectiveThatDoesNotExist(t *testing.T) {
	e := openEngine(t)
	if _, err := e.Add("Write the spec", Changes{Objective: of(99)}); !errors.Is(err, ErrNoSuchObjective) {
		t.Errorf("err = %v, want ErrNoSuchObjective", err)
	}
}

func TestEngine_setsATaskLooseFromItsObjective(t *testing.T) {
	e := openEngine(t)
	o := addObjective(t, e, "Ship cairn 1.0")
	task, err := e.Add("Write the spec", Changes{Objective: of(o.Number)})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	loose, err := e.Edit(task.Number, Changes{Objective: of(domain.NoObjective)})
	if err != nil {
		t.Fatalf("Edit: %v", err)
	}
	if loose.Objective != domain.NoObjective {
		t.Errorf("Objective = %d, want none", loose.Objective)
	}
}

// The three refusals below are one rule seen from three sides: a task never
// carries a parent and an objective at the same time.

func TestEngine_refusesAnObjectiveOnASubTask(t *testing.T) {
	e := openEngine(t)
	o := addObjective(t, e, "Ship cairn 1.0")
	parent := add(t, e, "Ship it")
	child, err := e.Add("Write the spec", Changes{Parent: of(parent.Number)})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	_, err = e.Edit(child.Number, Changes{Objective: of(o.Number)})
	if !errors.Is(err, ErrSubTaskServesNoObjective) {
		t.Errorf("err = %v, want ErrSubTaskServesNoObjective", err)
	}
}

func TestEngine_refusesAParentAndAnObjectiveAtOnce(t *testing.T) {
	e := openEngine(t)
	o := addObjective(t, e, "Ship cairn 1.0")
	parent := add(t, e, "Ship it")

	_, err := e.Add("Write the spec", Changes{
		Parent:    of(parent.Number),
		Objective: of(o.Number),
	})
	if !errors.Is(err, ErrSubTaskServesNoObjective) {
		t.Errorf("err = %v, want ErrSubTaskServesNoObjective", err)
	}
}

func TestEngine_refusesAParentOnATaskServingAnObjective(t *testing.T) {
	e := openEngine(t)
	o := addObjective(t, e, "Ship cairn 1.0")
	parent := add(t, e, "Ship it")
	serving, err := e.Add("Write the spec", Changes{Objective: of(o.Number)})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if _, err := e.Edit(serving.Number, Changes{Parent: of(parent.Number)}); !errors.Is(err, ErrSubTaskServesNoObjective) {
		t.Errorf("err = %v, want ErrSubTaskServesNoObjective", err)
	}
	// And nothing moved.
	still, err := e.Show(serving.Number)
	if err != nil {
		t.Fatalf("Show: %v", err)
	}
	if still.Objective != o.Number || still.Parent != domain.NoParent {
		t.Errorf("%+v changed despite the refusal", still)
	}
}

// The refusal is not a dead end: set the task loose and hang it up in one go.
func TestEngine_lettingGoOfTheObjectiveMakesRoomForAParent(t *testing.T) {
	e := openEngine(t)
	o := addObjective(t, e, "Ship cairn 1.0")
	parent := add(t, e, "Ship it")
	serving, err := e.Add("Write the spec", Changes{Objective: of(o.Number)})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	moved, err := e.Edit(serving.Number, Changes{
		Objective: of(domain.NoObjective),
		Parent:    of(parent.Number),
	})
	if err != nil {
		t.Fatalf("Edit: %v", err)
	}
	if moved.Parent != parent.Number || moved.Objective != domain.NoObjective {
		t.Errorf("%+v, want it hanging under %d and serving nothing", moved, parent.Number)
	}
}

// A sub-task that never had an objective is still a sub-task: hanging it
// somewhere else must not be mistaken for the refusal above.
func TestEngine_stillMovesASubTaskThatServesNothing(t *testing.T) {
	e := openEngine(t)
	first := add(t, e, "Ship it")
	second := add(t, e, "Ship it again")
	child, err := e.Add("Write the spec", Changes{Parent: of(first.Number)})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	moved, err := e.Edit(child.Number, Changes{Parent: of(second.Number)})
	if err != nil {
		t.Fatalf("Edit: %v", err)
	}
	if moved.Parent != second.Number {
		t.Errorf("Parent = %d, want %d", moved.Parent, second.Number)
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
func serving(t *testing.T, e *Engine, title string, objective int) domain.Task {
	t.Helper()
	task, err := e.Add(title, Changes{Objective: of(objective)})
	if err != nil {
		t.Fatalf("Add(%q): %v", title, err)
	}
	return task
}

// under adds a sub-task, failing the test if it cannot.
func under(t *testing.T, e *Engine, title string, parent int) domain.Task {
	t.Helper()
	task, err := e.Add(title, Changes{Parent: of(parent)})
	if err != nil {
		t.Fatalf("Add(%q): %v", title, err)
	}
	return task
}

func TestEngine_countsTheTasksUnderAnObjective(t *testing.T) {
	e := openEngine(t)
	o := addObjective(t, e, "Ship cairn 1.0")
	first := serving(t, e, "Write the spec", o.Number)
	serving(t, e, "Write the tests", o.Number)
	done(t, e, first.Number)

	got, err := e.ObjectiveProgress(o.Number)
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
	parent := serving(t, e, "Write the spec", o.Number)
	child := under(t, e, "Write the tests", parent.Number)
	done(t, e, child.Number)

	got, err := e.ObjectiveProgress(o.Number)
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
	serving(t, e, "Write the spec", o.Number)
	elsewhere := serving(t, e, "Read the book", other.Number)
	under(t, e, "Take notes", elsewhere.Number)
	loose := add(t, e, "Nothing to do with it")
	under(t, e, "Nor this", loose.Number)

	got, err := e.ObjectiveProgress(o.Number)
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

	got, err := e.ObjectiveProgress(o.Number)
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
	first := under(t, e, "Write the spec", parent.Number)
	under(t, e, "Write the tests", parent.Number)
	done(t, e, first.Number)

	got, err := e.TaskProgress(parent.Number)
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
	if got, err := e.TaskProgress(alone.Number); err != nil {
		t.Fatalf("TaskProgress: %v", err)
	} else if got.Any() {
		t.Errorf("Any() = true for %+v, want false", got)
	}
}
