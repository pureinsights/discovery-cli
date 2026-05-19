package tools

import (
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

const (
	// LongTool is the message that explains that the first argument must be the MCP server.
	LongTool string = " The first argument of this command must be the name or UUID of the MCP server that will contain the tool."
)

// NewToolCommand creates the tool command.
func NewToolCommand(d cli.Discovery) *cobra.Command {
	tool := &cobra.Command{
		Use:   "tool [subcommand] [flags]",
		Short: "The command to interact with an MCP server's tools in Discovery QueryFlow.",
	}

	tool.AddCommand(NewGetCommand(d))
	tool.AddCommand(NewStoreCommand(d))

	return tool
}
