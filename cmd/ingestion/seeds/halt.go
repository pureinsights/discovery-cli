package seeds

import (
	"github.com/google/uuid"
	"github.com/pureinsights/pdp-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/pdp-cli/discovery"
	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewHaltCommand creates the seed halt command to halt a seed execution
func NewHaltCommand(d cli.Discovery) *cobra.Command {
	var execution string
	halt := &cobra.Command{
		Use:   "halt",
		Short: "The command that halts a seed execution in Discovery Ingestion.",
		Long:  "halt is the command used to halt a seed execution in Discovery Ingestion. With the execution flag, the user can specify the specific execution that will halted. If there is no execution flag, all of the active executions are halted.",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			err = commands.CheckCredentials(d, profile, "Ingestion", "ingestion_url", "ingestion_key")
			if err != nil {
				return err
			}

			vpr := d.Config()

			ingestionClient := discoveryPackage.NewIngestion(vpr.GetString(profile+".ingestion_url"), vpr.GetString(profile+".ingestion_key"))
			printer := cli.GetObjectPrinter(vpr.GetString("output"))
			if execution == "" {
				return d.HaltSeed(ingestionClient.Seeds(), args[0], printer)
			} else {
				if executionId, err := uuid.Parse(execution); err == nil {
					seedsClient := ingestionClient.Seeds()
					if seedId, err := cli.GetSeedId(d, seedsClient, args[0]); err == nil {
						return d.HaltSeedExecution(seedsClient.Executions(seedId), executionId, printer)
					}
				}

				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Failed to convert the execution ID to UUID")
			}
		},
		Args: cobra.ExactArgs(1),
	}

	halt.Flags().StringVarP(&execution, "execution", "e", "", "The execution properties of the seed execution")

	return halt
}
