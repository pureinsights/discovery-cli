package statuscheck

import (
	"fmt"

	"github.com/pureinsights/discovery-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewStatusCommand creates the discovery queryflow status command that checks Discovery QueryFlow's health.
func NewStatusCommand(d cli.Discovery) *cobra.Command {
	status := &cobra.Command{
		Use:   "status",
		Short: "Check if Discovery QueryFlow is online",
		Long:  fmt.Sprintf(commands.LongStatusCheck, "QueryFlow"),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			queryflowClient := discoveryPackage.NewQueryFlow(vpr.GetString(profile+".queryflow_url"), vpr.GetString(profile+".queryflow_key"))
			return commands.StatusCheckCommand(d, queryflowClient.StatusChecker(), "QueryFlow", commands.GetCommandConfig(profile, vpr.GetString("output"), "QueryFlow", "queryflow_url"))
		},
		Args: cobra.NoArgs,
		Example: `	# Check the status of Discovery QueryFlow
	discovery queryflow status`,
	}

	return status
}
