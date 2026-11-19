package tokens

import (
	"github.com/pureinsights/discovery-cli/v2/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/v2/discovery"
	"github.com/pureinsights/discovery-cli/v2/internal/cli"
	"github.com/spf13/cobra"
)

// NewGetCommand creates the token get command.
func NewGetCommand(d cli.Discovery) *cobra.Command {
	var serviceAccount string
	get := &cobra.Command{
		Use:   "get [<token>]",
		Short: "The command that obtains tokens from Discovery Core.",
		Long:  "get is the command used to obtain the access tokens of a Discovery Core service account. The user can send a name or UUID to get a specific token. If no argument is given, then the command retrieves every token of the service account. The service account can be referenced by name or UUID.",
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
			return commands.SearchCommand(args, d, client, commands.GetCommandConfig(profile, vpr.GetString("output"), "Core", "core_url"), &[]string{})
		},
		Args: cobra.MaximumNArgs(1),
		Example: `	# Get a token by name
	discovery core token get "My Token for Deployment" --service-account "ci"

	# Get all tokens using the configuration in profile "cn"
	discovery core token get --service-account "ci" -p cn`,
	}

	addServiceAccountFlag(get, &serviceAccount)
	return get
}
