package buckets

import (
	"os"

	"github.com/pureinsights/discovery-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
)

// NewStoreCommand creates the bucket store command.
func NewStoreCommand(d cli.Discovery) *cobra.Command {
	var data string
	store := &cobra.Command{
		Use:   "store <bucket> [configFile]",
		Short: "The command that stores buckets to Discovery Staging.",
		Long:  "",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			err = commands.CheckCredentials(d, profile, "Staging", "staging_url")
			if err != nil {
				return err
			}

			vpr := d.Config()

			stagingClient := discoveryPackage.NewStaging(vpr.GetString(profile+".staging_url"), vpr.GetString(profile+".staging_key"))
			printer := cli.GetObjectPrinter(vpr.GetString("output"))
			options := gjson.Result{}
			if cmd.Flags().Changed("data") {
				if len(args) > 1 {
					return cli.NewError(cli.ErrorExitCode, "The data flag can only have the bucket name argument.")
				}
				options = gjson.Parse(data)
			}
			if len(args) > 1 {
				jsonBytes, err := os.ReadFile(args[1])
				if err != nil {
					err = cli.NormalizeReadFileError(args[1], err)
					return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not read file %q", args[1])
				}
				options = gjson.ParseBytes(jsonBytes)
			}

			return d.StoreBucket(stagingClient.Buckets(), args[0], options, printer)
		},
		Args: cobra.RangeArgs(1, 2),
		Example: `	# Store a bucket with the JSON configuration in a file
	discovery staging bucket store --file buckets.json

	# Store a bucket with the JSON configuration in the data flag
	discovery staging bucket store --data '{"type":"staging","name":"Search bucket","labels":[],"active":true,"id":"1d81d3d5-58a2-44a5-9acf-3fc8358afe09","creationTimestamp":"2025-09-04T15:50:08Z","lastUpdatedTimestamp":"2025-09-04T15:50:08Z","config":{"action":"scroll","bucket":"blogs"},"pipeline":"9a74bf3a-eb2a-4334-b803-c92bf1bc45fe","recordPolicy":{"errorPolicy":"FATAL","timeoutPolicy":{"slice":"PT1H"},"outboundPolicy":{"idPolicy":{},"batchPolicy":{"maxCount":25,"flushAfter":"PT1M"}}}}'`,
	}

	store.Flags().StringVarP(&data, "data", "d", "", "the JSON with the configurations that will be upserted")

	return store
}
