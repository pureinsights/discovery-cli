package labels

import (
	"fmt"

	"github.com/pureinsights/pdp-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/pdp-cli/discovery"
	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewDeleteCommand creates the label delete command
func NewDeleteCommand(d cli.Discovery) *cobra.Command {
	get := &cobra.Command{
		Use:   "delete",
		Short: "The command that deletes labels from Discovery Core.",
		Long:  fmt.Sprintf(commands.LongDeleteNoNames, "label", "Core"),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			coreClient := discoveryPackage.NewCore(vpr.GetString(profile+".core_url"), vpr.GetString(profile+".core_key"))
			return commands.DeleteCommand(args[0], d, coreClient.Labels(), commands.GetCommandConfig(profile, vpr.GetString("output"), "Core", "core_url", "core_key"))
		},
		Args: cobra.ExactArgs(1),
	}
	return get
}
