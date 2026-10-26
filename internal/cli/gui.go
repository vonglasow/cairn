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

var openGui = func(work engine.Work, pages engine.Pages, draft, sizes string, width, height int) error {
	gui.Run(work, pages, gui.DraftsAt(draft), sizesAt(sizes), width, height)
	return nil
}

func sizesAt(path string) gui.SizesKept {
	if path == "" {
		return nil
	}
	return gui.SizesAt(path)
}

func newGuiCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "gui",
		Short: "Open the board in a window",
		Args:  noneAtAll(),
		RunE: func(_ *cobra.Command, _ []string) error {
			held, err := engine.Hold(home(), engine.Through("gui", home()))
			if err != nil {
				return err
			}
			defer func() { _ = held.Close() }()
			at, err := engine.Locate(home())
			if err != nil {
				return err
			}
			sizes := ""
			if engine.KeepsSizes(home()) {
				sizes = at.Sizes()
			}
			return openGui(held, engine.PagesFrom(home()), at.Draft(), sizes, wide, high)
		},
	}
}
