package seed_schedules

import (
	"fmt"

	"github.com/pureinsights/discovery-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewStoreCommand creates the seed schedule store command.
func NewStoreCommand(d cli.Discovery) *cobra.Command {
	var abortOnError bool
	var data string
	store := &cobra.Command{
		Use:   "store [<files>...]",
		Short: "The command that stores seed schedules to Discovery Ingestion.",
		Long:  fmt.Sprintf(commands.LongStore, "seed schedule", "Ingestion"),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			ingestionClient := discoveryPackage.NewIngestion(vpr.GetString(profile+".ingestion_url"), vpr.GetString(profile+".ingestion_key"))
			return commands.SearchStoreCommand(d, ingestionClient.SeedSchedules(), commands.StoreCommandConfig(commands.GetCommandConfig(profile, vpr.GetString("output"), "Ingestion", "ingestion_url"), abortOnError, data, args))
		},
		Example: `	# Store a seed schedule with the JSON configuration in a file
	discovery ingestion seed-schedule store seed-schedules.json

	# Store a seed schedule with the JSON configuration in the data flag
	discovery ingestion seed-schedule store --data '{"name": "my-seed-schedule","expression": "0 0 * * *","properties": {"some-property": "a"},"seed": "ac7c5765-bef6-42cc-b519-c75df51ebf3b","scanType": "INCREMENTAL"}'`,
	}
	store.Flags().BoolVar(&abortOnError, "abort-on-error", false, "aborts the operation if there is an error")
	store.Flags().StringVarP(&data, "data", "d", "", "the JSON with the configurations that will be upserted")

	return store
}