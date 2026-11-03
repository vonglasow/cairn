package cli

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"cairn/internal/config"
	"cairn/internal/engine"
	"cairn/internal/tui"
)

var errOneBoard = errors.New("--gui or --tui, not both")

func newFixturesCmd() *cobra.Command {
	var fresh, window, terminal bool
	var from string
	cmd := &cobra.Command{
		Use:   "fixtures [folder]",
		Short: "Make a store full of believable work, and open a board on it",
		Long: "Make a store holding every case a board draws, in a folder of its own — " + engine.FixturesFolder() +
			" when none is named, never the store in use. A folder it made before is opened as it was left; --new makes it again. A folder it did not make is left alone.",
		Example: "cairn fixtures --gui\ncairn fixtures ~/try --new --tui",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if window && terminal {
				return engine.Misread(errOneBoard)
			}
			folder := engine.FixturesFolder()
			if len(args) == 1 {
				folder = args[0]
			}
			folder, err := filepath.Abs(folder)
			if err != nil {
				return err
			}
			made, err := engine.MakeFixtures(folder, from, fresh)
			if err != nil {
				return err
			}
			said := "opened as they were left"
			if made.Made {
				said = "made"
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "fixtures %s in %s\nstore  %s\n\n%s=%s %s=%s cairn gui\n",
				said, folder, made.Store, config.StoreVariable, made.Store, config.NotesVariable, made.Notes)
			at := withPlaces(home(), made)
			if window {
				return openWindow(cmd, at)
			}
			if terminal {
				return openTerminal(cmd, at, string(tui.NerdFont))
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&fresh, "new", false, "make the fixtures again from nothing, in a folder made before")
	cmd.Flags().BoolVar(&window, "gui", false, "open the window on them")
	cmd.Flags().BoolVar(&terminal, "tui", false, "open the board in the terminal on them")
	cmd.Flags().StringVar(&from, "from", "", "a file written by task export to fill the store with instead")
	return cmd
}

func withPlaces(h engine.Home, made engine.Fixtures) engine.Home {
	lookup := h.Lookup
	h.Lookup = func(key string) (string, bool) {
		switch key {
		case config.StoreVariable:
			return made.Store, true
		case config.NotesVariable:
			return made.Notes, true
		}
		return lookup(key)
	}
	return h
}
