package pipelines

import (
	"fmt"

	"github.com/pureinsights/pdp-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/pdp-cli/discovery"
	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewStoreCommand creates the pipeline store command
func NewStoreCommand(d cli.Discovery) *cobra.Command {
	var abortOnError bool
	var data string
	var file string
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
			return commands.StoreCommand(d, ingestionClient.Pipelines(), commands.StoreCommandConfig(commands.GetCommandConfig(profile, vpr.GetString("output"), "Ingestion", "ingestion_url"), abortOnError, data, file))
		},
		Args: cobra.NoArgs,
		Example: `	# Store a pipeline with the JSON configuration in a file
	discovery ingestion pipeline store --file pipelines.json
	{"active":true,"creationTimestamp":"2025-10-31T19:41:13Z","id":"36f8ce72-f23d-4768-91e8-58693ff1b272","initialState":"ingestionState","labels":[],"lastUpdatedTimestamp":"2025-10-31T19:54:23Z","name":"Search pipeline","recordPolicy":{"errorPolicy":"FAIL","idPolicy":{},"outboundPolicy":{"batchPolicy":{"flushAfter":"PT1M","maxCount":25},"mode":"INLINE","splitPolicy":{"children":{"idPolicy":{},"snapshotPolicy":{}},"source":{"snapshotPolicy":{}}}},"retryPolicy":{"active":true,"maxRetries":3},"timeoutPolicy":{"record":"PT1M"}},"states":{"ingestionState":{"processors":[{"active":true,"id":"516d4a8a-e8ae-488c-9e37-d5746a907454","outputField":"header"},{"active":true,"id":"aa0186f1-746f-4b20-b1b0-313bd79e78b8"}],"type":"processor"}}}
	{"code":1003,"messages":["Entity not found: 5888b852-d7d4-4761-9058-738b2ad1b5c9"],"status":404,"timestamp":"2025-10-31T19:55:34.723693100Z"}
	{"active":true,"creationTimestamp":"2025-10-31T19:55:34.758757Z","id":"bfb882a7-59e6-4cd6-afe4-7732163637f1","initialState":"ingestionState","labels":[],"lastUpdatedTimestamp":"2025-10-31T19:55:34.758757Z","name":"Search pipeline 2","recordPolicy":{"errorPolicy":"FAIL","idPolicy":{},"outboundPolicy":{"batchPolicy":{"flushAfter":"PT1M","maxCount":25},"mode":"INLINE","splitPolicy":{"children":{"idPolicy":{},"snapshotPolicy":{}},"source":{"snapshotPolicy":{}}}},"retryPolicy":{"active":true,"maxRetries":3},"timeoutPolicy":{"record":"PT1M"}},"states":{"ingestionState":{"processors":[{"active":true,"id":"516d4a8a-e8ae-488c-9e37-d5746a907454","outputField":"header"},{"active":true,"id":"aa0186f1-746f-4b20-b1b0-313bd79e78b8"}],"type":"processor"}}}

	# Store a pipeline with the JSON configuration in the data flag
	discovery ingestion pipeline store --data '{"name":"Search pipeline","labels":[],"active":true,"id":"36f8ce72-f23d-4768-91e8-58693ff1b272","creationTimestamp":"2025-10-31T19:41:13Z","lastUpdatedTimestamp":"2025-10-31T19:41:13Z","initialState":"ingestionState","states":{"ingestionState":{"type":"processor","processors":[{"id":"516d4a8a-e8ae-488c-9e37-d5746a907454","outputField":"header","active":true},{"id":"aa0186f1-746f-4b20-b1b0-313bd79e78b8","active":true}]}},"recordPolicy":{"idPolicy":{},"retryPolicy":{"active":true,"maxRetries":3},"timeoutPolicy":{"record":"PT1M"},"errorPolicy":"FAIL","outboundPolicy":{"batchPolicy":{"maxCount":25,"flushAfter":"PT1M"}}}}'
	{"active":true,"creationTimestamp":"2025-10-31T19:41:13Z","id":"36f8ce72-f23d-4768-91e8-58693ff1b272","initialState":"ingestionState","labels":[],"lastUpdatedTimestamp":"2025-10-31T19:54:23Z","name":"Search pipeline","recordPolicy":{"errorPolicy":"FAIL","idPolicy":{},"outboundPolicy":{"batchPolicy":{"flushAfter":"PT1M","maxCount":25},"mode":"INLINE","splitPolicy":{"children":{"idPolicy":{},"snapshotPolicy":{}},"source":{"snapshotPolicy":{}}}},"retryPolicy":{"active":true,"maxRetries":3},"timeoutPolicy":{"record":"PT1M"}},"states":{"ingestionState":{"processors":[{"active":true,"id":"516d4a8a-e8ae-488c-9e37-d5746a907454","outputField":"header"},{"active":true,"id":"aa0186f1-746f-4b20-b1b0-313bd79e78b8"}],"type":"processor"}}}`,
	}
	store.Flags().BoolVar(&abortOnError, "abort-on-error", false, "Aborts the operation if there is an error")
	store.Flags().StringVarP(&data, "data", "d", "", "The JSON with the configurations that will be upserted")
	store.Flags().StringVarP(&file, "file", "f", "", "The path the file that contains the JSON data")

	store.MarkFlagsOneRequired("data", "file")
	store.MarkFlagsMutuallyExclusive("data", "file")
	return store
}
