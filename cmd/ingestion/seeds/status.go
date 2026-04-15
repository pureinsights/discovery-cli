package seeds

import (
	"github.com/google/uuid"
	"github.com/pureinsights/discovery-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewStatusCommand creates the seed status command
func NewStatusCommand(d cli.Discovery) *cobra.Command {
	var (
		executionId string
		details     bool
	)
	status := &cobra.Command{
		Use:   "status <seed>",
		Short: "The command that gets the status of seed executions.",
		Long:  "",
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

			if !cmd.Flags().Changed("execution") {
				return nil
			}

			ingestionClient := discoveryPackage.NewIngestion(vpr.GetString(profile+".ingestion_url"), vpr.GetString(profile+".ingestion_key"))
			seed, err := cli.SearchEntity(d, ingestionClient.Seeds(), args[0])
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not search for entity with id %q", args[0])
			}

			seedId, err := uuid.Parse(seed.Get("id").String())
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get seed id")
			}

			output := vpr.GetString("output")
			printer := cli.GetObjectPrinter(output)

			return getSeedExecution(d, seedId, executionId, profile, details, printer)
		},
		Args:    cobra.ExactArgs(1),
		Example: ``,
	}

	status.Flags().StringVar(&executionId, "execution", "", "the id of the seed execution that will be checked")
	status.Flags().BoolVar(&details, "details", false, "gets more information when checking the status of a seed execution, like the audited changes and record and job summaries")

	return status
}
