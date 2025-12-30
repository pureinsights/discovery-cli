package pipelines

import (
	"fmt"

	"github.com/pureinsights/discovery-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewStoreCommand creates the pipeline store command.
func NewStoreCommand(d cli.Discovery) *cobra.Command {
	var abortOnError bool
	var data string
	store := &cobra.Command{
		Use:   "store",
		Short: "The command that stores pipelines to Discovery Ingestion.",
		Long:  fmt.Sprintf(commands.LongStore, "pipeline", "Ingestion"),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			ingestionClient := discoveryPackage.NewIngestion(vpr.GetString(profile+".ingestion_url"), vpr.GetString(profile+".ingestion_key"))
			return commands.StoreCommand(d, ingestionClient.Pipelines(), commands.StoreCommandConfig(commands.GetCommandConfig(profile, vpr.GetString("output"), "Ingestion", "ingestion_url"), abortOnError, data, args))
		},
		Args: cobra.NoArgs,
		Example: `	# Store a pipeline with the JSON configuration in a file
	discovery ingestion pipeline store pipelines.json

	# Store a pipeline with the JSON configuration in the data flag
	discovery ingestion pipeline store --data '{"name":"my-pipeline","labels":[],"active":true,"id":"36f8ce72-f23d-4768-91e8-58693ff1b272","creationTimestamp":"2025-10-31T19:41:13Z","lastUpdatedTimestamp":"2025-10-31T19:41:13Z","initialState":"ingestionState","states":{"ingestionState":{"type":"processor","processors":[{"id":"516d4a8a-e8ae-488c-9e37-d5746a907454","outputField":"header","active":true},{"id":"aa0186f1-746f-4b20-b1b0-313bd79e78b8","active":true}]}},"recordPolicy":{"idPolicy":{},"retryPolicy":{"active":true,"maxRetries":3},"timeoutPolicy":{"record":"PT1M"},"errorPolicy":"FAIL","outboundPolicy":{"batchPolicy":{"maxCount":25,"flushAfter":"PT1M"}}}}'`,
	}
	store.Flags().BoolVar(&abortOnError, "abort-on-error", false, "aborts the operation if there is an error")
	store.Flags().StringVarP(&data, "data", "d", "", "the JSON with the configurations that will be upserted")

	return store
}
