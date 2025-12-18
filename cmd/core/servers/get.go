package servers

import (
	"fmt"

	"github.com/pureinsights/discovery-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewGetCommand creates the server get command.
func NewGetCommand(d cli.Discovery) *cobra.Command {
	var filters []string
	get := &cobra.Command{
		Use:   "get [server]",
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
		Example: `	# Get server by name
	discovery core server get "my-server"

	# Get servers using filters
	discovery core server get --filter label=A:A -f type=mongo

	# Get all servers using the configuration in profile "cn"
	discovery core server get -p cn`,
	}

	get.Flags().StringArrayVarP(&filters, "filter", "f", []string{}, `apply filters in the format "filter=key:value". The available filters are:
- Label: The format is label={key}[:{value}], where the value is optional
- Type: The format is type={type}`)
	return get
}
