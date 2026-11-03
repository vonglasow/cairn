package cli

import (
	"github.com/spf13/cobra"

	"cairn/internal/debug"
	"cairn/internal/engine"
	"cairn/internal/gui"
)

const (
	wide = 1280
	high = 840
)

var openGui = func(work engine.Work, pages engine.Pages, draft, sizes string, noted *debug.Log, width, height int) error {
	gui.Run(work, pages, gui.DraftsAt(draft), sizesAt(sizes), noted, width, height)
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
		RunE: func(cmd *cobra.Command, _ []string) error {
			return openWindow(cmd, home())
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
	sizes := ""
	if engine.KeepsSizes(h) {
		sizes = at.Sizes()
	}
	return openGui(held, engine.PagesFrom(h), at.Draft(), sizes, notedIn(cmd.Context()), wide, high)
}
