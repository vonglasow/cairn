package store

import (
	"errors"
	"strings"
	"sync"
	"testing"

	"cairn/internal/task"
)

func TestWithin_keepsNothingOfAnActThatFailedHalfway(t *testing.T) {
	s := openStore(t)
	failed := errors.New("halfway")
	err := s.Within(func(tx *Store) error {
		keep(t, tx, "Write the spec")
		return failed
	})
	if !errors.Is(err, failed) {
		t.Fatalf("Within = %v, want the act's own failure", err)
	}
	all, err := s.All()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 0 {
		t.Errorf("%d tasks kept from an act that failed, want none", len(all))
	}
}

func TestWithin_keepsWhatAnActThatWorkedWrote(t *testing.T) {
	s := openStore(t)
	if err := s.Within(func(tx *Store) error {
		keep(t, tx, "Write the spec")
		keep(t, tx, "Ship it")
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	all, err := s.All()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Errorf("%d tasks kept, want 2", len(all))
	}
}

func TestShared_beginsEveryTransactionAsAWrite(t *testing.T) {
	if got := shared("/somewhere/cairn.db"); !strings.Contains(got, "_txlock=immediate") {
		t.Errorf("the store is opened as %q, want its transactions to begin as writes", got)
	}
}

func TestWithin_givesTasksAddedTogetherAPlaceEach(t *testing.T) {
	path := openStore(t).path
	const together = 6
	var wg sync.WaitGroup
	errs := make(chan error, together)
	for range together {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s, err := Open(path)
			if err != nil {
				errs <- err
				return
			}
			defer func() { _ = s.Close() }()
			built, err := task.NewTask("together")
			if err != nil {
				errs <- err
				return
			}
			errs <- s.Within(func(tx *Store) error {
				_, err := tx.Keep(built)
				return err
			})
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()
	rows, err := s.db.Query(`SELECT position, COUNT(*) FROM tasks GROUP BY state, position HAVING COUNT(*) > 1`)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	if rows.Next() {
		t.Error("two tasks added together were given the same place")
	}
}
