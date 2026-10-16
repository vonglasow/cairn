package cli

import (
	"github.com/spf13/cobra"

	"cairn/internal/engine"
	"cairn/internal/gui"
)

const (
	wide = 1280
	high = 840
)

var openGui = func(work engine.Work, pages engine.Pages, width, height int) error {
	gui.Run(work, pages, width, height)
	return nil
}

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
			return openGui(held, engine.PagesFrom(home()), wide, high)
		},
	}
}
