package secrets

import (
	"fmt"

	"github.com/pureinsights/pdp-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/pdp-cli/discovery"
	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewGetCommand creates the secret get command
func NewGetCommand(d cli.Discovery) *cobra.Command {
	get := &cobra.Command{
		Use:   "get",
		Short: "The command that obtains secrets from Discovery Core.",
		Long:  fmt.Sprintf(commands.LongGetNoNames, "secret", "Core"),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			coreClient := discoveryPackage.NewCore(vpr.GetString(profile+".core_url"), vpr.GetString(profile+".core_key"))
			return commands.GetCommand(args, d, coreClient.Secrets(), commands.GetCommandConfig(profile, vpr.GetString("output"), "Core", "core_url"))
		},
		Args: cobra.MaximumNArgs(1),
		Example: `	# Get a secret by id
	discovery core secret get 81ca1ac6-3058-4ecd-a292-e439827a675a
	{"active":true,"creationTimestamp":"2025-08-26T21:56:50Z","id":"81ca1ac6-3058-4ecd-a292-e439827a675a","labels":[],"lastUpdatedTimestamp":"2025-08-26T21:56:50Z","name":"openai-secret"}

	# Get all secrets using the configuration in profile "cn"
	discovery core secret get -p cn
	{"active":true,"creationTimestamp":"2025-08-26T21:56:50Z","id":"81ca1ac6-3058-4ecd-a292-e439827a675a","labels":[],"lastUpdatedTimestamp":"2025-08-26T21:56:50Z","name":"openai-secret"}
	{"active":true,"creationTimestamp":"2025-08-14T18:01:59Z","id":"cfa0ef51-1fd9-47e2-8fdb-262ac9712781","labels":[],"lastUpdatedTimestamp":"2025-08-14T18:01:59Z","name":"mongo-secret"}`,
	}
	return get
}
