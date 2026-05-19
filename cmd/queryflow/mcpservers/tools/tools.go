package tools

import (
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewToolCommand creates the tool command.
func NewToolCommand(d cli.Discovery) *cobra.Command {
	tool := &cobra.Command{
		Use:   "tool [subcommand] [flags]",
		Short: "The command to interact with an MCP server's tools in Discovery QueryFlow.",
	}

	tool.AddCommand(NewStoreCommand(d))

	return tool
}
