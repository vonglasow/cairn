package store

import (
	"errors"
	"testing"
	"time"

	"cairn/internal/domain"
	"cairn/internal/objective"
	"cairn/internal/task"
)

// keepObjective adds an objective titled title, failing the test if it cannot.
func keepObjective(t *testing.T, s *Store, title string) objective.Objective {
	t.Helper()
	o, err := objective.NewObjective(title)
	if err != nil {
		t.Fatalf("NewObjective: %v", err)
	}
	kept, err := s.KeepObjective(o)
	if err != nil {
		t.Fatalf("KeepObjective: %v", err)
	}
	return kept
}

func TestKeepObjective_numbersFromOne(t *testing.T) {
	s := openStore(t)
	if got := factsOfObjective(keepObjective(t, s, "Ship cairn 1.0")).number; got != 1 {
		t.Errorf("Number = %d, want 1", got)
	}
	if got := factsOfObjective(keepObjective(t, s, "Learn Go properly")).number; got != 2 {
		t.Errorf("Number = %d, want 2", got)
	}
}

// Objectives and tasks are counted apart, so the first of each is number one.
// One sequence for both would make `objective show 1` and `task show 1` two
// halves of the same numbering, which is not what either command says.
func TestKeepObjective_numbersApartFromTasks(t *testing.T) {
	s := openStore(t)
	keep(t, s, "Write the spec")
	keep(t, s, "Ship it")
	if got := factsOfObjective(keepObjective(t, s, "Ship cairn 1.0")).number; got != 1 {
		t.Errorf("Number = %d, want 1: objectives are numbered apart", got)
	}
}

func TestFindObjective_handsBackEverythingItWasGiven(t *testing.T) {
	s := openStore(t)
	built, err := objective.NewObjective("Ship cairn 1.0")
	if err != nil {
		t.Fatalf("NewObjective: %v", err)
	}
	built = changedObjective(built, func(a *objective.AsTold) {
		a.Described("the one that unblocks the rest")
		a.Standing(objective.OnHold)
		a.Carries(domain.High)
		a.Labelled("perso")
		a.Labelled("spec")
		a.LabelsTold()
		a.Due(time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC))
		a.Estimated(90 * time.Minute)
	})
	kept, err := s.KeepObjective(built)
	if err != nil {
		t.Fatalf("KeepObjective: %v", err)
	}
	got, err := s.FindObjective(factsOfObjective(kept).number)
	if err != nil {
		t.Fatalf("FindObjective: %v", err)
	}
	gotFacts, wantFacts := factsOfObjective(got), factsOfObjective(kept)
	if gotFacts.title != wantFacts.title || gotFacts.description != wantFacts.description {
		t.Errorf("Title/Description = %q/%q, want %q/%q",
			gotFacts.title, gotFacts.description, wantFacts.title, wantFacts.description)
	}
	if gotFacts.state != wantFacts.state || gotFacts.priority != wantFacts.priority {
		t.Errorf("State/Priority = %q/%q, want %q/%q",
			gotFacts.state, gotFacts.priority, wantFacts.state, wantFacts.priority)
	}
	if !gotFacts.deadline.Equal(wantFacts.deadline) {
		t.Errorf("Deadline = %v, want %v", gotFacts.deadline, wantFacts.deadline)
	}
	if gotFacts.estimate != wantFacts.estimate {
		t.Errorf("Estimate = %v, want %v", gotFacts.estimate, wantFacts.estimate)
	}
	if len(gotFacts.labels) != 2 || gotFacts.labels[0] != "perso" || gotFacts.labels[1] != "spec" {
		t.Errorf("Labels = %v, want [perso spec]", gotFacts.labels)
	}
}

