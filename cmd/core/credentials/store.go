package credentials

import (
	"fmt"

	"github.com/pureinsights/discovery-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewStoreCommand creates the credential store command.
func NewStoreCommand(d cli.Discovery) *cobra.Command {
	var abortOnError bool
	var data string
	store := &cobra.Command{
		Use:   "store",
		Short: "The command that stores credentials to Discovery Core.",
		Long:  fmt.Sprintf(commands.LongStore, "credential", "Core"),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			coreClient := discoveryPackage.NewCore(vpr.GetString(profile+".core_url"), vpr.GetString(profile+".core_key"))
			return commands.StoreCommand(d, coreClient.Credentials(), commands.StoreCommandConfig(commands.GetCommandConfig(profile, vpr.GetString("output"), "Core", "core_url"), abortOnError, data, args))
		},
		Example: `	# Store a credential with the JSON configuration in a file
	discovery core credential store "credentialjsonfile.json"

	# Store a credential with the JSON configuration in the data flag
	discovery core credential store --data '{"type":"mongo","name":"my-credential","labels":[{"key":"A","value":"A"}],"active":true,"id":"3b32e410-2f33-412d-9fb8-17970131921c","creationTimestamp":"2025-10-17T22:37:57Z","lastUpdatedTimestamp":"2025-10-17T22:37:57Z","secret":"my-secret"}'`,
	}
	store.Flags().BoolVar(&abortOnError, "abort-on-error", false, "aborts the operation if there is an error")
	store.Flags().StringVarP(&data, "data", "d", "", "the JSON with the configurations that will be upserted")

	return store
}
