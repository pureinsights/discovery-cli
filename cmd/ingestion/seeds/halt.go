package seeds

import (
	"github.com/google/uuid"
	"github.com/pureinsights/discovery-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewHaltCommand creates the seed halt command to halt a seed execution.
func NewHaltCommand(d cli.Discovery) *cobra.Command {
	var execution string
	halt := &cobra.Command{
		Use:   "halt",
		Short: "The command that halts a seed execution in Discovery Ingestion.",
		Long:  "halt is the command used to halt a seed execution in Discovery Ingestion. With the --execution flag, the user can specify the specific execution that will be halted. If there is no --execution flag, all of the active executions are halted.",
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
			printer := cli.GetObjectPrinter(vpr.GetString("output"))
			if execution == "" {
				return d.HaltSeed(ingestionClient.Seeds(), args[0], printer)
			}

			executionId, err := uuid.Parse(execution)
			if err == nil {
				seedsClient := ingestionClient.Seeds()
				seedId, err := cli.GetEntityId(d, seedsClient, args[0])
				if err == nil {
					return d.HaltSeedExecution(seedsClient.Executions(seedId), executionId, printer)
				}
			}

			return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Failed to convert the execution ID to UUID")
		},
		Args: cobra.ExactArgs(1),
		Example: `	# Halt all active seed executions
	discovery ingestion seed halt 0ce1bece-5a01-4d4a-bf92-5ca3cd5327f3

	# Halt a single seed execution
	discovery ingestion seed halt 1d81d3d5-58a2-44a5-9acf-3fc8358afe09 --execution f63fbdb6-ec49-4fe5-90c9-f5c6de4efc36`,
	}

	halt.Flags().StringVarP(&execution, "execution", "e", "", "the UUID of the execution that will be halted")

	return halt
}
