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
	{"active":true,"config":{"action":"hydrate","collection":"blogs","data":{"author":"#{ data('/author') }","header":"#{ data('/header') }","link":"#{ data('/reference') }"},"database":"pureinsights"},"creationTimestamp":"2025-10-30T20:07:44Z","id":"e9c4173f-6906-43a8-b3ca-7319d3d24754","labels":[],"lastUpdatedTimestamp":"2025-10-30T20:07:44Z","name":"MongoDB store processor","server":{"credential":"9ababe08-0b74-4672-bb7c-e7a8227d6d4c","id":"f6950327-3175-4a98-a570-658df852424a"},"type":"mongo"}
	{"code":1003,"messages":["Entity not found: e9c4173f-6906-43a8-b3ca-7319d3d24755"],"status":404,"timestamp":"2025-10-30T20:09:29.314467Z"}
	{"active":true,"config":{"action":"hydrate","collection":"blogs","data":{"author":"#{ data('/author') }","header":"#{ data('/header') }","link":"#{ data('/reference') }"},"database":"pureinsights"},"creationTimestamp":"2025-10-30T20:09:29.346792Z","id":"aef648d8-171d-479a-a6fd-14ec9b235dc7","labels":[],"lastUpdatedTimestamp":"2025-10-30T20:09:29.346792Z","name":"MongoDB store processor 2","server":{"credential":"9ababe08-0b74-4672-bb7c-e7a8227d6d4c","id":"f6950327-3175-4a98-a570-658df852424a"},"type":"mongo"}

	# Store a processor with the JSON configuration in the data flag
	discovery ingestion processor store --data '{"type":"mongo","name":"MongoDB store processor","labels":[],"active":true,"id":"e9c4173f-6906-43a8-b3ca-7319d3d24754","creationTimestamp":"2025-10-30T20:07:43.825231Z","lastUpdatedTimestamp":"2025-10-30T20:07:43.825231Z","config":{"data":{"link":"#{ data('/reference') }","author":"#{ data('/author') }","header":"#{ data('/header') }"},"action":"hydrate","database":"pureinsights","collection":"blogs"},"server":{"id":"f6950327-3175-4a98-a570-658df852424a","credential":"9ababe08-0b74-4672-bb7c-e7a8227d6d4c"}}'
	{"active":true,"config":{"action":"hydrate","collection":"blogs","data":{"author":"#{ data(/author) }","header":"#{ data(/header) }","link":"#{ data(/reference) }"},"database":"pureinsights"},"creationTimestamp":"2025-10-30T20:07:44Z","id":"e9c4173f-6906-43a8-b3ca-7319d3d24754","labels":[],"lastUpdatedTimestamp":"2025-10-30T20:10:23.698799Z","name":"MongoDB store processor","server":{"credential":"9ababe08-0b74-4672-bb7c-e7a8227d6d4c","id":"f6950327-3175-4a98-a570-658df852424a"},"type":"mongo"}`,
	}
	store.Flags().BoolVar(&abortOnError, "abort-on-error", false, "Aborts the operation if there is an error")
	store.Flags().StringVarP(&data, "data", "d", "", "The JSON with the configurations that will be upserted")
	store.Flags().StringVarP(&file, "file", "f", "", "The path of the file that contains the JSON data")

	store.MarkFlagsOneRequired("data", "file")
	store.MarkFlagsMutuallyExclusive("data", "file")
	return store
}
