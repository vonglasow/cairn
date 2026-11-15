package tui

import (
	"fmt"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/cucumber/godog"
)

func (w *world) registerConfiguring(sc *godog.ScenarioContext) {
	sc.When(`^the beat comes$`, w.beatComes)
	sc.Then(`^the command line reads "([^"]*)"$`, func(said string) error {
		if w.b.typed != said {
			return fmt.Errorf("the command line reads %q, not %q", w.b.typed, said)
		}
		return nil
	})
}

func (w *world) beatComes() error {
	_, cmd := w.b.Update(Beat(stood))
	batch, ok := cmd().(tea.BatchMsg)
	if !ok || len(batch) == 0 {
		return fmt.Errorf("the beat asks nothing")
	}
	for _, one := range batch[:len(batch)-1] {
		w.settle(one)
	}
	return nil
}

func TestAKeyIsCompletedOnlyAfterSetOrUnset(t *testing.T) {
	b := New(newFake(), GlyphsFor(Unicode), time.Now, 100, 40)
	b.prompt = Commanding
	for typed, want := range map[string]int{"set pages.t": 2, "unset pages.t": 2, "set ": 5, "task edit": 0, "set pages.time ": 0, "set " + "--" + "global win": 4} {
		b.typed = typed
		if got := len(b.completions()); got != want {
			t.Errorf("%q completes %d keys, want %d", typed, got, want)
		}
	}
	b.prompt = Searching
	b.typed = "set pages.t"
	if len(b.completions()) != 0 {
		t.Error("a search completes keys")
	}
}
