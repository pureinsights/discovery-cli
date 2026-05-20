package mcpservers

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
		Short: "The command that stores MCP servers to Discovery QueryFlow.",
		Long:  fmt.Sprintf(commands.LongStore, "MCP server", "QueryFlow"),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			queryflowClient := discoveryPackage.NewQueryFlow(vpr.GetString(profile+".queryflow_url"), vpr.GetString(profile+".queryflow_key"))
			return commands.SearchStoreCommand(d, queryflowClient.MCPServers(), commands.StoreCommandConfig(commands.GetCommandConfig(profile, vpr.GetString("output"), "QueryFlow", "queryflow_url"), abortOnError, data, args))
		},
		Example: `	# Store an MCP server with the JSON configuration in a file
	discovery queryflow mcp-server store mcp-server-jsonfile.json

	# Store an MCP server with the JSON configuration in the data flag
	discovery queryflow mcp-server store --data '{"uri":"/my/mcp/server","name":"My MCP Server","pipeline":"4b558077-cb0f-4e1c-ab6b-ed96870529e4","capabilities":{"logging":{},"tools":{}},"serverInfo":{"name":"mcp-server","version":"1.0"},"requestTimeout":"60s","expireAfter":"1h","labels":[{"key":"A","value":"A"}]}'`,
	}
	store.Flags().BoolVar(&abortOnError, "abort-on-error", false, "aborts the operation if there is an error")
	store.Flags().StringVarP(&data, "data", "d", "", "the JSON with the configurations that will be upserted")

	return store
}
