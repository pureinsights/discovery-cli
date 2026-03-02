package pipelines

import (
	"fmt"

	"github.com/pureinsights/discovery-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewStoreCommand creates the pipeline store command
func NewStoreCommand(d cli.Discovery) *cobra.Command {
	var abortOnError bool
	var data string
	store := &cobra.Command{
		Use:   "store [<files>...]",
		Short: "The command that stores pipelines to Discovery QueryFlow.",
		Long:  fmt.Sprintf(commands.LongStore, "pipeline", "QueryFlow"),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			queryflowClient := discoveryPackage.NewQueryFlow(vpr.GetString(profile+".queryflow_url"), vpr.GetString(profile+".queryflow_key"))
			return commands.SearchStoreCommand(d, queryflowClient.Pipelines(), commands.StoreCommandConfig(commands.GetCommandConfig(profile, vpr.GetString("output"), "QueryFlow", "queryflow_url"), abortOnError, data, args))
		},
		Example: `	# Store a pipeline with the JSON configuration in a file
	discovery queryflow pipeline store pipelines.json

	# Store a pipeline with the JSON configuration in the data flag
	discovery queryflow pipeline store --data '{"name":"my-pipeline","initialState":"searchState","states":{"searchState":{"type":"processor","processors":[{"id":"38c35b42-56c2-42b3-85c5-b6dcd10b360b"},{"id":"4048e82c-efe9-437f-bfb1-e141e7335a53"}],"next":"responseState"},"responseState":{"type":"message","statusCode":200,"body":{"answer":"#{ data('/answer/choices/0/message/content') }"}}}}'`,
	}
	store.Flags().BoolVar(&abortOnError, "abort-on-error", false, "aborts the operation if there is an error")
	store.Flags().StringVarP(&data, "data", "d", "", "the JSON with the configurations that will be upserted")

	return store
}
