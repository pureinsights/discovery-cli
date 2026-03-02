package endpoints

import (
	"fmt"

	"github.com/pureinsights/discovery-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewStoreCommand creates the endpoint store command.
func NewStoreCommand(d cli.Discovery) *cobra.Command {
	var abortOnError bool
	var data string
	store := &cobra.Command{
		Use:   "store [<files>...]",
		Short: "The command that stores endpoints to Discovery QueryFlow.",
		Long:  fmt.Sprintf(commands.LongStore, "endpoint", "QueryFlow"),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			queryflowClient := discoveryPackage.NewQueryFlow(vpr.GetString(profile+".queryflow_url"), vpr.GetString(profile+".queryflow_key"))
			return commands.SearchStoreCommand(d, queryflowClient.Endpoints(), commands.StoreCommandConfig(commands.GetCommandConfig(profile, vpr.GetString("output"), "QueryFlow", "queryflow_url"), abortOnError, data, args))
		},
		Example: `	# Store an endpoint with the JSON configuration in a file
	discovery queryflow endpoint store endpointjsonfile.json

	# Store an endpoint with the JSON configuration in the data flag
	discovery queryflow endpoint store --data '{"type":"default","uri":"/my/endpoint","httpMethod":"GET","name":"my-endpoint","pipeline":"782bfece-20a2-4382-bacb-1c9c550e2d58","timeout":"60s"}'`,
	}
	store.Flags().BoolVar(&abortOnError, "abort-on-error", false, "aborts the operation if there is an error")
	store.Flags().StringVarP(&data, "data", "d", "", "the JSON with the configurations that will be upserted")

	return store
}
