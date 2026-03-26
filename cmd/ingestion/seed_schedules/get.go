package seed_schedules

import (
	"fmt"

	"github.com/pureinsights/discovery-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewGetCommand creates the seed schedule get command.
func NewGetCommand(d cli.Discovery) *cobra.Command {
	var filters []string
	get := &cobra.Command{
		Use:   "get [<seed-schedule>]",
		Short: "The command that obtains seed schedules from Discovery Ingestion.",
		Long:  fmt.Sprintf(commands.LongGetSearch, "seed schedule", "Ingestion"),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			ingestionClient := discoveryPackage.NewIngestion(vpr.GetString(profile+".ingestion_url"), vpr.GetString(profile+".ingestion_key"))
			return commands.SearchCommand(args, d, ingestionClient.SeedSchedules(), commands.GetCommandConfig(profile, vpr.GetString("output"), "Ingestion", "ingestion_url"), &filters)
		},
		Args: cobra.MaximumNArgs(1),
		Example: `	# Get seed schedule by name
	discovery ingestion seed-schedule get "my-seed-schedule"

	# Get seed schedules using filters
	discovery ingestion seed-schedule get --filter label=A:A

	# Get all seed schedules using the configuration in profile "cn"
	discovery ingestion seed-schedule get -p cn`,
	}

	get.Flags().StringArrayVarP(&filters, "filter", "f", []string{}, `apply filters in the format "filter=key:value". The available filter is:
- Label: The format is label={key}[:{value}], where the value is optional`)
	return get
}
