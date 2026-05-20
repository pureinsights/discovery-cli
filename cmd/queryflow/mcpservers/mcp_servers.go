package mcpservers

import (
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewMCPServerCommand creates the mcp-server command.
func NewMCPServerCommand(d cli.Discovery) *cobra.Command {
	mcpServer := &cobra.Command{
		Use:   "mcp-server [subcommand] [flags]",
		Short: "The command to interact with Discovery QueryFlow's MCP servers.",
	}

	mcpServer.AddCommand(NewGetCommand(d))

	return mcpServer
}
