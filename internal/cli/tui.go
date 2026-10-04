package cli

// The `tui` command: the same cairn, laid out on a screen that stays open.
// One verb and no arguments — a board is a place you go, not a question you
// ask, so there is nothing to say about it on the command line.
//
// Named for where it is drawn and not for what it draws, which is what the two
// surfaces have in common: `tui` beside `gui`. It was `board`, and that named
// the four columns — which both surfaces draw, so it said nothing about which
// one you were opening.

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"

	"cairn/internal/engine"
	"cairn/internal/tui"
)

// openTui is how this command starts a bubbletea program. It is a variable so a
// scenario can drive the board without a terminal: what is worth testing is the
// board, and a program that needs a tty is not something a test should need.
var openTui = func(cmd *cobra.Command, model tea.Model) error {
	program := tea.NewProgram(model,
		tea.WithContext(cmd.Context()),
		tea.WithInput(cmd.InOrStdin()),
		tea.WithOutput(cmd.OutOrStdout()))
	_, err := program.Run()
	return err
}

// newTuiCmd builds the `tui` command.
func newTuiCmd() *cobra.Command {
	var icons string
	cmd := &cobra.Command{
		Use:   "tui",
		Short: "Open the board in the terminal",
		Args:  noneAtAll(),
		RunE: func(cmd *cobra.Command, _ []string) error {
			set, err := tui.ParseSet(icons)
			if err != nil {
				return err
			}
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
			// The size is asked for by bubbletea the moment the program starts,
			// so nothing here has to guess one: what is passed is only what the
			// board draws with until the first answer arrives.
			return openTui(cmd, tui.New(held, tui.GlyphsFor(set), time.Now, 0, 0))
		},
	}
	cmd.Flags().StringVar(&icons, "icons", string(tui.NerdFont), "which glyphs to draw with: nerdfont or unicode")
	return cmd
}
