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
		Use:   "store [files]",
		Short: "The command that stores endpoints to Discovery QueryFlow.",
		Long:  fmt.Sprintf(commands.LongStore, "endpoint", "QueryFlow"),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			queryflowClient := discoveryPackage.NewQueryFlow(vpr.GetString(profile+".queryflow_url"), vpr.GetString(profile+".queryflow_key"))
			return commands.StoreCommand(d, queryflowClient.Endpoints(), commands.StoreCommandConfig(commands.GetCommandConfig(profile, vpr.GetString("output"), "QueryFlow", "queryflow_url"), abortOnError, data, args))
		},
		Example: `	# Store an endpoint with the JSON configuration in a file
	discovery queryflow endpoint store endpointjsonfile.json

	# Store an endpoint with the JSON configuration in the data flag
	discovery queryflow endpoint store --data '{"type":"default","name":"my-endpoint","labels":[{"key":"A","value":"B"}],"active":true,"id":"cf56470f-0ab4-4754-b05c-f760669315af","creationTimestamp":"2025-11-20T00:10:53Z","lastUpdatedTimestamp":"2025-11-20T00:10:53Z","httpMethod":"GET","uri":"/wikis-search","timeout":"PT1H","initialState":"searchState","states":{"searchState":{"type":"processor","processors":[{"id":"b5c25cd3-e7c9-4fd2-b7e6-2bcf6e2caf89","continueOnError":false,"active":true},{"id":"a5ee116b-bd95-474e-9d50-db7be988b196","continueOnError":false,"active":true},{"id":"86e7f920-a4e4-4b64-be84-5437a7673db8","continueOnError":false,"active":true},{"id":"8a399b1c-95fc-406c-a220-7d321aaa7b0e","outputField":"answer","continueOnError":false,"active":true}],"mode":{"type":"group"},"next":"responseState"},"responseState":{"type":"response","statusCode":200,"body":{"answer":"#{ data('/answer/choices/0/message/content') }"}}}}'`,
	}
	store.Flags().BoolVar(&abortOnError, "abort-on-error", false, "aborts the operation if there is an error")
	store.Flags().StringVarP(&data, "data", "d", "", "the JSON with the configurations that will be upserted")

	return store
}
