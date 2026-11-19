package tokens

import (
	"github.com/pureinsights/discovery-cli/v2/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/v2/discovery"
	"github.com/pureinsights/discovery-cli/v2/internal/cli"
	"github.com/spf13/cobra"
)

// NewUpdateCommand creates the token update command.
func NewUpdateCommand(d cli.Discovery) *cobra.Command {
	var flags tokenFlags
	var serviceAccount string
	update := &cobra.Command{
		Use:   "update <token>",
		Short: "The command that updates tokens in Discovery Core.",
		Long:  "update is the command used to update a token in Discovery Core. The user can send the name or the UUID of the token and the fields to change with flags. Fields without a flag keep their current value. The scopes can be replaced with --scopes, or edited with --add-scopes and --remove-scopes. If no flag is sent, the command asks the user for each field interactively, showing the current values. The service account can be referenced by name or UUID.",
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

			token, err := client.Get(id)
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get token with id %q", id.String())
			}

			config := token.Raw
			if anyFlagChanged(cmd) {
				config, err = applyFlags(cmd, config, flags)
			} else {
				config, err = askToken(d.IOStreams(), config, false)
			}
			if err != nil {
				return err
			}

			return commands.StoreCommand(d, client, commands.StoreCommandConfig(commands.GetCommandConfig(profile, vpr.GetString("output"), "Core", "core_url"), true, config, nil))
		},
		Args: cobra.ExactArgs(1),
		Example: `	# Add and remove scopes of a token
	discovery core token update "My Token" --service-account "ci" --add-scopes "staging:bucket:create,staging:content:*" --remove-scopes "core:server:read"

	# Replace all the scopes of a token
	discovery core token update "My Token" --service-account "ci" --scopes "core:*"

	# Update a token interactively
	discovery core token update "My Token" --service-account "ci"`,
	}

	addTokenFlags(update, &flags, false)
	addScopeEditFlags(update, &flags)
	addServiceAccountFlag(update, &serviceAccount)
	return update
}
