package sqlite

import (
	"cairn/internal/engine/filter"
	"errors"
	"slices"
	"testing"

	"cairn/internal/engine/objective"
	"cairn/internal/engine/task"
)

func keepObjective(t *testing.T, s *Store, title string, labels ...string) objective.Objective {
	t.Helper()
	built, err := objective.NewObjective(title)
	if err != nil {
		t.Fatalf("NewObjective: %v", err)
	}
	told := objective.From(built)
	for _, label := range labels {
		told.Labelled(label)
	}
	kept, err := s.KeepObjective(told.Objective())
	if err != nil {
		t.Fatalf("KeepObjective: %v", err)
	}
	return kept
}

func objectiveTitles(all []objective.Objective) []string {
	titles := []string{}
	for _, each := range all {
		titles = append(titles, each.Carried().Title)
	}
	return titles
}

func TestUpdateObjective_saysThereIsNoSuchObjective(t *testing.T) {
	s := openStore(t)
	built, err := objective.NewObjective("Nope")
	if err != nil {
		t.Fatal(err)
	}
	told := objective.From(built)
	told.Numbered(9)
	if err := s.UpdateObjective(told.Objective()); !errors.Is(err, objective.ErrNoSuchObjective) {
		t.Errorf("changing objective 9 answered %v", err)
	}
}

func TestDeleteObjective_removesItAndSaysWhenThereIsNone(t *testing.T) {
	s := openStore(t)
	keepObjective(t, s, "Ship it", "q4")
	keepObjective(t, s, "Stay")
	if err := s.DeleteObjective(1); err != nil {
		t.Fatalf("DeleteObjective: %v", err)
	}
	all, err := s.AllObjectives()
	if err != nil {
		t.Fatal(err)
	}
	if got := objectiveTitles(all); !slices.Equal(got, []string{"Stay"}) {
		t.Errorf("left %v, want [Stay]", got)
	}
	if left, err := s.labelsOf(objectiveLabels, 1); err != nil || len(left) != 0 {
		t.Errorf("the objective deleted still carries %v (%v)", left, err)
	}
	if err := s.DeleteObjective(1); !errors.Is(err, objective.ErrNoSuchObjective) {
		t.Errorf("deleting objective 1 twice answered %v", err)
	}
}

func TestNarrowedObjectives_keepsWhatCarriesTheLabel(t *testing.T) {
	s := openStore(t)
	keepObjective(t, s, "Ship it", "q4")
	keepObjective(t, s, "Stay")
	kept := &filter.Narrowing{}
	kept.Labelled([]string{"q4"})
	got, err := s.NarrowedObjectives(kept)
	if err != nil {
		t.Fatalf("NarrowedObjectives: %v", err)
	}
	if titles := objectiveTitles(got); !slices.Equal(titles, []string{"Ship it"}) {
		t.Fatalf("kept %v, want [Ship it]", titles)
	}
	if labels := got[0].Carried().Labels; !slices.Equal(labels, []string{"q4"}) {
		t.Errorf("read back labelled %v, want [q4]", labels)
	}
}

func TestNarrowed_keepsWhatServesAnyOfTheObjectivesNamedTogether(t *testing.T) {
	s := openStore(t)
	keepObjective(t, s, "Home automation")
	keepObjective(t, s, "Enclosure")
	for _, serves := range []int{1, 2, 0} {
		if _, err := s.Keep(changed(mustTask(t, "Task"), func(a *task.AsTold) {
			if serves != 0 {
				a.Serves(serves)
			}
		})); err != nil {
			t.Fatal(err)
		}
	}
	kept := filter.Narrowing{}
	kept.Serving([]int{1, 2})
	got, err := s.Narrowed(&kept)
	if err != nil || len(got) != 2 {
		t.Errorf("Narrowed serving 1 or 2 = %d tasks, %v, want 2", len(got), err)
	}
}
