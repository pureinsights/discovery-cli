package seeds

import (
	"github.com/pureinsights/discovery-cli/v2/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/v2/discovery"
	"github.com/pureinsights/discovery-cli/v2/internal/cli"
	"github.com/spf13/cobra"
)

// NewResetCommand creates the seed reset command to reset a seed metadata.
func NewResetCommand(d cli.Discovery) *cobra.Command {
	reset := &cobra.Command{
		Use:   "reset <seed>",
		Short: "The command that resets a seed metadata in Discovery Ingestion.",
		Long:  "reset is the command used to reset a seed metadata in Discovery Ingestion. If the seed has no active executions, then the seed's metadata is reset and its records deleted.",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			err = commands.CheckCredentials(d, profile, "Ingestion", "ingestion_url")
			if err != nil {
				return err
			}

			vpr := d.Config()

			ingestionClient := discoveryPackage.NewIngestion(vpr.GetString(profile+".ingestion_url"), vpr.GetString(profile+".ingestion_key"))

			return d.ResetSeed(ingestionClient.Seeds(), args[0], cli.GetObjectPrinter(vpr.GetString("output")))
		},
		Args: cobra.ExactArgs(1),
		Example: `	# Reset the seed metadata
	discovery ingestion seed reset 1d81d3d5-58a2-44a5-9acf-3fc8358afe09`,
	}

	return reset
}
