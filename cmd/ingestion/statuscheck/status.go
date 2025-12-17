package statuscheck

import (
	"fmt"

	"github.com/pureinsights/discovery-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewStatusCommand creates the discovery ingestion status command that checks Discovery Ingestion's health.
func NewStatusCommand(d cli.Discovery) *cobra.Command {
	status := &cobra.Command{
		Use:   "status",
		Short: "Check if Discovery Ingestion is online",
		Long:  fmt.Sprintf(commands.LongStatusCheck, "Ingestion"),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			ingestionClient := discoveryPackage.NewIngestion(vpr.GetString(profile+".ingestion_url"), vpr.GetString(profile+".ingestion_key"))
			return commands.StatusCheckCommand(d, ingestionClient.StatusChecker(), "Ingestion", commands.GetCommandConfig(profile, vpr.GetString("output"), "Ingestion", "ingestion_url"))
		},
		Args: cobra.NoArgs,
		Example: `	# Check the status of Discovery Ingestion
	discovery ingestion status`,
	}

	return status
}
