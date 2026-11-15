package sqlite

import (
	"database/sql"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"cairn/internal/engine/task"
)

func openStore(t *testing.T) *Store {
	t.Helper()
	s, err := Open(filepath.Join(t.TempDir(), "cairn.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func keep(t *testing.T, s *Store, title string) task.Task {
	t.Helper()
	built, err := task.NewTask(title)
	if err != nil {
		t.Fatalf("NewTask: %v", err)
	}
	kept, err := s.Keep(built)
	if err != nil {
		t.Fatalf("Keep: %v", err)
	}
	return kept
}

func changed(built task.Task, mutate func(a *task.AsTold)) task.Task {
	a := &task.AsTold{}
	*a = task.From(built)
	mutate(a)
	return a.Task()
}

func factsOf(t task.Task) task.Carried {
	f := t.Carried()
	return f
}

func closed(t *testing.T) *Store {
	t.Helper()
	s := openStore(t)
	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	return s
}

func TestStore_reportsLabelsItCannotRead(t *testing.T) {
	s := openStore(t)
	kept := keep(t, s, "Write the spec")
	if _, err := s.db.Exec(`DROP TABLE labels`); err != nil {
		t.Fatalf("dropping labels: %v", err)
	}

	if _, err := s.Find(factsOf(kept).Number); err == nil {
		t.Error("Find should report labels it cannot read")
	}
	if _, err := s.All(); err == nil {
		t.Error("All should report labels it cannot read")
	}
}

func TestStore_reportsLabelsItCannotKeep(t *testing.T) {
	s := openStore(t)
	built, _ := task.NewTask("Write the spec")
	built = changed(built, func(a *task.AsTold) {
		a.Labelled("spec")
		a.Labelled("spec")
	})

	if _, err := s.Keep(built); err == nil {
		t.Error("Keep should report labels it cannot record")
	}
}

func TestStore_reportsLabelsItCannotReplace(t *testing.T) {
	s := openStore(t)
	built := keep(t, s, "Write the spec")
	built = changed(built, func(a *task.AsTold) {
		a.Labelled("spec")
		a.Labelled("spec")
	})

	if err := s.Update(built); err == nil {
		t.Error("Update should report labels it cannot record")
	}
}

func TestStore_reportsADeadlineItCannotRead(t *testing.T) {
	s := openStore(t)
	kept := keep(t, s, "Write the spec")
	if _, err := s.db.Exec(`UPDATE tasks SET deadline = 'soon' WHERE number = ?`, factsOf(kept).Number); err != nil {
		t.Fatalf("writing a bad deadline: %v", err)
	}

	if _, err := s.Find(factsOf(kept).Number); err == nil {
		t.Error("Find should report a deadline it cannot make sense of")
	}
	if _, err := s.All(); err == nil {
		t.Error("All should report a deadline it cannot make sense of")
	}
}

func TestStore_reportsLabelsItCannotClear(t *testing.T) {
	s := openStore(t)
	built := keep(t, s, "Write the spec")
	if _, err := s.db.Exec(`DROP TABLE labels`); err != nil {
		t.Fatalf("dropping labels: %v", err)
	}

	if err := s.Update(built); err == nil {
		t.Error("Update should report labels it cannot clear")
	}
	if err := s.Delete(factsOf(built).Number); err == nil {
		t.Error("Delete should report labels it cannot clear")
	}
}

func TestOpen_waitsForAStoreSomebodyElseIsMaking(t *testing.T) {
	path := filepath.Join(t.TempDir(), "made-at-once", "cairn.db")

	const together = 4
	var wg sync.WaitGroup
	failed := make(chan error, together)
	for range together {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s, err := Open(path)
			if err != nil {
				failed <- err
				return
			}
			_ = s.Close()
		}()
	}
	wg.Wait()
	close(failed)

	for err := range failed {
		t.Errorf("one of the %d refused to wait: %v", together, err)
	}
}

func TestOpen_reportsAShapeItCouldNotLayDown(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cairn.db")

	db, err := sql.Open("sqlite", shared(path))
	if err != nil {
		t.Fatalf("opening the file: %v", err)
	}
	if _, err := db.Exec(ledger); err != nil {
		t.Fatalf("making the ledger: %v", err)
	}
	if _, err := db.Exec(`CREATE VIEW tasks AS SELECT 1 AS number`); err != nil {
		t.Fatalf("putting something in the way: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("closing: %v", err)
	}

	_, err = Open(path)
	if err == nil {
		t.Fatal("Open said nothing, want it to report a shape it could not lay down")
	}
	if !strings.Contains(err.Error(), "applying") {
		t.Errorf("Open said %q, want it to name the change it could not apply", err)
	}
}

func mustTask(t *testing.T, title string) task.Task {
	t.Helper()
	built, err := task.NewTask(title)
	if err != nil {
		t.Fatal(err)
	}
	return built
}
