package tools

import (
	"fmt"

	"github.com/pureinsights/discovery-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewGetCommand creates the mcp-server tool get command.
func NewGetCommand(d cli.Discovery) *cobra.Command {
	var filters []string
	get := &cobra.Command{
		Use:   "get <mcp-server> [<tool>]",
		Short: "The command that obtains tools from MCP servers in Discovery QueryFlow.",
		Long:  fmt.Sprintf(commands.LongGetSearch, "MCP tool", "QueryFlow") + LongTool,
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

			return commands.SearchCommand(args[1:], d, queryflowClient.Tools(mcpServerID), commands.GetCommandConfig(profile, vpr.GetString("output"), "QueryFlow", "queryflow_url"), &filters)
		},
		Args: cobra.RangeArgs(1, 2),
		Example: `	# Get an MCP server's tool by name
	discovery queryflow mcp-server tool get my-mcp-server my-tool
	
	# Get the tools of an MCP server using filters
	discovery queryflow mcp-server tool get my-mcp-server --filter label=A:B

	# Get all the tools of an MCP server using the configuration in profile "cn"
	discovery queryflow mcp-server tool get my-mcp-server -p cn`,
	}

	get.Flags().StringArrayVarP(&filters, "filter", "f", []string{}, `apply filters in the format "filter=key:value". The available filter is:
- Label: The format is label={key}[:{value}], where the value is optional`)
	return get
}
