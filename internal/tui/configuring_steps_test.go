package tui

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/cucumber/godog"

	"cairn/internal/engine"
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

func TestTheBeatReadsTheStoreAndTheConfigurationAndBeatsAgain(t *testing.T) {
	b := New(newFake(), GlyphsFor(Unicode), time.Now, 100, 40)
	b.configured.config = engine.Config{File: filepath.Join(t.TempDir(), "config.toml")}
	_, cmd := b.Update(Beat(time.Now()))
	if batch, ok := cmd().(tea.BatchMsg); !ok || len(batch) != 3 || batch[1] == nil {
		t.Errorf("the beat asks %v", cmd())
	}
}

func TestABoardWithNoConfigurationReadReadsNoneAgain(t *testing.T) {
	b := New(newFake(), GlyphsFor(Unicode), time.Now, 100, 40)
	if b.rereading() != nil {
		t.Error("a board that read no configuration reads one again")
	}
}

func TestAKeyIsCompletedOnlyAfterSetOrUnset(t *testing.T) {
	b := New(newFake(), GlyphsFor(Unicode), time.Now, 100, 40)
	b.prompt = Commanding
	for typed, want := range map[string]int{"set pages.t": 2, "unset pages.t": 2, "set ": 5, "task edit": 0, "set pages.time ": 0, "set " + "--" + "global win": 3} {
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
