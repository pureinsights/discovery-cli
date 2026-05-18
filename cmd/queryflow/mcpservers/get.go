package mcpservers

import (
	"fmt"

	"github.com/pureinsights/discovery-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewGetCommand creates the mcp-server get command.
func NewGetCommand(d cli.Discovery) *cobra.Command {
	var filters []string
	get := &cobra.Command{
		Use:   "get [<mcp-server>]",
		Short: "The command that obtains MCP servers from Discovery QueryFlow.",
		Long:  fmt.Sprintf(commands.LongGetSearch, "MCP server", "QueryFlow"),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			queryflowClient := discoveryPackage.NewQueryFlow(vpr.GetString(profile+".queryflow_url"), vpr.GetString(profile+".queryflow_key"))
			return commands.SearchCommand(args, d, queryflowClient.MCPServers(), commands.GetCommandConfig(profile, vpr.GetString("output"), "QueryFlow", "queryflow_url"), &filters)
		},
		Args: cobra.MaximumNArgs(1),
		Example: `	# Get an MCP server by name
	discovery queryflow mcp-server get "my-mcp-server"
	
	# Get mcp-servers using filters
	discovery queryflow mcp-server get --filter label=A:B

	# Get all mcp-servers using the configuration in profile "cn"
	discovery queryflow mcp-server get -p cn`,
	}

	get.Flags().StringArrayVarP(&filters, "filter", "f", []string{}, `apply filters in the format "filter=key:value". The available filter is:
- Label: The format is label={key}[:{value}], where the value is optional`)
	return get
}
