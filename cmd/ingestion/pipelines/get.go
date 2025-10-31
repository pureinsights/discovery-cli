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
			return commands.SearchCommand(args, d, ingestionClient.Pipelines(), commands.GetCommandConfig(profile, vpr.GetString("output"), "Ingestion", "ingestion_url", "ingestion_key"), &filters)
		},
		Args: cobra.MaximumNArgs(1),
	}

	get.Flags().StringArrayVarP(&filters, "filter", "f", []string{}, `Apply filters in the format "filter=key:value". The available filter is:
- Label: The format is label={key}[:{value}], where the value is optional.`)
	return get
}
