package tokens

import (
	"github.com/pureinsights/discovery-cli/v2/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/v2/discovery"
	"github.com/pureinsights/discovery-cli/v2/internal/cli"
	"github.com/spf13/cobra"
	"github.com/tidwall/sjson"
)

// NewCreateCommand creates the access token create command.
func NewCreateCommand(d cli.Discovery) *cobra.Command {
	flags := tokenFlags{}
	var serviceAccount string

	create := &cobra.Command{
		Use:   "create [<name>]",
		Short: "The command that creates access tokens for Discovery service accounts.",
		Long:  "create is the command used to create an access token for a service account in Discovery Core. The user can send the name as an argument and the rest of the fields with flags. The name and the scopes are required. If neither the name nor any flag is sent, the command asks the user for each field interactively. The service account can be referenced by name or UUID.",
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

			config := `{"active":true}`
			if len(args) == 0 && !anyFlagChanged(cmd) {
				config, err = askToken(d.IOStreams(), config, true)
			} else {
				if len(args) == 1 {
					config, _ = sjson.Set(config, "name", args[0])
				}
				config, err = applyFlags(cmd, config, flags)
			}
			if err != nil {
				return err
			}

			client := discoveryPackage.NewCore(coreUrl, coreKey).Tokens(accountId)
			return commands.StoreCommand(d, client, commands.StoreCommandConfig(commands.GetCommandConfig(profile, vpr.GetString("output"), "Core", "core_url"), true, config, nil))
		},
		Args: cobra.MaximumNArgs(1),
		Example: `	# Create a token with flags
	discovery core token create "My Token for Deployment" --service-account "ci" --scopes "ingestion:seed:write,core:server:read,queryflow:*" --ttl "30d"

	# Create a disabled token with only the required fields
	discovery core token create "My Token for Deployment" --service-account "ci" --scopes "ingestion:seed:write,core:server:read,queryflow:*" --enabled=false

	# Create a token interactively
	discovery core token create --service-account "ci"`,
	}

	addTokenFlags(create, &flags, true)
	addServiceAccountFlag(create, &serviceAccount)
	return create
}
