package cli

import (
	"cairn/internal/board"
	"cairn/internal/engine/workspace"
	"cairn/internal/gui/design"
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"cairn/internal/debug"
	"cairn/internal/engine"
	"cairn/internal/gui"
)

const (
	wide = 1280
	high = 840
)

type opensWindow = func(work board.Work, pages engine.Pages, look design.Skin, noted *debug.Log, width, height int, here gui.Here) error

var openGui opensWindow = func(work board.Work, pages engine.Pages, look design.Skin, noted *debug.Log, width, height int, here gui.Here) error {
	gui.Run(work, pages, look, noted, width, height, here)
	return nil
}

type windowKey struct{}

func withWindow(ctx context.Context, open opensWindow) context.Context {
	return context.WithValue(ctx, windowKey{}, open)
}

func windowFor(cmd *cobra.Command) opensWindow {
	if open, given := carried[opensWindow](cmd, windowKey{}); given {
		return open
	}
	return openGui
}

func newGuiCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "gui",
		Short: "Open the board in a window",
		Args:  noneAtAll(),
		RunE: func(cmd *cobra.Command, _ []string) error {
			return openWindow(cmd, homeOf(cmd))
		},
	}
}

func openWindow(cmd *cobra.Command, h engine.Home) error {
	held, err := engine.Hold(h, engine.Through("gui", h))
	if err != nil {
		return err
	}
	defer func() { _ = held.Close() }()
	at, err := engine.Locate(h)
	if err != nil {
		return err
	}
	worn, unworn := engine.Wearing(h)
	if unworn != "" {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "theme: %s\n", unworn)
	}
	here := gui.Here{Store: at.Store, Reach: func(to engine.Elsewhere) (workspace.Apart, error) { return engine.OpenApart(held, to) }}
	return windowFor(cmd)(held, engine.PagesFrom(h), design.Skinned(worn), notedIn(cmd.Context()), wide, high, here)
}
