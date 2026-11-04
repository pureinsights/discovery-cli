package serviceaccounts

import (
	"github.com/pureinsights/discovery-cli/v2/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/v2/discovery"
	"github.com/pureinsights/discovery-cli/v2/internal/cli"
	"github.com/spf13/cobra"
	"github.com/tidwall/sjson"
)

// NewCreateCommand creates the service account create command.
func NewCreateCommand(d cli.Discovery) *cobra.Command {
	var flags accountFlags
	create := &cobra.Command{
		Use:   "create [<name>]",
		Short: "The command that creates service accounts in Discovery Core.",
		Long:  "create is the command used to create a service account in Discovery Core. The user can send the name as an argument and the rest of the fields with flags. The name and the role are required. If neither the name nor any flag is sent, the command asks the user for each field interactively.",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			err = commands.CheckCredentials(d, profile, "Core", "core_url")
			if err != nil {
				return err
			}

			config := `{"active":true}`
			if len(args) == 0 && !anyFlagChanged(cmd) {
				config, err = askAccount(d.IOStreams(), config)
			} else {
				if len(args) == 1 {
					config, _ = sjson.Set(config, "name", args[0])
				}
				config, err = applyFlags(cmd, config, flags)
			}
			if err != nil {
				return err
			}

			vpr := d.Config()

			coreClient := discoveryPackage.NewCore(vpr.GetString(profile+".core_url"), vpr.GetString(profile+".core_key"))
			return commands.StoreCommand(d, coreClient.ServiceAccounts(), commands.StoreCommandConfig(commands.GetCommandConfig(profile, vpr.GetString("output"), "Core", "core_url"), true, config, nil))
		},
		Args: cobra.MaximumNArgs(1),
		Example: `	# Create a service account with flags
	discovery core service-account create "ci" --description "Continuous Integration account" --role "OPERATE" --labels "keyA:valueA,keyB:valueB" --enabled

	# Create a disabled service account with only the required fields
	discovery core service-account create "ci" --role "OPERATE" --enabled=false

	# Create a service account interactively
	discovery core service-account create`,
	}

	addAccountFlags(create, &flags)
	return create
}
