package labels

import (
	"fmt"

	"github.com/pureinsights/discovery-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewGetCommand creates the label get command.
func NewGetCommand(d cli.Discovery) *cobra.Command {
	get := &cobra.Command{
		Use:   "get [labelId]",
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
		Example: `	# Get a label by id
	discovery core label get 3d51beef-8b90-40aa-84b5-033241dc6239
	
	# Get all labels using the configuration in profile "cn"
	discovery core label get -p cn`,
	}
	return get
}
