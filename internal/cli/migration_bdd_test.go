package cli

import (
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
	var draft string
	var board *tui.Looked
	opened := false
	sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		draft, opened = "", false
		was := openGui
		openGui = func(_ engine.Work, _ engine.Pages, kept, _ string, _ *debug.Log, _, _ int) error {
			draft, opened = kept, true
			return nil
		}
		t.Cleanup(func() { openGui = was })
		board = nil
		wasTui := openTui
		openTui = func(_ *cobra.Command, model tea.Model) error {
			first, _ := model.Init()().(tea.BatchMsg)[0]().(tui.Looked)
			board = &first
			return nil
		}
		t.Cleanup(func() { openTui = wasTui })
		return ctx, nil
	})
	sc.Step(`^the board is opened on that store$`, func() error {
		if board == nil {
			return fmt.Errorf("the board was never opened:\n%s%s", w.Out.Stdout, w.Out.Stderr)
		}
		if !errors.Is(board.Err, engine.ErrStoreBehind) {
			return fmt.Errorf("the board was not opened on a store behind: %w", board.Err)
		}
		return nil
	})
	sc.Step(`^the window is opened on that store$`, func() error {
		if !opened {
			return fmt.Errorf("the window was never opened:\n%s%s", w.Out.Stdout, w.Out.Stderr)
		}
		return test.WantContains(draft, w.StorePath, "the draft of the window")
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
