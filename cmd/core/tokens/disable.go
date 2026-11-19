package tokens

import (
	"github.com/pureinsights/discovery-cli/v2/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/v2/discovery"
	"github.com/pureinsights/discovery-cli/v2/internal/cli"
	"github.com/spf13/cobra"
)

// NewDisableCommand creates the token disable command.
func NewDisableCommand(d cli.Discovery) *cobra.Command {
	var serviceAccount string
	disableCmd := &cobra.Command{
		Use:   "disable <token>",
		Short: "The command that disables tokens in Discovery Core.",
		Long:  "disable is the command used to disable Discovery Core's tokens. The user can send the name or the UUID of the token and the --service-account flag with the name or UUID of its service account. If it is disabled, it should acknowledge the configuration.",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			if err = commands.CheckCredentials(d, profile, "Core", "core_url"); err != nil {
				return err
			}

			vpr := d.Config()
			coreUrl, coreKey := vpr.GetString(profile+".core_url"), vpr.GetString(profile+".core_key")
			accountId, err := serviceAccountId(d, coreUrl, coreKey, serviceAccount)
			if err != nil {
				return err
			}

			printer := cli.GetObjectPrinter(vpr.GetString("output"))
			return d.DisableEntity(discoveryPackage.NewCore(coreUrl, coreKey).Tokens(accountId), args[0], printer)
		},
		Args: cobra.ExactArgs(1),
		Example: `	# Disable a token by name
	discovery core token disable "My Token for Deployment" --service-account "ci"

	# Disable a token by id
	discovery core token disable 3d51beef-8b90-40aa-84b5-033241dc6239 --service-account "ci"`,
	}

	addServiceAccountFlag(disableCmd, &serviceAccount)
	return disableCmd
}
