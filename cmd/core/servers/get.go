package servers

import (
	"fmt"

	"github.com/pureinsights/pdp-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/pdp-cli/discovery"
	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewGetCommand creates the server get command
func NewGetCommand(d cli.Discovery) *cobra.Command {
	var filters []string
	get := &cobra.Command{
		Use:   "get",
		Short: "The command that obtains servers from Discovery Core.",
		Long:  fmt.Sprintf(commands.LongGetSearch, "server", "Core"),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			coreClient := discoveryPackage.NewCore(vpr.GetString(profile+".core_url"), vpr.GetString(profile+".core_key"))
			return commands.SearchCommand(args, d, coreClient.Servers(), commands.GetCommandConfig(profile, vpr.GetString("output"), "Core", "core_url"), &filters)
		},
		Args: cobra.MaximumNArgs(1),
		Example: `
	# Get server by name
	discovery core server get "MongoDB Atlas server"
	{"active":true,"config":{"connection":{"connectTimeout":"1m","readTimeout":"30s"},"credentialId":"9ababe08-0b74-4672-bb7c-e7a8227d6d4c","servers":["mongodb+srv://cluster0.dleud.mongodb.net/"]},"creationTimestamp":"2025-09-29T15:50:19Z","id":"21029da3-041c-43b5-a67e-870251f2f6a6","labels":[{"key":"A","value":"A"}],"lastUpdatedTimestamp":"2025-09-29T15:50:19Z","name":"MongoDB Atlas server","type":"mongo"}

	# Get servers using filters
	discovery core server get --filter label=A:A -f type=mongo
	{"active":true,"creationTimestamp":"2025-09-29T15:50:19Z","id":"21029da3-041c-43b5-a67e-870251f2f6a6","labels":[{"key":"A","value":"A"}],"lastUpdatedTimestamp":"2025-09-29T15:50:19Z","name":"MongoDB Atlas server","type":"mongo"}
	{"active":true,"creationTimestamp":"2025-09-29T15:50:21Z","id":"a798cd5b-aa7a-4fc5-9292-1de6fe8e8b7f","labels":[{"key":"A","value":"A"}],"lastUpdatedTimestamp":"2025-09-29T15:50:21Z","name":"MongoDB Atlas server 2","type":"mongo"}

	# Get all servers using the configuration in profile "cn"
	discovery core server get -p cn
	{"active":true,"creationTimestamp":"2025-09-29T15:50:37Z","id":"025347a7-e2bd-4ba1-880f-db3e51319abb","labels":[],"lastUpdatedTimestamp":"2025-09-29T15:50:37Z","name":"MongoDB Atlas server","type":"mongo"}
	{"active":true,"creationTimestamp":"2025-10-15T20:26:27Z","id":"192c3793-600a-4366-9778-7d80a0df07ce","labels":[{"key":"E","value":"G"},{"key":"H","value":"F"},{"key":"D","value":"D"}],"lastUpdatedTimestamp":"2025-10-15T20:26:27Z","name":"OpenAI Server","type":"openai"}`,
	}

	get.Flags().StringArrayVarP(&filters, "filter", "f", []string{}, `Apply filters in the format "filter=key:value". The available filters are:
- Label: The format is label={key}[:{value}], where the value is optional.
- Type: The format is type={type}.`)
	return get
}
