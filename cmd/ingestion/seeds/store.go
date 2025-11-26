package seeds

import (
	"fmt"

	"github.com/pureinsights/pdp-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/pdp-cli/discovery"
	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewStoreCommand creates the seed store command
func NewStoreCommand(d cli.Discovery) *cobra.Command {
	var abortOnError bool
	var data string
	var file string
	store := &cobra.Command{
		Use:   "store",
		Short: "The command that stores seeds to Discovery Ingestion.",
		Long:  fmt.Sprintf(commands.LongStore, "seed", "Ingestion"),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			ingestionClient := discoveryPackage.NewIngestion(vpr.GetString(profile+".ingestion_url"), vpr.GetString(profile+".ingestion_key"))
			return commands.StoreCommand(d, ingestionClient.Seeds(), commands.StoreCommandConfig(commands.GetCommandConfig(profile, vpr.GetString("output"), "Ingestion", "ingestion_url"), abortOnError, data, file))
		},
		Args: cobra.NoArgs,
		Example: `	# Store a seed with the JSON configuration in a file
	discovery ingestion seed store --file seeds.json
	{"active":true,"config":{"action":"scroll","bucket":"blogs"},"creationTimestamp":"2025-09-04T15:50:08Z","id":"1d81d3d5-58a2-44a5-9acf-3fc8358afe09","labels":[],"lastUpdatedTimestamp":"2025-09-04T15:50:08Z","name":"Search seed","pipeline":"9a74bf3a-eb2a-4334-b803-c92bf1bc45fe","recordPolicy":{"errorPolicy":"FATAL","outboundPolicy":{"batchPolicy":{"flushAfter":"PT1M","maxCount":25},"idPolicy":{}},"timeoutPolicy":{"slice":"PT1H"}},"type":"staging"}
	{"code":1003,"messages":["Entity not found: 1d81d3d5-58a2-44a5-9acf-3fc8358afe00"],"status":404,"timestamp":"2025-10-31T20:32:39.832877700Z"}
	{"active":true,"config":{"action":"scroll","bucket":"blogs"},"creationTimestamp":"2025-10-31T20:32:39.855952Z","id":"d818d852-18ac-4059-8f17-37a1b649bbfd","labels":[],"lastUpdatedTimestamp":"2025-10-31T20:32:39.855952Z","name":"Search seed 2","pipeline":"9a74bf3a-eb2a-4334-b803-c92bf1bc45fe","recordPolicy":{"errorPolicy":"FATAL","outboundPolicy":{"batchPolicy":{"flushAfter":"PT1M","maxCount":25},"idPolicy":{}},"timeoutPolicy":{"slice":"PT1H"}},"type":"staging"}

	# Store a seed with the JSON configuration in the data flag
	discovery ingestion seed store --data '{"type":"staging","name":"Search seed","labels":[],"active":true,"id":"1d81d3d5-58a2-44a5-9acf-3fc8358afe09","creationTimestamp":"2025-09-04T15:50:08Z","lastUpdatedTimestamp":"2025-09-04T15:50:08Z","config":{"action":"scroll","bucket":"blogs"},"pipeline":"9a74bf3a-eb2a-4334-b803-c92bf1bc45fe","recordPolicy":{"errorPolicy":"FATAL","timeoutPolicy":{"slice":"PT1H"},"outboundPolicy":{"idPolicy":{},"batchPolicy":{"maxCount":25,"flushAfter":"PT1M"}}}}'
	{"active":true,"config":{"action":"scroll","bucket":"blogs"},"creationTimestamp":"2025-09-04T15:50:08Z","id":"1d81d3d5-58a2-44a5-9acf-3fc8358afe09","labels":[],"lastUpdatedTimestamp":"2025-09-04T15:50:08Z","name":"Search seed","pipeline":"9a74bf3a-eb2a-4334-b803-c92bf1bc45fe","recordPolicy":{"errorPolicy":"FATAL","outboundPolicy":{"batchPolicy":{"flushAfter":"PT1M","maxCount":25},"idPolicy":{}},"timeoutPolicy":{"slice":"PT1H"}},"type":"staging"}`,
	}
	store.Flags().BoolVar(&abortOnError, "abort-on-error", false, "Aborts the operation if there is an error")
	store.Flags().StringVarP(&data, "data", "d", "", "The JSON with the configurations that will be upserted")
	store.Flags().StringVarP(&file, "file", "f", "", "The path of the file that contains the JSON data")

	store.MarkFlagsOneRequired("data", "file")
	store.MarkFlagsMutuallyExclusive("data", "file")
	return store
}
