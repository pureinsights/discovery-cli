package credentials

import (
	"fmt"

	"github.com/pureinsights/pdp-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/pdp-cli/discovery"
	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewGetCommand creates the credential get command
func NewGetCommand(d cli.Discovery) *cobra.Command {
	var filters []string
	get := &cobra.Command{
		Use:   "get",
		Short: "The command that obtains credentials from Discovery Core.",
		Long:  fmt.Sprintf(commands.LongGetSearch, "credential", "Core"),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			coreClient := discoveryPackage.NewCore(vpr.GetString(profile+".core_url"), vpr.GetString(profile+".core_key"))
			return commands.SearchCommand(args, d, coreClient.Credentials(), commands.GetCommandConfig(profile, vpr.GetString("output"), "Core", "core_url"), &filters)
		},
		Args: cobra.MaximumNArgs(1),
		Example: `
	# Get credential by name
	discovery core credential get "my-credential"
	{"active":true,"creationTimestamp":"2025-10-17T22:37:57Z","id":"3b32e410-2f33-412d-9fb8-17970131921c","labels":[{"key":"A","value":"A"}],"lastUpdatedTimestamp":"2025-10-17T22:37:57Z","name":"my-credential","type":"mongo"}

	# Get credentials using filters
	discovery core credential get --filter label=A:A --filter type=mongo
	{"active":true,"creationTimestamp":"2025-10-17T15:33:58Z","id":"8c243a1d-9384-421d-8f99-4ef28d4e0ab0","labels":[{"key":"A","value":"A"}],"lastUpdatedTimestamp":"2025-10-17T15:33:58Z","name":"my-credential","type":"mongo"}
	{"active":true,"creationTimestamp":"2025-10-17T22:37:53Z","id":"4957145b-6192-4862-a5da-e97853974e9f","labels":[{"key":"A","value":"A"}],"lastUpdatedTimestamp":"2025-10-17T22:37:53Z","name":"mongo-credential-2","type":"mongo"}

	# Get all credentials using the configuration in profile "cn"
	discovery core credential get -p cn
	{"active":true,"creationTimestamp":"2025-10-17T22:37:57Z","id":"3b32e410-2f33-412d-9fb8-17970131921c","labels":[{"key":"A","value":"A"}],"lastUpdatedTimestamp":"2025-10-17T22:37:57Z","name":"my-credential","type":"mongo"}
	{"active":true,"creationTimestamp":"2025-10-17T22:40:15Z","id":"458d245a-6ed2-4c2b-a73f-5540d550a479","labels":[{"key":"A","value":"B"}],"lastUpdatedTimestamp":"2025-10-17T22:40:15Z","name":"openai-credential","type":"openai"}`,
	}

	get.Flags().StringArrayVarP(&filters, "filter", "f", []string{}, `Apply filters in the format "filter=key:value". The available filters are:
- Label: The format is label={key}[:{value}], where the value is optional.
- Type: The format is type={type}.`)
	return get
}
