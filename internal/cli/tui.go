package cli

import (
	"io"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"

	"cairn/internal/engine"
	"cairn/internal/tui"
)

var openTui = runTui

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

var errNotATerminal = engine.Refusing("not-a-terminal", "the board is drawn in a terminal, and this is not one")

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
			held, err := engine.Hold(home(), engine.Through("tui", home()))
			if err != nil {
				return err
			}
			defer func() { _ = held.Close() }()
			if _, err := engine.Look(held); err != nil {
				return err
			}
			return openTui(cmd, tui.New(held, tui.GlyphsFor(set), time.Now, 0, 0).Paged(engine.PagesFrom(home())).Noting(notedIn(cmd.Context())))
		},
	}
	cmd.Flags().StringVar(&icons, "icons", string(tui.NerdFont), "which glyphs to draw with: nerdfont or unicode")
	return cmd
}
