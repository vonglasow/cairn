package cli

// The `gui` command: the same cairn in a window a mouse can reach. One verb and
// no arguments — a window is a place you go, not a question you ask, so there
// is nothing to say about it on the command line.

import (
	"github.com/spf13/cobra"

	"cairn/internal/engine"
	"cairn/internal/gui"
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
var openGui = func(work engine.Work, width, height int) error {
	gui.Run(work, width, height)
	return nil
}

// newGuiCmd builds the `gui` command.
func newGuiCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "gui",
		Short: "Open the board in a window",
		Args:  noneAtAll(),
		RunE: func(_ *cobra.Command, _ []string) error {
			held, err := engine.Hold(home())
			if err != nil {
				return err
			}
			defer func() { _ = held.Close() }()
			return openGui(held, wide, high)
		},
	}
}
