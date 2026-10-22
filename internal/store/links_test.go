package store

import (
	"database/sql"
	"slices"
	"testing"

	"cairn/internal/objective"
	"cairn/internal/task"
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
	if asked.queries > 5 {
		t.Errorf("reading six tasks asked %d questions, want the tasks and the four things they carry — labels, blockers, objectives, steps — at most 5", asked.queries)
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

func TestLinkIsReadFromBothEnds(t *testing.T) {
	s := openStore(t)
	keep(t, s, "one")
	two := keep(t, s, "two")
	c := two.Carried()
	c.Linked = []int{1}
	if err := s.Update(task.Made(c)); err != nil {
		t.Fatal(err)
	}
	for number, want := range map[int][]int{1: {2}, 2: {1}} {
		found, err := s.Find(number)
		if err != nil {
			t.Fatal(err)
		}
		if got := found.Carried().Linked; !slices.Equal(got, want) {
			t.Errorf("Find(%d).Linked = %v, want %v", number, got, want)
		}
	}
	all, err := s.All()
	if err != nil {
		t.Fatal(err)
	}
	for _, listed := range all {
		c := listed.Carried()
		want := map[int][]int{1: {2}, 2: {1}}[c.Number]
		if !slices.Equal(c.Linked, want) {
			t.Errorf("All: task %d Linked = %v, want %v", c.Number, c.Linked, want)
		}
	}
}

func TestDeletingATaskDropsItsLinks(t *testing.T) {
	s := openStore(t)
	keep(t, s, "one")
	two := keep(t, s, "two")
	keep(t, s, "three")
	c := two.Carried()
	c.Linked = []int{1, 3}
	if err := s.Update(task.Made(c)); err != nil {
		t.Fatal(err)
	}
	if err := s.Delete(2); err != nil {
		t.Fatal(err)
	}
	for _, number := range []int{1, 3} {
		found, err := s.Find(number)
		if err != nil {
			t.Fatal(err)
		}
		if got := found.Carried().Linked; len(got) != 0 {
			t.Errorf("Find(%d).Linked = %v, want none", number, got)
		}
	}
	var named int
	if err := s.on.QueryRow(`SELECT COUNT(*) FROM links WHERE low = 2 OR high = 2`).Scan(&named); err != nil {
		t.Fatal(err)
	}
	if named != 0 {
		t.Errorf("%d links still name task 2, want 0", named)
	}
}

func TestLinksHoldEachPairOnce(t *testing.T) {
	s := openStore(t)
	if _, err := s.on.Exec(`INSERT INTO links (low, high) VALUES (2, 1)`); err == nil {
		t.Errorf("the pair (2, 1) was held, want the lower number first")
	}
	if _, err := s.on.Exec(`INSERT INTO links (low, high) VALUES (1, 2)`); err != nil {
		t.Fatalf("the pair (1, 2) was refused: %v", err)
	}
	if _, err := s.on.Exec(`INSERT INTO links (low, high) VALUES (1, 2)`); err == nil {
		t.Errorf("the pair (1, 2) was held twice, want it held once")
	}
}
