package secrets

import (
	"fmt"

	"github.com/pureinsights/discovery-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cliry-cliry-cli/discovery"
	"github.com/pureinsights/discovery-cliry-cliry-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewStoreCommand creates the secret store command
func NewStoreCommand(d cli.Discovery) *cobra.Command {
	var abortOnError bool
	var data string
	var file string
	store := &cobra.Command{
		Use:   "store",
		Short: "The command that stores secrets to Discovery Core.",
		Long:  fmt.Sprintf(commands.LongStore, "secret", "Core"),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			coreClient := discoveryPackage.NewCore(vpr.GetString(profile+".core_url"), vpr.GetString(profile+".core_key"))
			return commands.StoreCommand(d, coreClient.Secrets(), commands.StoreCommandConfig(commands.GetCommandConfig(profile, vpr.GetString("output"), "Core", "core_url"), abortOnError, data, file))
		},
		Args: cobra.NoArgs,
		Example: `	# Store a secret with the JSON configuration in a file
	discovery core secret store --file "secretjsonfile.json"

	# Store a secret with the JSON configuration in the data flag
	discovery core secret store --data  '{"name":"openai-secret-test","active":true,"id":"b8bd5ec3-8f60-4502-b25e-8f6d36c98410","content":{"apiKey":"apiKey"}}'`,
	}
	store.Flags().BoolVar(&abortOnError, "abort-on-error", false, "Aborts the operation if there is an error")
	store.Flags().StringVarP(&data, "data", "d", "", "The JSON with the configurations that will be upserted")
	store.Flags().StringVarP(&file, "file", "f", "", "The path of the file that contains the JSON data")

	store.MarkFlagsOneRequired("data", "file")
	store.MarkFlagsMutuallyExclusive("data", "file")
	return store
}
