package credentials

import (
	"fmt"

	"github.com/pureinsights/discovery-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewDeleteCommand creates the credential delete command
func NewDeleteCommand(d cli.Discovery) *cobra.Command {
	get := &cobra.Command{
		Use:   "delete",
		Short: "The command that deletes credentials from Discovery Core.",
		Long:  fmt.Sprintf(commands.LongDeleteNoNames, "credential", "Core"),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			coreClient := discoveryPackage.NewCore(vpr.GetString(profile+".core_url"), vpr.GetString(profile+".core_key"))
			return commands.SearchDeleteCommand(args[0], d, coreClient.Credentials(), commands.GetCommandConfig(profile, vpr.GetString("output"), "Core", "core_url"))
		},
		Args: cobra.ExactArgs(1),
		Example: `	# Delete a credential by id
	discovery core credential delete 3d51beef-8b90-40aa-84b5-033241dc6239

	# Delete a credential by name
	discovery core credential delete my-credential`,
	}
	return get
}
