package processors

import (
	"fmt"

	"github.com/pureinsights/discovery-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewDeleteCommand creates the processor delete command.
func NewDeleteCommand(d cli.Discovery) *cobra.Command {
	get := &cobra.Command{
		Use:   "delete",
		Short: "The command that deletes processors from Discovery Ingestion.",
		Long:  fmt.Sprintf(commands.LongDeleteNoNames, "processor", "Ingestion"),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			ingestionClient := discoveryPackage.NewIngestion(vpr.GetString(profile+".ingestion_url"), vpr.GetString(profile+".ingestion_key"))
			return commands.SearchDeleteCommand(args[0], d, ingestionClient.Processors(), commands.GetCommandConfig(profile, vpr.GetString("output"), "Ingestion", "ingestion_url"))
		},
		Args: cobra.ExactArgs(1),
		Example: `	# Delete a processor by id
	discovery ingestion processor delete 83a009d5-5d2f-481c-b8bf-f96d3a35c240
	{"acknowledged":true}

	# Delete a processor by name
	discovery ingestion processor delete my-processor"
	{"acknowledged":true}`,
	}
	return get
}
