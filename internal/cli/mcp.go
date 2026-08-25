package cli

// The `mcp` group: the same cairn, asked by an agent instead of by a person.
// `cairn mcp serve` starts the server, because everything else here reads
// `cairn <thing> <verb>` and a server is no reason to break that.

import (
	"fmt"

	"github.com/spf13/cobra"

	"cairn/internal/mcp"
)

// newMCPCmd builds the `mcp` group.
func newMCPCmd() *cobra.Command {
	group := &cobra.Command{
		Use:   "mcp",
		Short: "Answer a machine over MCP",
		Args:  noneAtAll(),
		// Without a RunE the group would accept a stray argument, print its help
		// and exit zero — so a typo would look like a success.
		RunE: func(cmd *cobra.Command, _ []string) error {
			return fmt.Errorf("mcp needs a command — usage: %s", cmd.UseLine())
		},
	}
	group.AddCommand(newMCPServeCmd())
	return group
}

// newMCPServeCmd builds the `serve` verb: it reads the level, stops if it does
// not understand it, and otherwise answers over the command's own standard
// input and output, under the name and version this build reports.
func newMCPServeCmd() *cobra.Command {
	var access string
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Answer an agent over MCP",
		Args:  noneAtAll(),
		RunE: func(cmd *cobra.Command, _ []string) error {
			level, err := mcp.ParseAccess(access)
			if err != nil {
				return err
			}
			path, err := storePath()
			if err != nil {
				return err
			}
			// The name comes from the tree and the version from the build, so
			// an agent and a person are told the same two things by the same
			// two authorities. Written here, it would be a third.
			server := mcp.Server{Name: cmd.Root().Name(), Version: stamps().Version, Access: level, Path: path}
			return mcp.Serve(cmd.Context(), server, cmd.InOrStdin(), cmd.OutOrStdout())
		},
	}
	cmd.Flags().StringVar(&access, "access", "write", "how far the agent may reach: read, write or full")
	return cmd
}
