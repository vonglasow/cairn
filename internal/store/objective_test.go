package store

import (
	"errors"
	"testing"
	"time"

	"cairn/internal/domain"
)

// keepObjective adds an objective titled title, failing the test if it cannot.
func keepObjective(t *testing.T, s *Store, title string) domain.Objective {
	t.Helper()
	o, err := domain.NewObjective(title)
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
	if got := keepObjective(t, s, "Ship cairn 1.0").Number; got != 1 {
		t.Errorf("Number = %d, want 1", got)
	}
	if got := keepObjective(t, s, "Learn Go properly").Number; got != 2 {
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
	if got := keepObjective(t, s, "Ship cairn 1.0").Number; got != 1 {
		t.Errorf("Number = %d, want 1: objectives are numbered apart", got)
	}
}

func TestFindObjective_handsBackEverythingItWasGiven(t *testing.T) {
	s := openStore(t)
	want := domain.Objective{
		Title:       "Ship cairn 1.0",
		Description: "the one that unblocks the rest",
		State:       domain.OnHold,
		Priority:    domain.High,
		Labels:      []string{"perso", "spec"},
		Deadline:    time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC),
		Estimate:    90 * time.Minute,
	}
	kept, err := s.KeepObjective(want)
	if err != nil {
		t.Fatalf("KeepObjective: %v", err)
	}
	got, err := s.FindObjective(kept.Number)
	if err != nil {
		t.Fatalf("FindObjective: %v", err)
	}
	want.Number = kept.Number
	if got.Title != want.Title || got.Description != want.Description {
		t.Errorf("Title/Description = %q/%q, want %q/%q", got.Title, got.Description, want.Title, want.Description)
	}
	if got.State != want.State || got.Priority != want.Priority {
		t.Errorf("State/Priority = %q/%q, want %q/%q", got.State, got.Priority, want.State, want.Priority)
	}
	if !got.Deadline.Equal(want.Deadline) {
		t.Errorf("Deadline = %v, want %v", got.Deadline, want.Deadline)
	}
	if got.Estimate != want.Estimate {
		t.Errorf("Estimate = %v, want %v", got.Estimate, want.Estimate)
	}
	if len(got.Labels) != 2 || got.Labels[0] != "perso" || got.Labels[1] != "spec" {
		t.Errorf("Labels = %v, want [perso spec]", got.Labels)
	}
}

