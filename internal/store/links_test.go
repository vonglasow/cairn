package store

import (
	"database/sql"
	"testing"

	"cairn/internal/objective"
)

type counting struct {
	runs
	queries int
}

func (c *counting) Query(query string, args ...any) (*sql.Rows, error) {
	c.queries++
	return c.runs.Query(query, args...)
}

func TestNarrowed_readsWhatTasksCarryAllAtOnce(t *testing.T) {
	s := openStore(t)
	for _, title := range []string{"one", "two", "three", "four", "five", "six"} {
		keep(t, s, title)
	}
	asked := &counting{runs: s.on}
	s.on = asked
	all, err := s.All()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 6 {
		t.Fatalf("read %d tasks, want 6", len(all))
	}
	if asked.queries > 4 {
		t.Errorf("reading six tasks asked %d questions, want the tasks and what they carry, at most 4", asked.queries)
	}
}

func TestNarrowedObjectives_readsTheirLabelsAllAtOnce(t *testing.T) {
	s := openStore(t)
	for _, title := range []string{"one", "two", "three", "four"} {
		built, err := objective.NewObjective(title)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := s.KeepObjective(built); err != nil {
			t.Fatal(err)
		}
	}
	asked := &counting{runs: s.on}
	s.on = asked
	all, err := s.AllObjectives()
	if err != nil || len(all) != 4 {
		t.Fatalf("read %d objectives, %v", len(all), err)
	}
	if asked.queries > 2 {
		t.Errorf("reading four objectives asked %d questions, want at most 2", asked.queries)
	}
}
