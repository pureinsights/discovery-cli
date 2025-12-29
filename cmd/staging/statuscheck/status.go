package statuscheck

import (
	"fmt"

	"github.com/pureinsights/discovery-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewStatusCommand creates the discovery staging status command that checks Discovery Staging's health.
func NewStatusCommand(d cli.Discovery) *cobra.Command {
	status := &cobra.Command{
		Use:   "status",
		Short: "Check if Discovery Staging is online",
		Long:  fmt.Sprintf(commands.LongStatusCheck, "Staging"),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			stagingClient := discoveryPackage.NewStaging(vpr.GetString(profile+".staging_url"), vpr.GetString(profile+".staging_key"))
			return commands.StatusCheckCommand(d, stagingClient.StatusChecker(), "Staging", commands.GetCommandConfig(profile, vpr.GetString("output"), "Staging", "staging_url"))
		},
		Args: cobra.NoArgs,
		Example: `	# Check the status of Discovery Staging
	discovery staging status`,
	}

	return status
}
