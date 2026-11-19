package tokens

import (
	"github.com/pureinsights/discovery-cli/v2/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/v2/discovery"
	"github.com/pureinsights/discovery-cli/v2/internal/cli"
	"github.com/spf13/cobra"
)

// NewRenewCommand creates the access token renew command.
func NewRenewCommand(d cli.Discovery) *cobra.Command {
	var serviceAccount, ttl string
	renewCmd := &cobra.Command{
		Use:   "renew <token>",
		Short: "The command that renews access tokens for Discovery service accounts.",
		Long:  "renew is the command used to renew Discovery Core's access tokens. The user can send the name or the UUID of the access token, the --service-account flag with the name or UUID of its service account, and the --ttl flag with the new time-to-live. Renewing issues a new token value, which is printed only once, and computes the new expiration from now.",
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
			id, err := cli.GetEntityId(d, client, args[0])
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the token ID.")
			}

			renewed, err := client.Renew(id, ttl)
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not renew token with id %q", id.String())
			}

			return cli.GetObjectPrinter(vpr.GetString("output"))(*d.IOStreams(), renewed)
		},
		Args: cobra.ExactArgs(1),
		Example: `	# Renew an access token of a service account by id for 60 days
	discovery core token renew 3d51beef-8b90-40aa-84b5-033241dc6239 --service-account "ci" --ttl "60d"

	# Renew an access token of a service account by name so it never expires
	discovery core token renew "My Token for Deployment" --service-account "ci" --ttl "-1"`,
	}

	addServiceAccountFlag(renewCmd, &serviceAccount)
	renewCmd.Flags().StringVar(&ttl, "ttl", "", "the new time-to-live of the token, such as 3600, 90s, 15m, 12h or 60d (-1 for a token that never expires)")
	_ = renewCmd.MarkFlagRequired("ttl")
	return renewCmd
}
