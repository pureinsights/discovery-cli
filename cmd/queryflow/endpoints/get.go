package endpoints

import (
	"fmt"

	"github.com/pureinsights/pdp-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/pdp-cli/discovery"
	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewGetCommand creates the endpoint get command
func NewGetCommand(d cli.Discovery) *cobra.Command {
	var filters []string
	get := &cobra.Command{
		Use:   "get",
		Short: "The command that obtains endpoints from Discovery QueryFlow.",
		Long:  fmt.Sprintf(commands.LongGetSearch, "endpoint", "QueryFlow"),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			queryflowClient := discoveryPackage.NewQueryFlow(vpr.GetString(profile+".queryflow_url"), vpr.GetString(profile+".queryflow_key"))
			return commands.SearchCommand(args, d, queryflowClient.Endpoints(), commands.GetCommandConfig(profile, vpr.GetString("output"), "QueryFlow", "queryflow_url"), &filters)
		},
		Args: cobra.MaximumNArgs(1),
	}

	get.Flags().StringArrayVarP(&filters, "filter", "f", []string{}, `Apply filters in the format "filter=key:value". The available filters are:
- Label: The format is label={key}[:{value}], where the value is optional.
- Type: The format is type={type}.`)
	return get
}
