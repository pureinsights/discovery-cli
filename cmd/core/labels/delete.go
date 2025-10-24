package labels

import (
	"fmt"

	discoveryPackage "github.com/pureinsights/pdp-cli/discovery"
	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewDeleteCommand creates the label delete command
func NewDeleteCommand(d cli.Discovery) *cobra.Command {
	get := &cobra.Command{
		Use:   "get",
		Short: "The command that obtains labels from Discovery Core.",
		Long:  fmt.Sprintf(cli.LongDeleteNoNames, "label", "Core"),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			coreClient := discoveryPackage.NewCore(vpr.GetString(profile+".core_url"), vpr.GetString(profile+".core_key"))
			return cli.DeleteCommand(args, d, coreClient.Labels(), cli.GetCommandConfig(profile, vpr.GetString("output"), "Core", "core_url", "core_key"))
		},
		Args: cobra.MaximumNArgs(1),
	}
	return get
}
