package engine

import (
	"cairn/internal/engine/migration"
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

func engineOn(t *testing.T) (*Engine, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "cairn.db")
	e, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = e.Close() })
	return e, path
}

func TestAsk_followsAStorePutBackUnderIt(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cairn.db")
	held, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = held.Close() })
	if _, err := held.Ask(TaskNew, given("title", "Write the spec")); err != nil {
		t.Fatal(err)
	}
	made, err := backupsAt(path).Make()
	if err != nil {
		t.Fatal(err)
	}
	before, err := held.Version()
	if err != nil {
		t.Fatal(err)
	}
	if err := backupsAt(path).Restore(made.Number); err != nil {
		t.Fatal(err)
	}
	if after, err := held.Version(); err != nil || after <= before {
		t.Errorf("Version = %d, %v after a backup was put back, want it above %d", after, err, before)
	}
	if _, err := held.Ask(TaskNew, given("title", "Added on the board")); err != nil {
		t.Fatal(err)
	}
	listed, err := Do(TaskList, path, Input{})
	if err != nil {
		t.Fatal(err)
	}
	if len(listed.Tasks) != 2 {
		t.Errorf("%d tasks in the store, want the one put back and the one added after", len(listed.Tasks))
	}
}

func TestAsk_aStoreFallenBehindIsRefusedToWhatNeedsItCurrent(t *testing.T) {
	path := behindStore(t)
	held, err := OpenBehind(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = held.Close() })
	if _, err := held.Ask(TaskList, Input{}); !errors.Is(err, migration.ErrStoreBehind) {
		t.Errorf("Ask(TaskList) on a store behind = %v, want ErrStoreBehind", err)
	}
	if _, err := held.Ask(MigrationList, Input{}); err != nil {
		t.Errorf("Ask(MigrationList) on a store behind = %v, want it to pass", err)
	}
}

func TestAskRefusesASearchForNothing(t *testing.T) {
	e, _ := engineOn(t)
	add(t, e, "Write the spec")

	for _, named := range [][]string{{""}, {"   "}, {" ", ""}} {
		_, err := e.Ask(TaskSearch, Input{Listed: map[string][]string{"named": named}})
		if err == nil {
			t.Fatalf("Ask(TaskSearch, %q) answered, want it refused", named)
		}
		if !strings.Contains(err.Error(), "there is nothing to look for") {
			t.Errorf("Ask(TaskSearch, %q) = %v, want it to say there is nothing to look for", named, err)
		}
	}
}
