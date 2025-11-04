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
		Long:  "halt is the command used to halt a seed execution in Discovery Ingestion. With the properties flag, the user set the execution properties with which to run the seed. With the scan-type flag, the user can set the scan type of the execution: FULL or INCREMENTAL.",
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
				err = d.HaltSeed(ingestionClient.Seeds(), args[0], printer)
			} else {
				executionId := uuid.Parse(execution)
				err = d.HaltSeedExecution(ingestionClient.Seeds(), args[0], printer)
			}
			return err
		},
		Args: cobra.ExactArgs(1),
	}

	halt.Flags().StringVarP(&execution, "execution", "e", "", "The execution properties of the seed execution")

	return halt
}
