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
	"io"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"

	"cairn/internal/engine"
	"cairn/internal/tui"
)

// openTui is how this command starts a bubbletea program. It is a variable so a
// scenario can drive the board without a terminal: what is worth testing is the
// board, and a program that needs a tty is not something a test should need.
var openTui = runTui

// runTui starts the board on the command's own streams, once they are a
// terminal.
func runTui(cmd *cobra.Command, model tea.Model) error {
	if err := needsATerminal(cmd.InOrStdin(), cmd.OutOrStdout()); err != nil {
		return err
	}
	program := tea.NewProgram(model,
		tea.WithContext(cmd.Context()),
		tea.WithInput(cmd.InOrStdin()),
		tea.WithOutput(cmd.OutOrStdout()))
	_, err := program.Run()
	return err
}

// errNotATerminal is the board asked for where it cannot be drawn.
var errNotATerminal = engine.Refusing("not-a-terminal", "the board is drawn in a terminal, and this is not one")

// needsATerminal refuses streams that are not a terminal: a board drawn into
// a file or a pipe is a screenful of escapes nobody can read, and a program
// waiting for keys on a closed input waits for ever.
func needsATerminal(in io.Reader, out io.Writer) error {
	for _, stream := range []any{in, out} {
		file, ok := stream.(*os.File)
		if !ok {
			return errNotATerminal
		}
		info, err := file.Stat()
		if err != nil || info.Mode()&os.ModeCharDevice == 0 {
			return errNotATerminal
		}
	}
	return nil
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
			held, err := engine.Hold(home())
			if err != nil {
				return err
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
