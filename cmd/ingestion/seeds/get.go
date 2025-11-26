package seeds

import (
	"fmt"

	"github.com/pureinsights/pdp-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/pdp-cli/discovery"
	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewGetCommand creates the seed get command
func NewGetCommand(d cli.Discovery) *cobra.Command {
	var filters []string
	get := &cobra.Command{
		Use:   "get",
		Short: "The command that obtains seeds from Discovery Ingestion.",
		Long:  fmt.Sprintf(commands.LongGetSearch, "seed", "Ingestion"),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			ingestionClient := discoveryPackage.NewIngestion(vpr.GetString(profile+".ingestion_url"), vpr.GetString(profile+".ingestion_key"))
			return commands.SearchCommand(args, d, ingestionClient.Seeds(), commands.GetCommandConfig(profile, vpr.GetString("output"), "Ingestion", "ingestion_url"), &filters)
		},
		Args: cobra.MaximumNArgs(1),
		Example: `	# Get seed by name
	discovery ingestion seed get "Search seed"
	{"active":true,"creationTimestamp":"2025-10-31T22:54:08Z","id":"7251d693-7382-452f-91dc-859add803a43","labels":[{"key":"A","value":"A"}],"lastUpdatedTimestamp":"2025-10-31T22:54:08Z","name":"Search seed","type":"staging"}

	# Get seeds using filters
	discovery ingestion seed get --filter label=A:A -f type=staging
	{"active":true,"creationTimestamp":"2025-10-31T22:53:54Z","id":"326a6b94-8931-4d18-b3a6-77adad14f2c0","labels":[{"key":"A","value":"A"}],"lastUpdatedTimestamp":"2025-10-31T22:53:54Z","name":"Search seed","type":"staging"}
	{"active":true,"creationTimestamp":"2025-10-31T22:54:04Z","id":"8fbd57cd-9f82-409d-8d85-98e4b9225f3a","labels":[{"key":"A","value":"A"}],"lastUpdatedTimestamp":"2025-10-31T22:54:04Z","name":"Staging store seed","type":"staging"}

	# Get all seeds using the configuration in profile "cn"
	discovery ingestion seed get -p cn
	{"active":true,"creationTimestamp":"2025-09-05T19:19:30Z","id":"026c6cf3-cba4-4d68-9806-1e534eebb99d","labels":[],"lastUpdatedTimestamp":"2025-09-05T19:19:30Z","name":"Search seed","type":"staging"}
	{"active":true,"creationTimestamp":"2025-10-31T22:54:05Z","id":"028120a6-1859-47c7-b69a-f417e54b4a4a","labels":[{"key":"A","value":"A"}],"lastUpdatedTimestamp":"2025-10-31T22:54:05Z","name":"ChatGPT seed","type":"staging"}
	{"active":true,"creationTimestamp":"2025-09-05T19:48:00Z","id":"0517a87a-86f7-4a71-bb3f-adfa0c87a269","labels":[],"lastUpdatedTimestamp":"2025-09-05T19:48:00Z","name":"Staging ingestion seed","type":"staging"}`,
	}

	get.Flags().StringArrayVarP(&filters, "filter", "f", []string{}, `Apply filters in the format "filter=key:value". The available filters are:
- Label: The format is label={key}[:{value}], where the value is optional.
- Type: The format is type={type}.`)
	return get
}
