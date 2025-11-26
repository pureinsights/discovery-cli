package pipelines

import (
	"fmt"

	"github.com/pureinsights/pdp-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/pdp-cli/discovery"
	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewGetCommand creates the pipeline get command
func NewGetCommand(d cli.Discovery) *cobra.Command {
	var filters []string
	get := &cobra.Command{
		Use:   "get",
		Short: "The command that obtains pipelines from Discovery Ingestion.",
		Long:  fmt.Sprintf(commands.LongGetSearch, "pipeline", "Ingestion"),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			ingestionClient := discoveryPackage.NewIngestion(vpr.GetString(profile+".ingestion_url"), vpr.GetString(profile+".ingestion_key"))
			return commands.SearchCommand(args, d, ingestionClient.Pipelines(), commands.GetCommandConfig(profile, vpr.GetString("output"), "Ingestion", "ingestion_url"), &filters)
		},
		Args: cobra.MaximumNArgs(1),
		Example: `	# Get pipeline by name
	discovery ingestion pipeline get "Search pipeline"
	{"active":true,"creationTimestamp":"2025-10-31T22:06:27Z","id":"8d9560b3-631b-490b-994f-4708d9880e3b","labels":[{"key":"A","value":"A"}],"lastUpdatedTimestamp":"2025-10-31T22:06:27Z","name":"Search pipeline"}

	# Get pipelines using filters
	discovery ingestion pipeline get --filter label=A:A
	{"active":true,"creationTimestamp":"2025-10-31T22:06:27Z","id":"8d9560b3-631b-490b-994f-4708d9880e3b","labels":[{"key":"A","value":"A"}],"lastUpdatedTimestamp":"2025-10-31T22:06:27Z","name":"Search pipeline"}
	{"active":true,"creationTimestamp":"2025-10-31T22:07:00Z","id":"e15ca96b-3d42-4ab9-be16-3c6d6713b04e","labels":[{"key":"A","value":"A"}],"lastUpdatedTimestamp":"2025-10-31T22:07:00Z","name":"Search pipeline 2"}

	# Get all pipelines using the configuration in profile "cn"
	discovery ingestion pipeline get -p cn
	{"active":true,"creationTimestamp":"2025-10-31T22:07:02Z","id":"04536687-f083-4353-8ecc-b7348e14b748","labels":[{"key":"A","value":"A"}],"lastUpdatedTimestamp":"2025-10-31T22:07:02Z","name":"Search pipeline"}
	{"active":true,"creationTimestamp":"2025-10-31T19:41:16Z","id":"0d3f476d-9003-4fc8-b9a9-8ba6ebf9445b","labels":[],"lastUpdatedTimestamp":"2025-10-31T19:41:16Z","name":"Mongo Ingestion pipeline"}
	{"active":true,"creationTimestamp":"2025-10-31T22:07:00Z","id":"22b1f0fe-d7c1-476f-a609-1a12ee97655f","labels":[{"key":"A","value":"A"}],"lastUpdatedTimestamp":"2025-10-31T22:07:00Z","name":"Generate embedddings pipeline"}`,
	}

	get.Flags().StringArrayVarP(&filters, "filter", "f", []string{}, `Apply filters in the format "filter=key:value". The available filter is:
- Label: The format is label={key}[:{value}], where the value is optional.`)
	return get
}
