package serviceaccounts

import (
	"fmt"

	"github.com/pureinsights/discovery-cli/v2/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/v2/discovery"
	"github.com/pureinsights/discovery-cli/v2/internal/cli"
	"github.com/spf13/cobra"
)

// NewGetCommand creates the service account get command.
func NewGetCommand(d cli.Discovery) *cobra.Command {
	var filters []string
	get := &cobra.Command{
		Use:   "get [<service-account>]",
		Short: "The command that obtains service accounts from Discovery Core.",
		Long:  fmt.Sprintf(commands.LongGetSearch, "service-account", "Core"),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			coreClient := discoveryPackage.NewCore(vpr.GetString(profile+".core_url"), vpr.GetString(profile+".core_key"))
			return commands.SearchCommand(args, d, coreClient.ServiceAccounts(), commands.GetCommandConfig(profile, vpr.GetString("output"), "Core", "core_url"), &filters)
		},
		Args: cobra.MaximumNArgs(1),
		Example: `	# Get service accounts by name
	discovery core service-account get "ci"

	# Get service accounts using filters
	discovery core service-account get --filter label=A:A

	# Get all service accounts using the configuration in profile "cn"
	discovery core service-account get -p cn`,
	}

	get.Flags().StringArrayVarP(&filters, "filter", "f", []string{}, `apply filters in the format "filter=key:value". The available filters are:
- Label: The format is label={key}[:{value}], where the value is optional`)
	return get
}
