package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"cairn/internal/build"
	"cairn/internal/engine"
)

func newUpgradeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "upgrade",
		Short: "Put the newest release of cairn in the place of this one",
		Long: "Take the newest release of cairn from GitHub, check its archive against the release's SHA256SUMS, and put its binary in the place of the cairn that was run. " +
			"Whatever runs cairn from there runs the new one: the terminal, the Mac application, and a cairn mcp serve already running, at its next call. " +
			"A cairn already the newest is left as it is; one built from its source, or installed by Homebrew, is left to whoever keeps it.",
		Example: "  cairn upgrade",
		Args:    noneAtAll(),
		RunE: func(cmd *cobra.Command, _ []string) error {
			running, err := build.Running()
			if err != nil {
				return err
			}
			b := stamps()
			upgraded, err := engine.Upgrade(b.Version, running, b.Platform)
			if err != nil {
				return err
			}
			_, err = fmt.Fprint(cmd.OutOrStdout(), upgraded)
			return err
		},
	}
}
