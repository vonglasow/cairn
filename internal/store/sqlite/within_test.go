package sqlite

import (
	"sync"
	"testing"

	"cairn/internal/engine/task"
)

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
