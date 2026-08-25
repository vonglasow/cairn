package engine

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// storeAt and behindStore are declared in store_test.go, which needs the same
// two shapes.

// garbageStore gives a path to a file that is not a store at all, so opening
// it fails with something other than the store being behind.
func garbageStore(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "cairn.db")
	if err := os.WriteFile(path, []byte("not a store"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	return path
}

// walFileGone reports whether the store at path was left with nothing
// checkpointed and waiting — which only happens once every connection to it
// has closed.
func walFileGone(t *testing.T, path string) bool {
	t.Helper()
	_, err := os.Stat(path + "-wal")
	return errors.Is(err, os.ErrNotExist)
}

func TestDo_opensTheStoreAndClosesItAgain(t *testing.T) {
	path := storeAt(t)
	op := Operation{
		Needs: AStore,
		Run: func(on On, _ Input) (Answer, error) {
			if on.Engine == nil {
				t.Fatal("on.Engine is nil, want an engine open on the store")
			}
			if _, err := on.Engine.Add("hello", Changes{}); err != nil {
				t.Fatalf("Add: %v", err)
			}
			return Answer{}, nil
		},
	}

	if _, err := Do(op, path, Input{}); err != nil {
		t.Fatalf("Do: %v", err)
	}

	if !walFileGone(t, path) {
		t.Errorf("the store was not closed: a write-ahead log is still waiting")
	}
}

func TestDo_needsAStoreBehind_opensTheStoreWhateverShapeItIsIn(t *testing.T) {
	path := behindStore(t)
	called := false
	op := Operation{
		Needs: AStoreBehind,
		Run: func(on On, _ Input) (Answer, error) {
			called = true
			if on.Engine == nil {
				t.Fatal("on.Engine is nil, want an engine open on the store")
			}
			if _, err := on.Engine.Held(); err != nil {
				t.Fatalf("Held: %v", err)
			}
			return Answer{}, nil
		},
	}

	if _, err := Do(op, path, Input{}); err != nil {
		t.Fatalf("Do: %v", err)
	}
	if !called {
		t.Errorf("Run was never called")
	}
}

func TestDo_needsAStore_reportsWhenTheStoreIsBehind(t *testing.T) {
	path := behindStore(t)
	op := Operation{
		Needs: AStore,
		Run: func(On, Input) (Answer, error) {
			t.Fatal("Run was called on a store that should have been refused")
			return Answer{}, nil
		},
	}

	_, err := Do(op, path, Input{})
	if !errors.Is(err, ErrStoreBehind) {
		t.Fatalf("err = %v, want ErrStoreBehind", err)
	}
	if !strings.Contains(err.Error(), "migration apply") {
		t.Errorf("err = %q, want it to say to run the migration", err)
	}
}

func TestDo_needsAStore_passesThroughAnErrorThatIsNotAboutBeingBehind(t *testing.T) {
	path := garbageStore(t)
	op := Operation{
		Needs: AStore,
		Run: func(On, Input) (Answer, error) {
			t.Fatal("Run was called on a store that could not be opened")
			return Answer{}, nil
		},
	}

	_, err := Do(op, path, Input{})
	if err == nil {
		t.Fatal("Do: want an error, got none")
	}
	if errors.Is(err, ErrStoreBehind) {
		t.Errorf("err = %v, want an error unrelated to the store being behind", err)
	}
}

func TestDo_needsNoStore_opensNothingButHandsBackBackups(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cairn.db")
	op := Operation{
		Needs: NoStore,
		Run: func(on On, _ Input) (Answer, error) {
			if on.Engine != nil {
				t.Errorf("on.Engine = %v, want nil: nothing should have been opened", on.Engine)
			}
			if on.Backups == nil {
				t.Errorf("on.Backups = nil, want the on.Backups kept beside the store")
			}
			return Answer{}, nil
		},
	}

	if _, err := Do(op, path, Input{}); err != nil {
		t.Fatalf("Do: %v", err)
	}
}

func TestDo_closesTheStoreEvenWhenRunFails(t *testing.T) {
	path := storeAt(t)
	failure := errors.New("boom")
	op := Operation{
		Needs: AStore,
		Run: func(on On, _ Input) (Answer, error) {
			if on.Engine == nil {
				t.Fatal("on.Engine is nil, want an engine open on the store")
			}
			if _, err := on.Engine.Add("hello", Changes{}); err != nil {
				t.Fatalf("Add: %v", err)
			}
			return Answer{}, failure
		},
	}

	_, err := Do(op, path, Input{})
	if !errors.Is(err, failure) {
		t.Fatalf("err = %v, want %v", err, failure)
	}
	if !walFileGone(t, path) {
		t.Errorf("the store was not closed after Run failed: a write-ahead log is still waiting")
	}
}

func TestDo_refusesANumberThatDoesNotParseBeforeOpeningTheStore(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cairn.db")
	op := Operation{
		Needs: AStore,
		Takes: []Argument{{Name: "number", Kind: Number, Of: "a task"}},
		Run: func(On, Input) (Answer, error) {
			t.Fatal("Run was called on a number that does not parse")
			return Answer{}, nil
		},
	}

	_, err := Do(op, path, Input{Given: map[string]string{"number": "spec"}})
	if err == nil {
		t.Fatal("Do: want an error, got none")
	}
	if _, statErr := os.Stat(path); !errors.Is(statErr, os.ErrNotExist) {
		t.Errorf("the store exists at %s, want a refusal to leave it untouched", path)
	}
}

func TestInput_tellsNotNamedFromNamedAsNothing(t *testing.T) {
	in := Input{
		Given:  map[string]string{"desc": "", "state": "todo"},
		Listed: map[string][]string{"label": {}},
	}

	if in.Named("priority") {
		t.Error("priority was never named, yet Named says it was")
	}
	for _, name := range []string{"desc", "state", "label"} {
		if !in.Named(name) {
			t.Errorf("%s was named, yet Named says it was not", name)
		}
	}

	if got := in.Change("priority"); got != nil {
		t.Errorf(`Change("priority") = %q, want nothing: it was never named`, *got)
	}
	if got := in.Change("desc"); got == nil || *got != "" {
		t.Errorf(`Change("desc") = %v, want a pointer to "": it was named as nothing`, got)
	}
	if got := in.Change("state"); got == nil || *got != "todo" {
		t.Errorf(`Change("state") = %v, want "todo"`, got)
	}

	if got := in.Changed("priority"); got != nil {
		t.Errorf(`Changed("priority") = %v, want nothing: it was never named`, *got)
	}
	if got := in.Changed("label"); got == nil || len(*got) != 0 {
		t.Errorf(`Changed("label") = %v, want an empty set: it was named as nothing`, got)
	}
}
