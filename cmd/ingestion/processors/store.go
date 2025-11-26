package processors

import (
	"fmt"

	"github.com/pureinsights/pdp-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/pdp-cli/discovery"
	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewStoreCommand creates the processor store command
func NewStoreCommand(d cli.Discovery) *cobra.Command {
	var abortOnError bool
	var data string
	var file string
	store := &cobra.Command{
		Use:   "store",
		Short: "The command that stores processors to Discovery Ingestion.",
		Long:  fmt.Sprintf(commands.LongStore, "processor", "Ingestion"),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			ingestionClient := discoveryPackage.NewIngestion(vpr.GetString(profile+".ingestion_url"), vpr.GetString(profile+".ingestion_key"))
			return commands.StoreCommand(d, ingestionClient.Processors(), commands.StoreCommandConfig(commands.GetCommandConfig(profile, vpr.GetString("output"), "Ingestion", "ingestion_url"), abortOnError, data, file))
		},
		Args: cobra.NoArgs,
		Example: `	# Store a processor with the JSON configuration in a file
	discovery ingestion processor store --file "ingestionprocessorjsonfile.json"

	# Store a processor with the JSON configuration in the data flag
	discovery ingestion processor store --data '{"type":"mongo","name":"MongoDB store processor","labels":[],"active":true,"id":"e9c4173f-6906-43a8-b3ca-7319d3d24754","creationTimestamp":"2025-10-30T20:07:43.825231Z","lastUpdatedTimestamp":"2025-10-30T20:07:43.825231Z","config":{"data":{"link":"#{ data('/reference') }","author":"#{ data('/author') }","header":"#{ data('/header') }"},"action":"hydrate","database":"pureinsights","collection":"blogs"},"server":{"id":"f6950327-3175-4a98-a570-658df852424a","credential":"9ababe08-0b74-4672-bb7c-e7a8227d6d4c"}}'`,
	}
	store.Flags().BoolVar(&abortOnError, "abort-on-error", false, "Aborts the operation if there is an error")
	store.Flags().StringVarP(&data, "data", "d", "", "The JSON with the configurations that will be upserted")
	store.Flags().StringVarP(&file, "file", "f", "", "The path of the file that contains the JSON data")

	store.MarkFlagsOneRequired("data", "file")
	store.MarkFlagsMutuallyExclusive("data", "file")
	return store
}
