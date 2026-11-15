package sqlite

import (
	"errors"
	"testing"

	"cairn/internal/engine/task"
)

func threeInTheBacklog(t *testing.T) *Store {
	t.Helper()
	s := openStore(t)
	for _, title := range []string{"one", "two", "three"} {
		keep(t, s, title)
	}
	return s
}

func TestMove_saysThereIsNoSuchTask(t *testing.T) {
	s := threeInTheBacklog(t)
	if _, err := s.MoveAfter(1, 99); !errors.Is(err, task.ErrNoSuchTask) {
		t.Errorf("moving behind task 99 answered %v", err)
	}
	if _, err := s.MoveFirst(99); !errors.Is(err, task.ErrNoSuchTask) {
		t.Errorf("moving task 99 answered %v", err)
	}
}
