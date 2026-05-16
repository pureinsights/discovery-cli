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
		Long:  "status is the command to check the status of a seed. It can check the status of seed by its name or UUID. When the command only receives the seed, it returns the information of the last five seed executions and a summary of the records processed. If there are no executions, it shows an empty array. If there are no records, the records field is not included in the response. Also, just like the get command, it has the --execution and --details flags to get more information about a specific seed execution.",
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

			printer := cli.GetObjectPrinter(vpr.GetString("output"))

			ingestionClient := discoveryPackage.NewIngestion(vpr.GetString(profile+".ingestion_url"), vpr.GetString(profile+".ingestion_key"))
			seed, err := cli.SearchEntity(d, ingestionClient.Seeds(), args[0])
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not search for entity with id %q", args[0])
			}

			seedId, err := uuid.Parse(seed.Get("id").String())
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get seed id")
			}

			if !cmd.Flags().Changed("execution") {
				return d.StatusOfSeedExecutions(ingestionClient.Seeds().Executions(seedId), ingestionClient.Seeds().Records(seedId), printer)
			}

			return getSeedExecution(d, seedId, executionId, profile, details, printer)
		},
		Args: cobra.ExactArgs(1),
		Example: `	# Check the status of a seed
	discovery ingestion seed status "my-seed"
	
	# Check the status of a seed execution and with details
	discovery ingestion seed status "my-seed" --execution 0f20f984-1854-4741-81ea-30f8b965b007 --details`,
	}

	status.Flags().StringVar(&executionId, "execution", "", "the id of the seed execution that will be checked")
	status.Flags().BoolVar(&details, "details", false, "gets more information when checking the status of a seed execution, like the audited changes and record and job summaries")

	return status
}
