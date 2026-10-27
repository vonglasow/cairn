package cli

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"cairn/internal/engine"
	"cairn/internal/mcp"
)

func newMCPCmd() *cobra.Command {
	group := &cobra.Command{
		Use:   "mcp",
		Short: "Answer a machine over MCP",
		Args:  noneAtAll(),
		RunE: func(cmd *cobra.Command, _ []string) error {
			return engine.Misread(fmt.Errorf("mcp needs a command — usage: %s", cmd.UseLine()))
		},
	}
	group.AddCommand(newMCPServeCmd())
	return group
}

func newMCPServeCmd() *cobra.Command {
	var access string
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Answer an agent over MCP",
		Args:  noneAtAll(),
		RunE: func(cmd *cobra.Command, _ []string) error {
			level, err := engine.ParseLevel(access)
			if err != nil {
				return err
			}
			keepingThePluginCurrent(cmd.ErrOrStderr())
			server := mcp.Server{Name: cmd.Root().Name(), Version: stamps().Version, Access: level, Home: home(), Noted: notedIn(cmd.Context())}
			return mcp.Serve(cmd.Context(), server, cmd.InOrStdin(), cmd.OutOrStdout())
		},
	}
	cmd.Flags().StringVar(&access, "access", "read", "how far the agent may reach: read, write or full")
	return cmd
}

func keepingThePluginCurrent(said io.Writer) {
	at, err := engine.Locate(home())
	if err != nil {
		return
	}
	rewrote, err := engine.KeepPluginCurrent(at, stamps().Version)
	if err != nil {
		_, _ = fmt.Fprintf(said, "cairn: the Claude Code plugin beside the store could not be brought to %s: %v\n", stamps().Version, err)
		return
	}
	if rewrote {
		_, _ = fmt.Fprintf(said, "cairn: the Claude Code plugin beside the store is now %s; it is read from the next session\n", stamps().Version)
	}
}
