package endpoints

import (
	"fmt"

	"github.com/pureinsights/discovery-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewDeleteCommand creates the endpoint delete command.
func NewDeleteCommand(d cli.Discovery) *cobra.Command {
	get := &cobra.Command{
		Use:   "delete",
		Short: "The command that deletes endpoints from Discovery QueryFlow.",
		Long:  fmt.Sprintf(commands.LongDeleteSearch, "endpoint", "QueryFlow"),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			queryflowClient := discoveryPackage.NewQueryFlow(vpr.GetString(profile+".queryflow_url"), vpr.GetString(profile+".queryflow_key"))
			return commands.SearchDeleteCommand(args[0], d, queryflowClient.Endpoints(), commands.GetCommandConfig(profile, vpr.GetString("output"), "QueryFlow", "queryflow_url"))
		},
		Args: cobra.ExactArgs(1),
		Example: `	# Delete an endpoint by id
	discovery queryflow endpoint delete ea02fc14-f07b-49f2-b185-e9ceaedcb367

	# Delete an endpoint by name
	discovery queryflow endpoint delete my-endpoint`,
	}
	return get
}