// The labels of an objective live in a table of their own, so a task and an
// objective sharing a number must not share labels.
func TestFindObjective_doesNotPickUpATasksLabels(t *testing.T) {
	s := openStore(t)
	built, err := task.NewTask("Write the spec")
	if err != nil {
		t.Fatalf("NewTask: %v", err)
	}
	built = changed(built, func(a *task.AsTold) {
		a.Labelled("spec")
		a.LabelsTold()
	})
	if _, err := s.Keep(built); err != nil {
		t.Fatalf("Keep: %v", err)
	}
	got := keepObjective(t, s, "Ship cairn 1.0")
	found, err := s.FindObjective(factsOfObjective(got).number)
	if err != nil {
		t.Fatalf("FindObjective: %v", err)
	}
	if labels := factsOfObjective(found).labels; len(labels) != 0 {
		t.Errorf("Labels = %v, want none: task 1 and objective 1 are not the same thing", labels)
	}
}

func TestFindObjective_saysWhenThereIsNone(t *testing.T) {
	if _, err := openStore(t).FindObjective(99); !errors.Is(err, ErrNoSuchObjective) {
		t.Errorf("err = %v, want ErrNoSuchObjective", err)
	}
}

func TestAllObjectives_handsThemBackInTheOrderTheyWereAdded(t *testing.T) {
	s := openStore(t)
	keepObjective(t, s, "Ship cairn 1.0")
	keepObjective(t, s, "Learn Go properly")

	all, err := s.AllObjectives()
	if err != nil {
		t.Fatalf("AllObjectives: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("%d objectives, want 2", len(all))
	}
	if factsOfObjective(all[0]).title != "Ship cairn 1.0" || factsOfObjective(all[1]).title != "Learn Go properly" {
		t.Errorf("got %q then %q, want them in the order they were added",
			factsOfObjective(all[0]).title, factsOfObjective(all[1]).title)
	}
}

func TestAllObjectives_handsBackNothingWhenThereIsNone(t *testing.T) {
	all, err := openStore(t).AllObjectives()
	if err != nil {
		t.Fatalf("AllObjectives: %v", err)
	}
	if len(all) != 0 {
		t.Errorf("%d objectives, want none", len(all))
	}
}

func TestUpdateObjective_replacesWhatIsHeld(t *testing.T) {
	s := openStore(t)
	o := keepObjective(t, s, "Ship cairn 1.0")
	o = changedObjective(o, func(a *objective.AsTold) {
		a.Titled("Ship cairn 1.0, for real")
		a.Standing(objective.Reached)
		a.NotLabelled()
		a.Labelled("done")
		a.LabelsTold()
	})
	if err := s.UpdateObjective(o); err != nil {
		t.Fatalf("UpdateObjective: %v", err)
	}
	got, err := s.FindObjective(factsOfObjective(o).number)
	if err != nil {
		t.Fatalf("FindObjective: %v", err)
	}
	gotFacts, wantFacts := factsOfObjective(got), factsOfObjective(o)
	if gotFacts.title != wantFacts.title || gotFacts.state != objective.Reached {
		t.Errorf("got %q/%q, want %q/%q", gotFacts.title, gotFacts.state, wantFacts.title, objective.Reached)
	}
	if len(gotFacts.labels) != 1 || gotFacts.labels[0] != "done" {
		t.Errorf("Labels = %v, want [done]", gotFacts.labels)
	}
}

// An objective carries a set of labels, so one left out of an update is one
// taken away — the same rule a task follows.
func TestUpdateObjective_replacesTheLabelsRatherThanAddingToThem(t *testing.T) {
	s := openStore(t)
	o := keepObjective(t, s, "Ship cairn 1.0")
	o = changedObjective(o, func(a *objective.AsTold) {
		a.NotLabelled()
		a.Labelled("perso")
		a.Labelled("spec")
		a.LabelsTold()
	})
	if err := s.UpdateObjective(o); err != nil {
		t.Fatalf("UpdateObjective: %v", err)
	}
	o = changedObjective(o, func(a *objective.AsTold) {
		a.NotLabelled()
		a.Labelled("spec")
		a.LabelsTold()
	})
	if err := s.UpdateObjective(o); err != nil {
		t.Fatalf("UpdateObjective: %v", err)
	}
	got, err := s.FindObjective(factsOfObjective(o).number)
	if err != nil {
		t.Fatalf("FindObjective: %v", err)
	}
	if labels := factsOfObjective(got).labels; len(labels) != 1 || labels[0] != "spec" {
		t.Errorf("Labels = %v, want [spec]", labels)
	}
}

func TestUpdateObjective_saysWhenThereIsNone(t *testing.T) {
	built, err := objective.NewObjective("Nowhere")
	if err != nil {
		t.Fatalf("NewObjective: %v", err)
	}
	built = changedObjective(built, func(a *objective.AsTold) { a.Numbered(99) })
	if err := openStore(t).UpdateObjective(built); !errors.Is(err, ErrNoSuchObjective) {
		t.Errorf("err = %v, want ErrNoSuchObjective", err)
	}
}

func TestDeleteObjective_removesIt(t *testing.T) {
	s := openStore(t)
	o := keepObjective(t, s, "Ship cairn 1.0")
	if err := s.DeleteObjective(factsOfObjective(o).number); err != nil {
		t.Fatalf("DeleteObjective: %v", err)
	}
	if _, err := s.FindObjective(factsOfObjective(o).number); !errors.Is(err, ErrNoSuchObjective) {
		t.Errorf("err = %v, want ErrNoSuchObjective after a delete", err)
	}
}

func TestDeleteObjective_takesItsLabelsWithIt(t *testing.T) {
	s := openStore(t)
	o := keepObjective(t, s, "Ship cairn 1.0")
	o = changedObjective(o, func(a *objective.AsTold) {
		a.NotLabelled()
		a.Labelled("perso")
		a.LabelsTold()
	})
	if err := s.UpdateObjective(o); err != nil {
		t.Fatalf("UpdateObjective: %v", err)
	}
	if err := s.DeleteObjective(factsOfObjective(o).number); err != nil {
		t.Fatalf("DeleteObjective: %v", err)
	}
	left, err := s.labelsOf(objectiveLabels, factsOfObjective(o).number)
	if err != nil {
		t.Fatalf("labelsOf: %v", err)
	}
	if len(left) != 0 {
		t.Errorf("%v left behind: a deleted objective takes its labels with it", left)
	}
}

func TestDeleteObjective_saysWhenThereIsNone(t *testing.T) {
	if err := openStore(t).DeleteObjective(99); !errors.Is(err, ErrNoSuchObjective) {
		t.Errorf("err = %v, want ErrNoSuchObjective", err)
	}
}

func TestKeep_holdsTheObjectiveATaskServes(t *testing.T) {
	s := openStore(t)
	o := keepObjective(t, s, "Ship cairn 1.0")

	built, err := task.NewTask("Write the spec")
	if err != nil {
		t.Fatalf("NewTask: %v", err)
	}
	built = changed(built, func(a *task.AsTold) { a.Serves(factsOfObjective(o).number) })
	kept, err := s.Keep(built)
	if err != nil {
		t.Fatalf("Keep: %v", err)
	}
	found, err := s.Find(factsOf(kept).number)
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	if facts := factsOf(found); facts.objective != factsOfObjective(o).number {
		t.Errorf("Objective = %d, want %d", facts.objective, factsOfObjective(o).number)
	}
}

func TestUpdate_changesTheObjectiveATaskServes(t *testing.T) {
	s := openStore(t)
	o := keepObjective(t, s, "Ship cairn 1.0")
	built := keep(t, s, "Write the spec")

	built = changed(built, func(a *task.AsTold) { a.Serves(factsOfObjective(o).number) })
	if err := s.Update(built); err != nil {
		t.Fatalf("Update: %v", err)
	}
	found, err := s.Find(factsOf(built).number)
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	if facts := factsOf(found); facts.objective != factsOfObjective(o).number {
		t.Errorf("Objective = %d, want %d", facts.objective, factsOfObjective(o).number)
	}

	found = changed(found, func(a *task.AsTold) { a.ServesNothing() })
	if err := s.Update(found); err != nil {
		t.Fatalf("Update: %v", err)
	}
	loose, err := s.Find(factsOf(built).number)
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	if facts := factsOf(loose); facts.objective != task.NoObjective {
		t.Errorf("Objective = %d, want none", facts.objective)
	}
}

// All and Under read the same columns as Find, and the three lists are written
// out separately. A column added to one and forgotten in the others comes back
// as a zero nobody notices, so each is asked the same question.
func TestAllAndUnder_readTheObjectiveTaskstoo(t *testing.T) {
	s := openStore(t)
	o := keepObjective(t, s, "Ship cairn 1.0")
	parent := keep(t, s, "Ship it")
	parent = changed(parent, func(a *task.AsTold) { a.Serves(factsOfObjective(o).number) })
	if err := s.Update(parent); err != nil {
		t.Fatalf("Update: %v", err)
	}
	child := keep(t, s, "Write the spec")
	child = changed(child, func(a *task.AsTold) { a.Under(factsOf(parent).number) })
	if err := s.Update(child); err != nil {
		t.Fatalf("Update: %v", err)
	}

	all, err := s.All()
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if facts := factsOf(all[0]); facts.objective != factsOfObjective(o).number {
		t.Errorf("All: Objective = %d, want %d", facts.objective, factsOfObjective(o).number)
	}
	under, err := s.Under(factsOf(parent).number)
	if err != nil {
		t.Fatalf("Under: %v", err)
	}
	if len(under) != 1 || factsOf(under[0]).objective != task.NoObjective {
		t.Errorf("Under: %+v, want one sub-task serving no objective", under)
	}
}

func TestTasksServing_handsBackWhatWasPutUnderTheObjective(t *testing.T) {
	s := openStore(t)
	o := keepObjective(t, s, "Ship cairn 1.0")
	other := keepObjective(t, s, "Learn Go properly")

	for _, title := range []string{"Write the spec", "Ship it"} {
		built := keep(t, s, title)
		built = changed(built, func(a *task.AsTold) { a.Serves(factsOfObjective(o).number) })
		if err := s.Update(built); err != nil {
			t.Fatalf("Update: %v", err)
		}
	}
	elsewhere := keep(t, s, "Read the book")
	elsewhere = changed(elsewhere, func(a *task.AsTold) { a.Serves(factsOfObjective(other).number) })
	if err := s.Update(elsewhere); err != nil {
		t.Fatalf("Update: %v", err)
	}
	keep(t, s, "Nothing to do with it")

	serving, err := s.TasksServing(factsOfObjective(o).number)
	if err != nil {
		t.Fatalf("TasksServing: %v", err)
	}
	if len(serving) != 2 {
		t.Fatalf("%d tasks serving, want 2", len(serving))
	}
	if factsOf(serving[0]).title != "Write the spec" || factsOf(serving[1]).title != "Ship it" {
		t.Errorf("got %q then %q, want them in the order they were added",
			factsOf(serving[0]).title, factsOf(serving[1]).title)
	}
}

func TestTasksServing_handsBackNothingWhenNothingWasPutUnderIt(t *testing.T) {
	s := openStore(t)
	o := keepObjective(t, s, "Ship cairn 1.0")
	keep(t, s, "Nothing to do with it")

	serving, err := s.TasksServing(factsOfObjective(o).number)
	if err != nil {
		t.Fatalf("TasksServing: %v", err)
	}
	if len(serving) != 0 {
		t.Errorf("%d tasks serving, want none", len(serving))
	}
}

// Zero means "no objective", and it is the number no objective ever carries.
// Asking for it must not hand back every loose task in the store.
func TestTasksServing_doesNotTreatNoObjectiveAsAnObjective(t *testing.T) {
	s := openStore(t)
	keep(t, s, "Write the spec")
	keep(t, s, "Ship it")

	serving, err := s.TasksServing(task.NoObjective)
	if err != nil {
		t.Fatalf("TasksServing: %v", err)
	}
	if len(serving) != 0 {
		t.Errorf("%d tasks serving nothing, want none", len(serving))
	}
}
