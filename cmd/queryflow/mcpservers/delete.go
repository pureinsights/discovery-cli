package mcpservers

import (
	"fmt"

	"github.com/pureinsights/discovery-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewDeleteCommand creates the mcp-server delete command.
func NewDeleteCommand(d cli.Discovery) *cobra.Command {
	get := &cobra.Command{
		Use:   "delete <mcp-server>",
		Short: "The command that deletes mcp-servers from Discovery QueryFlow.",
		Long:  fmt.Sprintf(commands.LongDeleteSearch, "MCP server", "QueryFlow"),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			queryflowClient := discoveryPackage.NewQueryFlow(vpr.GetString(profile+".queryflow_url"), vpr.GetString(profile+".queryflow_key"))
			return commands.SearchDeleteCommand(args[0], d, queryflowClient.MCPServers(), commands.GetCommandConfig(profile, vpr.GetString("output"), "QueryFlow", "queryflow_url"))
		},
		Args: cobra.ExactArgs(1),
		Example: `	# Delete an MCP server by id
	discovery queryflow mcp-server delete ea02fc14-f07b-49f2-b185-e9ceaedcb367

	# Delete an MCP server by name
	discovery queryflow mcp-server delete my-mcp-server`,
	}
	return get
}
