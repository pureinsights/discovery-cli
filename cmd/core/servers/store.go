package servers

import (
	"fmt"

	"github.com/pureinsights/discovery-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewStoreCommand creates the server store command
func NewStoreCommand(d cli.Discovery) *cobra.Command {
	var abortOnError bool
	var data string
	var file string
	store := &cobra.Command{
		Use:   "store",
		Short: "The command that stores servers to Discovery Core.",
		Long:  fmt.Sprintf(commands.LongStore, "server", "Core"),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			coreClient := discoveryPackage.NewCore(vpr.GetString(profile+".core_url"), vpr.GetString(profile+".core_key"))
			return commands.StoreCommand(d, coreClient.Servers(), commands.StoreCommandConfig(commands.GetCommandConfig(profile, vpr.GetString("output"), "Core", "core_url"), abortOnError, data, file))
		},
		Args: cobra.NoArgs,
		Example: `	# Store a server with the JSON configuration in a file
	discovery core server store --file "serverjsonfile.json"

	# Store a server with the JSON configuration in the data flag
	discovery core server store --data '{"type":"mongo","name":"MongoDB Atlas server","labels":[],"active":true,"id":"2b839453-ddad-4ced-8e13-2c7860af60a7","creationTimestamp":"2025-09-29T15:50:26Z","lastUpdatedTimestamp":"2025-09-29T15:50:26Z","config":{"servers":["mongodb+srv://cluster0.dleud.mongodb.net/"],"connection":{"readTimeout":"30s","connectTimeout":"1m"},"credentialId":"9ababe08-0b74-4672-bb7c-e7a8227d6d4c"}}'`,
	}
	store.Flags().BoolVar(&abortOnError, "abort-on-error", false, "aborts the operation if there is an error")
	store.Flags().StringVarP(&data, "data", "d", "", "the JSON with the configurations that will be upserted")
	store.Flags().StringVarP(&file, "file", "f", "", "the path of the file that contains the JSON data")

	store.MarkFlagsOneRequired("data", "file")
	store.MarkFlagsMutuallyExclusive("data", "file")
	return store
}
