package processors

import (
	"fmt"

	"github.com/pureinsights/discovery-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewStoreCommand creates the processor store command.
func NewStoreCommand(d cli.Discovery) *cobra.Command {
	var abortOnError bool
	var data string
	store := &cobra.Command{
		Use:   "store [<files>...]",
		Short: "The command that stores processors to Discovery QueryFlow.",
		Long:  fmt.Sprintf(commands.LongStore, "processor", "QueryFlow"),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			queryflowClient := discoveryPackage.NewQueryFlow(vpr.GetString(profile+".queryflow_url"), vpr.GetString(profile+".queryflow_key"))
			return commands.StoreCommand(d, queryflowClient.Processors(), commands.StoreCommandConfig(commands.GetCommandConfig(profile, vpr.GetString("output"), "QueryFlow", "queryflow_url"), abortOnError, data, args))
		},
		Example: `	# Store a processor with the JSON configuration in a file
	discovery queryflow processor store "queryflowprocessorjsonfile.json"
	
	# Store a processor with the JSON configuration in the data flag
	discovery queryflow processor store --data '{"type":"mongo","name":"my-processor","labels":[],"active":true,"id":"3393f6d9-94c1-4b70-ba02-5f582727d998","creationTimestamp":"2025-11-20T00:08:23Z","lastUpdatedTimestamp":"2025-11-20T00:08:23Z","config":{"action":"aggregate","stages":[{"$match":{"$text":{"$search":"#{ data(\"/httpRequest/queryParams/q\") }"}}}],"database":"pureinsights","collection":"blogs"},"server":{"id":"f6950327-3175-4a98-a570-658df852424a","credential":"9ababe08-0b74-4672-bb7c-e7a8227d6d4c"}}'`,
	}
	store.Flags().BoolVar(&abortOnError, "abort-on-error", false, "aborts the operation if there is an error")
	store.Flags().StringVarP(&data, "data", "d", "", "the JSON with the configurations that will be upserted")

	return store
}
