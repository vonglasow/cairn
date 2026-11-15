package cli

import (
	"cairn/internal/board"
	"cairn/internal/engine/migration"
	"cairn/internal/gui"
	"cairn/internal/gui/design"
	"context"
	"errors"
	"fmt"
	"testing"

	"cairn/internal/debug"
	"cairn/internal/engine"
	"cairn/internal/tui"
	"cairn/test"

	tea "charm.land/bubbletea/v2"
	"github.com/cucumber/godog"
	"github.com/spf13/cobra"
)

func initMigration(t *testing.T) func(*godog.ScenarioContext) {
	return func(sc *godog.ScenarioContext) {
		w := newWorld(t)
		w.register(sc)
		w.registerStore(sc)
		w.registerTimesheet(sc)
		w.registerWindow(sc, t)
	}
}

func (w *world) registerWindow(sc *godog.ScenarioContext, t *testing.T) {
	var shown string
	var looked *tui.Looked
	opened := false
	sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		shown, opened = "", false
		w.window = func(_ board.Work, _ engine.Pages, _ design.Skin, _ *debug.Log, _, _ int, here gui.Here) error {
			shown, opened = here.Store, true
			return nil
		}
		looked = nil
		w.terminal = func(_ *cobra.Command, model tea.Model) error {
			first, _ := model.Init()().(tea.BatchMsg)[0]().(tui.Looked)
			looked = &first
			return nil
		}
		return ctx, nil
	})
	sc.Step(`^the board is opened on that store$`, func() error {
		if looked == nil {
			return fmt.Errorf("the board was never opened:\n%s%s", w.Out.Stdout, w.Out.Stderr)
		}
		if !errors.Is(looked.Err, migration.ErrStoreBehind) {
			return fmt.Errorf("the board was not opened on a store behind: %w", looked.Err)
		}
		return nil
	})
	sc.Step(`^the window is opened on that store$`, func() error {
		if !opened {
			return fmt.Errorf("the window was never opened:\n%s%s", w.Out.Stdout, w.Out.Stderr)
		}
		return test.WantContains(shown, w.StorePath, "the store the window shows")
	})
	sc.Step(`^(\d+) changes? (?:is|are) still waiting$`, func(n int) error {
		before := w.Out
		w.Run(`cairn migration list`)
		listing := w.Out.Stdout
		w.Out = before
		return test.WantContains(listing, fmt.Sprintf("%d change(s) waiting", n), "the listing")
	})
}

func TestAcceptance_Migration(t *testing.T) {
	runAcceptance(t, "migration.feature", initMigration(t), "~@later")
}
