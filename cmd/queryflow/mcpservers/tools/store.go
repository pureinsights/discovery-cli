package tools

import (
	"fmt"

	"github.com/pureinsights/discovery-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewStoreCommand creates the mcp-server store command.
func NewStoreCommand(d cli.Discovery) *cobra.Command {
	var abortOnError bool
	var data string
	store := &cobra.Command{
		Use:   "store [<files>...]",
		Short: "The command that stores tools in an MCP Server of Discovery QueryFlow.",
		Long:  fmt.Sprintf(commands.LongStore, "MCP tool", "QueryFlow") + LongTool,
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

			return commands.SearchStoreCommand(d, queryflowClient.Tools(mcpServerID), commands.StoreCommandConfig(commands.GetCommandConfig(profile, vpr.GetString("output"), "QueryFlow", "queryflow_url"), abortOnError, data, args[1:]))
		},
		Args: cobra.MinimumNArgs(1),
		Example: `	# Store an MCP server's tool with the JSON configuration in a file
	discovery queryflow mcp-server tool store my-mcp-server mcp-server-jsonfile.json

	# Store an MCP server's tool with the JSON configuration in the data flag
	discovery queryflow mcp-server tool store my-mcp-server --data '{"name":"my-mcp-server-tool","inputSchema":{"type":"object","properties":{"name":{"type":"string"},"age":{"type":"integer","minimum":0}},"required":["name"]},"pipeline":"4b558077-cb0f-4e1c-ab6b-ed96870529e4","timeout":"60s"}'`,
	}

	store.Flags().BoolVar(&abortOnError, "abort-on-error", false, "aborts the operation if there is an error")
	store.Flags().StringVarP(&data, "data", "d", "", "the JSON with the configurations that will be upserted")

	return store
}