// The labels of an objective live in a table of their own, so a task and an
// objective sharing a number must not share labels.
func TestFindObjective_doesNotPickUpATasksLabels(t *testing.T) {
	s := openStore(t)
	task, err := domain.NewTask("Write the spec")
	if err != nil {
		t.Fatalf("NewTask: %v", err)
	}
	task.Labels = []string{"spec"}
	if _, err := s.Keep(task); err != nil {
		t.Fatalf("Keep: %v", err)
	}
	got := keepObjective(t, s, "Ship cairn 1.0")
	found, err := s.FindObjective(got.Number)
	if err != nil {
		t.Fatalf("FindObjective: %v", err)
	}
	if len(found.Labels) != 0 {
		t.Errorf("Labels = %v, want none: task 1 and objective 1 are not the same thing", found.Labels)
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
	if all[0].Title != "Ship cairn 1.0" || all[1].Title != "Learn Go properly" {
		t.Errorf("got %q then %q, want them in the order they were added", all[0].Title, all[1].Title)
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
	o.Title = "Ship cairn 1.0, for real"
	o.State = domain.Reached
	o.Labels = []string{"done"}
	if err := s.UpdateObjective(o); err != nil {
		t.Fatalf("UpdateObjective: %v", err)
	}
	got, err := s.FindObjective(o.Number)
	if err != nil {
		t.Fatalf("FindObjective: %v", err)
	}
	if got.Title != o.Title || got.State != domain.Reached {
		t.Errorf("got %q/%q, want %q/%q", got.Title, got.State, o.Title, domain.Reached)
	}
	if len(got.Labels) != 1 || got.Labels[0] != "done" {
		t.Errorf("Labels = %v, want [done]", got.Labels)
	}
}

// An objective carries a set of labels, so one left out of an update is one
// taken away — the same rule a task follows.
func TestUpdateObjective_replacesTheLabelsRatherThanAddingToThem(t *testing.T) {
	s := openStore(t)
	o := keepObjective(t, s, "Ship cairn 1.0")
	o.Labels = []string{"perso", "spec"}
	if err := s.UpdateObjective(o); err != nil {
		t.Fatalf("UpdateObjective: %v", err)
	}
	o.Labels = []string{"spec"}
	if err := s.UpdateObjective(o); err != nil {
		t.Fatalf("UpdateObjective: %v", err)
	}
	got, err := s.FindObjective(o.Number)
	if err != nil {
		t.Fatalf("FindObjective: %v", err)
	}
	if len(got.Labels) != 1 || got.Labels[0] != "spec" {
		t.Errorf("Labels = %v, want [spec]", got.Labels)
	}
}

func TestUpdateObjective_saysWhenThereIsNone(t *testing.T) {
	err := openStore(t).UpdateObjective(domain.Objective{Number: 99, Title: "Nowhere"})
	if !errors.Is(err, ErrNoSuchObjective) {
		t.Errorf("err = %v, want ErrNoSuchObjective", err)
	}
}

func TestDeleteObjective_removesIt(t *testing.T) {
	s := openStore(t)
	o := keepObjective(t, s, "Ship cairn 1.0")
	if err := s.DeleteObjective(o.Number); err != nil {
		t.Fatalf("DeleteObjective: %v", err)
	}
	if _, err := s.FindObjective(o.Number); !errors.Is(err, ErrNoSuchObjective) {
		t.Errorf("err = %v, want ErrNoSuchObjective after a delete", err)
	}
}

func TestDeleteObjective_takesItsLabelsWithIt(t *testing.T) {
	s := openStore(t)
	o := keepObjective(t, s, "Ship cairn 1.0")
	o.Labels = []string{"perso"}
	if err := s.UpdateObjective(o); err != nil {
		t.Fatalf("UpdateObjective: %v", err)
	}
	if err := s.DeleteObjective(o.Number); err != nil {
		t.Fatalf("DeleteObjective: %v", err)
	}
	left, err := s.labelsOf(objectiveLabels, o.Number)
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

	task, err := domain.NewTask("Write the spec")
	if err != nil {
		t.Fatalf("NewTask: %v", err)
	}
	task.Objective = o.Number
	kept, err := s.Keep(task)
	if err != nil {
		t.Fatalf("Keep: %v", err)
	}
	found, err := s.Find(kept.Number)
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	if found.Objective != o.Number {
		t.Errorf("Objective = %d, want %d", found.Objective, o.Number)
	}
}

func TestUpdate_changesTheObjectiveATaskServes(t *testing.T) {
	s := openStore(t)
	o := keepObjective(t, s, "Ship cairn 1.0")
	task := keep(t, s, "Write the spec")

	task.Objective = o.Number
	if err := s.Update(task); err != nil {
		t.Fatalf("Update: %v", err)
	}
	found, err := s.Find(task.Number)
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	if found.Objective != o.Number {
		t.Errorf("Objective = %d, want %d", found.Objective, o.Number)
	}

	found.Objective = domain.NoObjective
	if err := s.Update(found); err != nil {
		t.Fatalf("Update: %v", err)
	}
	loose, err := s.Find(task.Number)
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	if loose.Objective != domain.NoObjective {
		t.Errorf("Objective = %d, want none", loose.Objective)
	}
}

// All and Under read the same columns as Find, and the three lists are written
// out separately. A column added to one and forgotten in the others comes back
// as a zero nobody notices, so each is asked the same question.
func TestAllAndUnder_readTheObjectiveTaskstoo(t *testing.T) {
	s := openStore(t)
	o := keepObjective(t, s, "Ship cairn 1.0")
	parent := keep(t, s, "Ship it")
	parent.Objective = o.Number
	if err := s.Update(parent); err != nil {
		t.Fatalf("Update: %v", err)
	}
	child := keep(t, s, "Write the spec")
	child.Parent = parent.Number
	if err := s.Update(child); err != nil {
		t.Fatalf("Update: %v", err)
	}

	all, err := s.All()
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if all[0].Objective != o.Number {
		t.Errorf("All: Objective = %d, want %d", all[0].Objective, o.Number)
	}
	under, err := s.Under(parent.Number)
	if err != nil {
		t.Fatalf("Under: %v", err)
	}
	if len(under) != 1 || under[0].Objective != domain.NoObjective {
		t.Errorf("Under: %+v, want one sub-task serving no objective", under)
	}
}

func TestTasksServing_handsBackWhatWasPutUnderTheObjective(t *testing.T) {
	s := openStore(t)
	o := keepObjective(t, s, "Ship cairn 1.0")
	other := keepObjective(t, s, "Learn Go properly")

	for _, title := range []string{"Write the spec", "Ship it"} {
		task := keep(t, s, title)
		task.Objective = o.Number
		if err := s.Update(task); err != nil {
			t.Fatalf("Update: %v", err)
		}
	}
	elsewhere := keep(t, s, "Read the book")
	elsewhere.Objective = other.Number
	if err := s.Update(elsewhere); err != nil {
		t.Fatalf("Update: %v", err)
	}
	keep(t, s, "Nothing to do with it")

	serving, err := s.TasksServing(o.Number)
	if err != nil {
		t.Fatalf("TasksServing: %v", err)
	}
	if len(serving) != 2 {
		t.Fatalf("%d tasks serving, want 2", len(serving))
	}
	if serving[0].Title != "Write the spec" || serving[1].Title != "Ship it" {
		t.Errorf("got %q then %q, want them in the order they were added", serving[0].Title, serving[1].Title)
	}
}

func TestTasksServing_handsBackNothingWhenNothingWasPutUnderIt(t *testing.T) {
	s := openStore(t)
	o := keepObjective(t, s, "Ship cairn 1.0")
	keep(t, s, "Nothing to do with it")

	serving, err := s.TasksServing(o.Number)
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

	serving, err := s.TasksServing(domain.NoObjective)
	if err != nil {
		t.Fatalf("TasksServing: %v", err)
	}
	if len(serving) != 0 {
		t.Errorf("%d tasks serving nothing, want none", len(serving))
	}
}
