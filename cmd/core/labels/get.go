package labels

import (
	"fmt"

	"github.com/pureinsights/pdp-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/pdp-cli/discovery"
	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewGetCommand creates the label get command
func NewGetCommand(d cli.Discovery) *cobra.Command {
	get := &cobra.Command{
		Use:   "get",
		Short: "The command that obtains labels from Discovery Core.",
		Long:  fmt.Sprintf(commands.LongGetNoNames, "label", "Core"),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			coreClient := discoveryPackage.NewCore(vpr.GetString(profile+".core_url"), vpr.GetString(profile+".core_key"))
			return commands.GetCommand(args, d, coreClient.Labels(), commands.GetCommandConfig(profile, vpr.GetString("output"), "Core", "core_url"))
		},
		Args: cobra.MaximumNArgs(1),
		Example: `
	# Get a label by id
	discovery core label get 3d51beef-8b90-40aa-84b5-033241dc6239
	{"creationTimestamp":"2025-08-27T19:22:06Z","id":"3d51beef-8b90-40aa-84b5-033241dc6239","key":"A","lastUpdatedTimestamp":"2025-08-27T19:22:47Z","value":"B"}
	
	# Get all labels using the configuration in profile "cn"
	discovery core label get -p cn
	{"creationTimestamp":"2025-10-15T20:28:39Z","id":"5467ab23-7827-4fae-aa78-dfd4800549ee","key":"D","lastUpdatedTimestamp":"2025-10-15T20:28:39Z","value":"F"}
	{"creationTimestamp":"2025-10-15T20:25:29Z","id":"7d0cb8c9-6555-4592-9b6c-1f4ed7fca9f4","key":"D","lastUpdatedTimestamp":"2025-10-15T20:25:29Z","value":"D"}`,
	}
	return get
}
