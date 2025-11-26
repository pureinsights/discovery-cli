package labels

import (
	"fmt"

	"github.com/pureinsights/pdp-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/pdp-cli/discovery"
	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewStoreCommand creates the label store command
func NewStoreCommand(d cli.Discovery) *cobra.Command {
	var abortOnError bool
	var data string
	var file string
	store := &cobra.Command{
		Use:   "store",
		Short: "The command that stores labels to Discovery Core.",
		Long:  fmt.Sprintf(commands.LongStore, "label", "Core"),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			coreClient := discoveryPackage.NewCore(vpr.GetString(profile+".core_url"), vpr.GetString(profile+".core_key"))
			return commands.StoreCommand(d, coreClient.Labels(), commands.StoreCommandConfig(commands.GetCommandConfig(profile, vpr.GetString("output"), "Core", "core_url"), abortOnError, data, file))
		},
		Args: cobra.NoArgs,
		Example: `	# Store a label with the JSON configuration in a file
	discovery core label store --file "labeljsonfile.json"
	{"creationTimestamp":"2025-08-27T19:22:06Z","id":"3d51beef-8b90-40aa-84b5-033241dc6239","key":"label1","lastUpdatedTimestamp":"2025-10-29T22:41:37Z","value":"value1"}
	{"code":1003,"messages":["Entity not found: 3d51beef-8b90-40aa-84b5-033241dc6230"],"status":404,"timestamp":"2025-10-30T00:05:35.995533500Z"}
	{"creationTimestamp":"2025-10-30T00:05:36.004363Z","id":"4967bc7b-ed89-4843-ab0f-1fd73daad30d","key":"label3","lastUpdatedTimestamp":"2025-10-30T00:05:36.004363Z","value":"value3"}
	
	# Store a label with the JSON configuration in the data flag
	discovery core label store --data  '[{"key":"label","value":"labelvalue"}]'
	{"creationTimestamp":"2025-10-30T00:07:07.244729Z","id":"e7870373-da6d-41af-b5ec-91cfd087ee91","key":"label","lastUpdatedTimestamp":"2025-10-30T00:07:07.244729Z","value":"labelvalue"}`,
	}
	store.Flags().BoolVar(&abortOnError, "abort-on-error", false, "Aborts the operation if there is an error")
	store.Flags().StringVarP(&data, "data", "d", "", "The JSON with the configurations that will be upserted")
	store.Flags().StringVarP(&file, "file", "f", "", "The path of the file that contains the JSON data")

	store.MarkFlagsOneRequired("data", "file")
	store.MarkFlagsMutuallyExclusive("data", "file")
	return store
}
