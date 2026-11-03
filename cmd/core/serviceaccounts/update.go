package serviceaccounts

import (
	"github.com/pureinsights/discovery-cli/v2/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/v2/discovery"
	"github.com/pureinsights/discovery-cli/v2/internal/cli"
	"github.com/spf13/cobra"
)

// NewUpdateCommand creates the service account update command.
func NewUpdateCommand(d cli.Discovery) *cobra.Command {
	var flags accountFlags
	update := &cobra.Command{
		Use:   "update <service-account>",
		Short: "The command that updates service accounts in Discovery Core.",
		Long:  "update is the command used to update a service account in Discovery Core. The user can send the name or the UUID of the service account and the fields to change with flags. Fields without a flag keep their current value. If no flag is sent, the command asks the user for each field interactively, showing the current values.",
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
			client := coreClient.ServiceAccounts()

			id, err := cli.GetEntityId(d, client, args[0])
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get service account ID.")
			}

			account, err := client.Get(id)
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get service account with id %q", id.String())
			}

			config := account.Raw
			if anyFlagChanged(cmd) {
				config, err = applyFlags(cmd, config, flags)
			} else {
				config, err = askAccount(d.IOStreams(), config)
			}
			if err != nil {
				return err
			}

			return commands.StoreCommand(d, client, commands.StoreCommandConfig(commands.GetCommandConfig(profile, vpr.GetString("output"), "Core", "core_url"), true, config, nil))
		},
		Args: cobra.ExactArgs(1),
		Example: `	# Update the role of a service account
	discovery core service-account update "ci" --role "ADMIN"

	# Update a service account interactively
	discovery core service-account update ci`,
	}

	addAccountFlags(update, &flags)
	return update
}
