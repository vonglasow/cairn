package engine_test

import (
	"errors"
	"path/filepath"
	"testing"

	"cairn/internal/engine"
	"cairn/test"
)

func aStoreANewerCairnChanged(t *testing.T, held *engine.Engine) {
	t.Helper()
	if err := test.AsANewerCairn("never heard of", func() error {
		_, err := held.Ask(engine.MigrationApply, engine.Input{})
		return err
	}); err != nil {
		t.Fatal(err)
	}
}

func TestAsk_aStoreANewerCairnChangedIsRefusedToEverythingThatTouchesIt(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cairn.db")
	held, err := engine.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = held.Close() })
	aStoreANewerCairnChanged(t, held)
	for _, op := range []engine.Operation{engine.TaskList, engine.MigrationList} {
		if _, err := held.Ask(op, engine.Input{}); !errors.Is(err, engine.ErrNewerStore) {
			t.Errorf("Ask(%s) on a store a newer cairn changed = %v, want ErrNewerStore", op.Name(), err)
		}
	}
}

func TestHold_holdsAStoreFromANewerCairnRefusingEveryAsk(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cairn.db")
	e, err := engine.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	aStoreANewerCairnChanged(t, e)
	if err := e.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.Open(path); !errors.Is(err, engine.ErrNewerStore) {
		t.Errorf("Open = %v, want ErrNewerStore", err)
	}
	held, err := engine.Hold(engine.Home{Dir: t.TempDir(), Lookup: func(k string) (string, bool) { return path, k == "CAIRN_STORE" }}, engine.Acting{Through: "gui", By: "sam"})
	if err != nil {
		t.Fatalf("Hold: %v", err)
	}
	for _, op := range []engine.Operation{engine.TaskList, engine.MigrationList} {
		if _, err := held.Ask(op, engine.Input{}); !errors.Is(err, engine.ErrNewerStore) {
			t.Errorf("Ask(%s) = %v, want ErrNewerStore", op.Name(), err)
		}
	}
	if _, err := held.Version(); err != nil {
		t.Errorf("Version: %v", err)
	}
	if err := held.Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
}
