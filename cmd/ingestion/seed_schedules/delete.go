package seed_schedules

import (
	"fmt"

	"github.com/pureinsights/discovery-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewDeleteCommand creates the seed schedule delete command.
func NewDeleteCommand(d cli.Discovery) *cobra.Command {
	get := &cobra.Command{
		Use:   "delete <seed-schedule>",
		Short: "The command that deletes seed schedules from Discovery Ingestion.",
		Long:  fmt.Sprintf(commands.LongDeleteSearch, "seed schedule", "Ingestion"),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			ingestionClient := discoveryPackage.NewIngestion(vpr.GetString(profile+".ingestion_url"), vpr.GetString(profile+".ingestion_key"))
			return commands.SearchDeleteCommand(args[0], d, ingestionClient.SeedSchedules(), commands.GetCommandConfig(profile, vpr.GetString("output"), "Ingestion", "ingestion_url"))
		},
		Args: cobra.ExactArgs(1),
		Example: `	# Delete a seed schedule by id
	discovery ingestion seed-schedule delete e9cec918-69a9-4053-946b-c2538a7a49be

	# Delete a seed schedule by name
	discovery ingestion seed-schedule delete "my-seed-schedule"`,
	}
	return get
}
