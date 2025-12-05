package servers

import (
	"fmt"

	"github.com/pureinsights/discovery-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cliry-cliry-cli/discovery"
	"github.com/pureinsights/discovery-cliry-cliry-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewDeleteCommand creates the server delete command
func NewDeleteCommand(d cli.Discovery) *cobra.Command {
	get := &cobra.Command{
		Use:   "delete",
		Short: "The command that deletes servers from Discovery Core.",
		Long:  fmt.Sprintf(commands.LongDeleteNoNames, "server", "Core"),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			coreClient := discoveryPackage.NewCore(vpr.GetString(profile+".core_url"), vpr.GetString(profile+".core_key"))
			return commands.SearchDeleteCommand(args[0], d, coreClient.Servers(), commands.GetCommandConfig(profile, vpr.GetString("output"), "Core", "core_url"))
		},
		Args: cobra.ExactArgs(1),
		Example: `	# Delete a server by id
	discovery core server delete 3d51beef-8b90-40aa-84b5-033241dc6239

	# Delete a server by name
	discovery core server delete server1`,
	}
	return get
}
