package serviceaccounts

import (
	"github.com/pureinsights/discovery-cli/v2/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/v2/discovery"
	"github.com/pureinsights/discovery-cli/v2/internal/cli"
	"github.com/spf13/cobra"
)

// NewDisableCommand creates the service account disable command.
func NewDisableCommand(d cli.Discovery) *cobra.Command {
	disableCmd := &cobra.Command{
		Use:   "disable <service-account>",
		Short: "The command that disables service accounts in Discovery Core.",
		Long:  "disable is the command used to disable Discovery Core's service accounts. The user can send the name or the UUID of the service account. If it is disabled, it should acknowledge the configuration.",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			err = commands.CheckCredentials(d, profile, "Core", "core_url")
			if err != nil {
				return err
			}

			vpr := d.Config()

			coreClient := discoveryPackage.NewCore(vpr.GetString(profile+".core_url"), vpr.GetString(profile+".core_key"))
			printer := cli.GetObjectPrinter(vpr.GetString("output"))
			return d.DisableEntity(coreClient.ServiceAccounts(), args[0], printer)
		},
		Args: cobra.ExactArgs(1),
		Example: `	# Disable a service account by name
	discovery core service-account disable ci

	# Disable a service account by id
	discovery core service-account disable 3d51beef-8b90-40aa-84b5-033241dc6239`,
	}

	return disableCmd
}
