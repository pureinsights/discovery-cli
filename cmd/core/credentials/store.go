package credentials

import (
	"fmt"

	"github.com/pureinsights/pdp-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/pdp-cli/discovery"
	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewStoreCommand creates the credential store command
func NewStoreCommand(d cli.Discovery) *cobra.Command {
	var abortOnError bool
	var data string
	var file string
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
			return commands.StoreCommand(d, coreClient.Credentials(), commands.StoreCommandConfig(commands.GetCommandConfig(profile, vpr.GetString("output"), "Core", "core_url"), abortOnError, data, file))
		},
		Args: cobra.NoArgs,
		Example: `	# Store a credential with the JSON configuration in a file
	discovery core credential store --file "credentialjsonfile.json"
	{"active":true,"creationTimestamp":"2025-10-17T22:37:57Z","id":"3b32e410-2f33-412d-9fb8-17970131921c","labels":[{"key":"A","value":"A"}],"lastUpdatedTimestamp":"2025-10-17T22:37:57Z","name":"my-credential-1","secret":"mongo-secret","type":"mongo"}
	{"code":1003,"messages":["Entity not found: 3b32e410-2f33-412d-9fb8-17970131921d"],"status":404,"timestamp":"2025-10-30T16:50:38.250661200Z"}
	{"active":true,"creationTimestamp":"2025-10-30T16:50:38.262086Z","id":"5b76ae0d-f383-47e5-be6f-90e9046092cd","labels":[{"key":"A","value":"B"}],"lastUpdatedTimestamp":"2025-10-30T16:50:38.262086Z","name":"my-credential-2","secret":"mongo-secret","type":"mongo"}

	# Store a credential with the JSON configuration in the data flag
	discovery core credential store --data '{"type":"mongo","name":"my-credential-1","labels":[{"key":"A","value":"A"}],"active":true,"id":"3b32e410-2f33-412d-9fb8-17970131921c","creationTimestamp":"2025-10-17T22:37:57Z","lastUpdatedTimestamp":"2025-10-17T22:37:57Z","secret":"mongo-secret"}'
	{"active":true,"creationTimestamp":"2025-10-17T22:37:57Z","id":"3b32e410-2f33-412d-9fb8-17970131921c","labels":[{"key":"A","value":"A"}],"lastUpdatedTimestamp":"2025-10-17T22:37:57Z","name":"my-credential-1","secret":"mongo-secret","type":"mongo"}`,
	}
	store.Flags().BoolVar(&abortOnError, "abort-on-error", false, "Aborts the operation if there is an error")
	store.Flags().StringVarP(&data, "data", "d", "", "The JSON with the configurations that will be upserted")
	store.Flags().StringVarP(&file, "file", "f", "", "The path of the file that contains the JSON data")

	store.MarkFlagsOneRequired("data", "file")
	store.MarkFlagsMutuallyExclusive("data", "file")
	return store
}
