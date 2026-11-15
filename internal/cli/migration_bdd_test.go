package cli

import (
	"cairn/internal/board"
	"cairn/internal/gui"
	"cairn/internal/gui/design"
	"context"
	"testing"

	"cairn/internal/debug"
	"cairn/internal/engine"

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

func (w *world) registerWindow(sc *godog.ScenarioContext, _ *testing.T) {
	sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		w.window = func(board.Work, engine.Pages, design.Skin, *debug.Log, int, int, gui.Here) error { return nil }
		w.terminal = func(*cobra.Command, tea.Model) error { return nil }
		return ctx, nil
	})
}

func TestAcceptance_Migration(t *testing.T) {
	runAcceptance(t, "migration.feature", initMigration(t), "~@later")
}
