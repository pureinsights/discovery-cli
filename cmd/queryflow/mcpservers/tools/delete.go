package tools

import (
	"fmt"

	"github.com/pureinsights/discovery-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewDeleteCommand creates the mcp-server tool delete command.
func NewDeleteCommand(d cli.Discovery) *cobra.Command {
	get := &cobra.Command{
		Use:   "delete tool <mcp-server> <mcp-tool>",
		Short: "The command that deletes tools in an MCP server from Discovery QueryFlow.",
		Long:  fmt.Sprintf(commands.LongDeleteSearch, "MCP tool", "QueryFlow") + LongTool,
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			err = commands.CheckCredentials(d, profile, "QueryFlow", "queryflow_url")
			if err != nil {
				return err
			}

			vpr := d.Config()

			queryflowClient := discoveryPackage.NewQueryFlow(vpr.GetString(profile+".queryflow_url"), vpr.GetString(profile+".queryflow_key")).MCPServers()
			mcpServerID, err := cli.GetEntityId(d, queryflowClient, args[0])
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the MCP server %q", args[0])
			}

			return commands.SearchDeleteCommand(args[1], d, queryflowClient.Tools(mcpServerID), commands.GetCommandConfig(profile, vpr.GetString("output"), "QueryFlow", "queryflow_url"))
		},
		Args: cobra.ExactArgs(2),
		Example: `	# Delete a tool from an MCP server by id
	discovery queryflow mcp-server tool delete my-mcp-server ea02fc14-f07b-49f2-b185-e9ceaedcb367

	# Delete a tool from an MCP server by name
	discovery queryflow mcp-server tool delete my-mcp-server my-tool`,
	}
	return get
}
