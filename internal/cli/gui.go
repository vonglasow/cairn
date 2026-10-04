package cli

// The `gui` command: the same cairn in a window a mouse can reach. One verb and
// no arguments — a window is a place you go, not a question you ask, so there
// is nothing to say about it on the command line.

import (
	"github.com/spf13/cobra"

	"cairn/internal/engine"
	"cairn/internal/gui"
	"cairn/internal/reading"
)

// The window it opens on, before anybody has dragged it anywhere. Wide enough
// for four columns and the band, and no wider: a window that opened filling the
// screen would be a window that decided something for somebody.
const (
	wide = 1280
	high = 840
)

// openGui is how this command opens a window. A variable so a scenario can
// drive the command without a display: what is worth testing is that the store
// is found and handed over, and a window that needs a screen is not something a
// test should need.
var openGui = func(work reading.Work, width, height int) error {
	return gui.Run(work, width, height)
}

// newGuiCmd builds the `gui` command.
func newGuiCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "gui",
		Short: "Open the board in a window",
		Args:  noneAtAll(),
		RunE: func(_ *cobra.Command, _ []string) error {
			at, err := places()
			if err != nil {
				return err
			}
			held, err := engine.OpenAt(at)
			if err != nil {
				// Told, because this command holds the store itself rather
				// than asking through an operation: without it a store behind
				// this cairn says so and leaves somebody to find out what to
				// type about it.
				return engine.Told(err)
			}
			defer func() { _ = held.Close() }()
			return openGui(held, wide, high)
		},
	}
}
