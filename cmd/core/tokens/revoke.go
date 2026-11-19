package tokens

import (
	"fmt"

	"github.com/pureinsights/discovery-cli/v2/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/v2/discovery"
	"github.com/pureinsights/discovery-cli/v2/internal/cli"
	"github.com/spf13/cobra"
)

// NewRevokeCommand creates the access token revoke command.
func NewRevokeCommand(d cli.Discovery) *cobra.Command {
	var serviceAccount string
	revokeCmd := &cobra.Command{
		Use:   "revoke <token>",
		Short: "The command that revokes access tokens from Discovery service accounts.",
		Long:  fmt.Sprintf(commands.LongDeleteSearch, "token", "Core"),
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

			client := discoveryPackage.NewCore(coreUrl, coreKey).Tokens(accountId)
			return commands.SearchDeleteCommand(args[0], d, client, commands.GetCommandConfig(profile, vpr.GetString("output"), "Core", "core_url"))
		},
		Args: cobra.ExactArgs(1),
		Example: `	# Revoke an access token of a service account by id
	discovery core token revoke 3d51beef-8b90-40aa-84b5-033241dc6239 --service-account "ci"

	# Revoke an access token of a service account by name
	discovery core token revoke "My Token for Deployment" --service-account "ci"`,
	}

	addServiceAccountFlag(revokeCmd, &serviceAccount)
	return revokeCmd
}
