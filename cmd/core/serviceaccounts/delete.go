package serviceaccounts

import (
	"fmt"

	"github.com/pureinsights/discovery-cli/v2/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/v2/discovery"
	"github.com/pureinsights/discovery-cli/v2/internal/cli"
	"github.com/spf13/cobra"
)

// NewDeleteCommand creates the service account delete command.
func NewDeleteCommand(d cli.Discovery) *cobra.Command {
	deleteCmd := &cobra.Command{
		Use:   "delete <service-account>",
		Short: "The command that deletes service accounts from Discovery Core.",
		Long:  fmt.Sprintf(commands.LongDeleteSearch, "service-account", "Core"),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			coreClient := discoveryPackage.NewCore(vpr.GetString(profile+".core_url"), vpr.GetString(profile+".core_key"))
			return commands.SearchDeleteCommand(args[0], d, coreClient.ServiceAccounts(), commands.GetCommandConfig(profile, vpr.GetString("output"), "Core", "core_url"))
		},
		Args: cobra.ExactArgs(1),
		Example: `	# Delete a service account by id
	discovery core service-account delete 3d51beef-8b90-40aa-84b5-033241dc6239

	# Delete a service account by name
	discovery core service-account delete ci`,
	}
	return deleteCmd
}
