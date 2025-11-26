package processors

import (
	"fmt"

	"github.com/pureinsights/pdp-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/pdp-cli/discovery"
	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewGetCommand creates the processor get command
func NewGetCommand(d cli.Discovery) *cobra.Command {
	var filters []string
	get := &cobra.Command{
		Use:   "get",
		Short: "The command that obtains processors from Discovery Ingestion.",
		Long:  fmt.Sprintf(commands.LongGetSearch, "processor", "Ingestion"),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			ingestionClient := discoveryPackage.NewIngestion(vpr.GetString(profile+".ingestion_url"), vpr.GetString(profile+".ingestion_key"))
			return commands.SearchCommand(args, d, ingestionClient.Processors(), commands.GetCommandConfig(profile, vpr.GetString("output"), "Ingestion", "ingestion_url"), &filters)
		},
		Args: cobra.MaximumNArgs(1),
		Example: `	# Get processor by name
	discovery ingestion processor get "MongoDB store processor"
	{"active":true,"creationTimestamp":"2025-10-30T20:07:43Z","id":"90675678-fc9f-47ec-8bab-89969dc204f0","labels":[],"lastUpdatedTimestamp":"2025-10-30T20:07:43Z","name":"MongoDB store processor","type":"mongo"}

	# Get processors using filters
	discovery ingestion processor get --filter label=A:A -f type=mongo
	{"active":true,"creationTimestamp":"2025-10-31T21:50:54Z","id":"89103d32-6007-489a-8e25-dc9a6001f8e8","labels":[{"key":"A","value":"A"}],"lastUpdatedTimestamp":"2025-10-31T21:50:54Z","name":"MongoDB processor","type":"mongo"}

	# Get all processors using the configuration in profile "cn"
	discovery ingestion processor get -p cn
	{"active":true,"creationTimestamp":"2025-08-21T21:52:02Z","id":"516d4a8a-e8ae-488c-9e37-d5746a907454","labels":[],"lastUpdatedTimestamp":"2025-08-21T21:52:02Z","name":"Template processor","type":"template"}
	{"active":true,"creationTimestamp":"2025-10-30T20:07:43Z","id":"7569f1a5-521e-4d8c-94d1-9f53ad065320","labels":[],"lastUpdatedTimestamp":"2025-10-30T20:07:43Z","name":"MongoDB store processor","type":"mongo"}`,
	}

	get.Flags().StringArrayVarP(&filters, "filter", "f", []string{}, `Apply filters in the format "filter=key:value". The available filters are:
- Label: The format is label={key}[:{value}], where the value is optional.
- Type: The format is type={type}.`)
	return get
}
