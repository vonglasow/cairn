package store

import (
	"errors"
	"slices"
	"testing"

	"cairn/internal/domain"
	"cairn/internal/objective"
	"cairn/internal/task"
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

func TestKeepObjective_handsBackTheObjectiveNumbered(t *testing.T) {
	s := openStore(t)
	keepObjective(t, s, "First")
	kept := keepObjective(t, s, "Ship it", "q4")
	if c := kept.Carried(); c.Number != 2 || c.Title != "Ship it" {
		t.Errorf("handed back %d %q, want 2 \"Ship it\"", c.Number, c.Title)
	}
}

func TestFindObjective_readsBackWhatWasKept(t *testing.T) {
	s := openStore(t)
	deadline, err := domain.ParseDeadline("2026-12-01")
	if err != nil {
		t.Fatal(err)
	}
	built, err := objective.NewObjective("Ship it")
	if err != nil {
		t.Fatal(err)
	}
	told := objective.From(built)
	told.Labelled("q4")
	told.Due(deadline)
	if _, err := s.KeepObjective(told.Objective()); err != nil {
		t.Fatal(err)
	}
	got, err := s.FindObjective(1)
	if err != nil {
		t.Fatalf("FindObjective: %v", err)
	}
	c := got.Carried()
	if c.Number != 1 || c.Title != "Ship it" || !slices.Equal(c.Labels, []string{"q4"}) || domain.FormatDay(c.Deadline) != "2026-12-01" {
		t.Errorf("read back %+v, want what was kept", c)
	}
}

func TestUpdateObjective_replacesWhatWasKept(t *testing.T) {
	s := openStore(t)
	kept := keepObjective(t, s, "Ship it", "q4")
	told := objective.From(kept)
	told.Titled("Ship it well")
	told.Labelled("team")
	if err := s.UpdateObjective(told.Objective()); err != nil {
		t.Fatalf("UpdateObjective: %v", err)
	}
	got, err := s.FindObjective(1)
	if err != nil {
		t.Fatal(err)
	}
	if c := got.Carried(); c.Title != "Ship it well" || !slices.Equal(c.Labels, []string{"q4", "team"}) {
		t.Errorf("read back %q %v, want the change", c.Title, c.Labels)
	}
}

func TestUpdateObjective_saysThereIsNoSuchObjective(t *testing.T) {
	s := openStore(t)
	built, err := objective.NewObjective("Nope")
	if err != nil {
		t.Fatal(err)
	}
	told := objective.From(built)
	told.Numbered(9)
	if err := s.UpdateObjective(told.Objective()); !errors.Is(err, ErrNoSuchObjective) {
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
	if err := s.DeleteObjective(1); !errors.Is(err, ErrNoSuchObjective) {
		t.Errorf("deleting objective 1 twice answered %v", err)
	}
}

func TestNarrowedObjectives_keepsWhatCarriesTheLabel(t *testing.T) {
	s := openStore(t)
	keepObjective(t, s, "Ship it", "q4")
	keepObjective(t, s, "Stay")
	kept := &domain.Narrowing{}
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

func TestNarrowedObjectives_keepsWhatStandsInAColumn(t *testing.T) {
	s := openStore(t)
	kept := keepObjective(t, s, "Ship it", "q4")
	keepObjective(t, s, "Stay", "q4")
	told := objective.From(kept)
	told.Standing(objective.Reached)
	if err := s.UpdateObjective(told.Objective()); err != nil {
		t.Fatal(err)
	}
	narrowing := &domain.Narrowing{}
	narrowing.InColumn(string(objective.Reached))
	narrowing.Labelled([]string{"q4"})
	got, err := s.NarrowedObjectives(narrowing)
	if err != nil {
		t.Fatalf("NarrowedObjectives: %v", err)
	}
	if titles := objectiveTitles(got); !slices.Equal(titles, []string{"Ship it"}) {
		t.Errorf("kept %v, want [Ship it]", titles)
	}
}

func TestTasksServing_handsBackWhatServesTheObjective(t *testing.T) {
	s := openStore(t)
	goal := keepObjective(t, s, "Ship it")
	keep(t, s, "Aside")
	if _, err := s.Keep(changed(mustTask(t, "Write the spec"), func(a *task.AsTold) { a.Serves(goal.Carried().Number) })); err != nil {
		t.Fatal(err)
	}
	got, err := s.TasksServing(1)
	if err != nil {
		t.Fatalf("TasksServing: %v", err)
	}
	if len(got) != 1 || factsOf(got[0]).Title != "Write the spec" {
		t.Errorf("serving objective 1: %v, want only Write the spec", got)
	}
}
