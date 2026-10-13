package engine

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

// An act that loses work is refused until it comes confirmed, wherever it is
// asked from: the engine decides which acts ask first, and a surface only says
// that it asked.
func TestAnActThatLosesWorkWaitsToBeConfirmed(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cairn.db")
	if _, err := Do(TaskNew, path, Input{Given: map[string]string{"title": "Write the spec"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := Do(TaskNew, path, Input{Given: map[string]string{"title": "Ship it"}}); err != nil {
		t.Fatal(err)
	}
	_, err := Do(TaskDelete, path, Input{Given: map[string]string{"number": "1"}, Listed: map[string][]string{"also": {"2"}}})
	if !errors.Is(err, ErrNotConfirmed) {
		t.Fatalf("an unconfirmed deletion answered %v", err)
	}
	if !strings.Contains(err.Error(), "task 1, task 2") {
		t.Errorf("the refusal %q does not name what it would take", err)
	}
	if _, err := Do(TaskShow, path, Input{Given: map[string]string{"number": "1"}}); err != nil {
		t.Errorf("the task went anyway: %v", err)
	}
	if _, err := Do(TaskDelete, path, Input{Given: map[string]string{"number": "1"}, Confirmed: true}); err != nil {
		t.Errorf("a confirmed deletion answered %v", err)
	}
}

// Sweeping everything a narrowing reaches asks first; one task named does not.
func TestASweepWaitsToBeConfirmed(t *testing.T) {
	sweep := Input{Given: map[string]string{"state": "done"}}
	one := Input{Given: map[string]string{"number": "1"}}
	if !TaskArchive.AsksFirst(sweep) {
		t.Error("a sweep does not ask first")
	}
	if TaskArchive.AsksFirst(one) {
		t.Error("archiving one named task asks first")
	}
	if !TaskDelete.AsksFirst(one) || TaskNew.AsksFirst(one) {
		t.Error("only what loses work asks first")
	}
	e := openEngine(t)
	if _, err := e.Ask(TaskArchive, sweep); !errors.Is(err, ErrNotConfirmed) {
		t.Errorf("an unconfirmed sweep through the engine answered %v", err)
	}
}
