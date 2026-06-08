package buckets

import (
	"fmt"

	"github.com/pureinsights/discovery-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewStoreCommand creates the bucket store command.
func NewStoreCommand(d cli.Discovery) *cobra.Command {
	var abortOnError bool
	var data string
	store := &cobra.Command{
		Use:   "store [<files>...]",
		Short: "The command that stores buckets to Discovery Staging.",
		Long:  fmt.Sprintf(commands.LongStore, "bucket", "Staging"),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			stagingClient := discoveryPackage.NewStaging(vpr.GetString(profile+".staging_url"), vpr.GetString(profile+".staging_key"))
			return commands.SearchStoreCommand(d, stagingClient.Buckets(), commands.StoreCommandConfig(commands.GetCommandConfig(profile, vpr.GetString("output"), "Staging", "staging_url"), abortOnError, data, args))
		},
		Example: `	# Store a bucket with the JSON configuration in a file
	discovery staging bucket store configFile.json

	# Store a bucket with the JSON configuration in the data flag
	discovery staging bucket store --data '{"name":"my-bucket","config":{"indices":[{"name":"myIndexA","fields":[{"fieldName":"ASC"}],"unique":false},{"name":"myIndexB","fields":[{"fieldName2":"DESC"}],"unique":false}],"config":{}}}'`,
	}

	store.Flags().BoolVar(&abortOnError, "abort-on-error", false, "aborts the operation if there is an error")
	store.Flags().StringVarP(&data, "data", "d", "", "the JSON with the configurations that will be upserted")

	return store
}
