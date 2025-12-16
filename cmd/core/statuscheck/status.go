package statuscheck

import (
	"fmt"

	"github.com/pureinsights/discovery-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewExportCommand creates the discovery core export command that exports Discovery Core's entities.
func NewStatusCommand(d cli.Discovery) *cobra.Command {
	export := &cobra.Command{
		Use:   "status",
		Short: "Check if Discovery Core is online",
		Long:  fmt.Sprintf(commands.LongStatusCheck, "Core"),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			coreClient := discoveryPackage.NewCore(vpr.GetString(profile+".core_url"), vpr.GetString(profile+".core_key"))
			return commands.StatusCheckCommand(d, coreClient.StatusChecker(), "Core", commands.GetCommandConfig(profile, vpr.GetString("output"), "Core", "core_url"))
		},
		Args: cobra.NoArgs,
		Example: `	# Check the status of Discovery Core
	discovery core status`,
	}

	return export
}
