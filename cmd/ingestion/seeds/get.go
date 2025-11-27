package seeds

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/pureinsights/pdp-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/pdp-cli/discovery"
	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewGetCommand creates the seed get command
func NewGetCommand(d cli.Discovery) *cobra.Command {
	var (
		filters     []string
		recordId    string
		executionId string
		details     bool
	)
	get := &cobra.Command{
		Use:   "get",
		Short: "The command that obtains seeds from Discovery Ingestion.",
		Long:  fmt.Sprintf(commands.LongGetSearch, "seed", "Ingestion") + "The get command can also get records from the seed with the record flag. Finally, the get command can also get seed execution with the execution flag and with the details flag, the user can obtain more information about the execution.",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			ingestionClient := discoveryPackage.NewIngestion(vpr.GetString(profile+".ingestion_url"), vpr.GetString(profile+".ingestion_key"))
			if !cmd.Flags().Changed("record") && !cmd.Flags().Changed("execution") {
				return commands.SearchCommand(args, d, ingestionClient.Seeds(), commands.GetCommandConfig(profile, vpr.GetString("output"), "Ingestion", "ingestion_url"), &filters)
			}

			err = commands.CheckCredentials(d, profile, "Ingestion", "ingestion_url")
			if err != nil {
				return err
			}

			if len(args) <= 0 {
				return cli.NewError(cli.ErrorExitCode, "Missing the seed")
			}

			seed, err := cli.SearchEntity(d, ingestionClient.Seeds(), args[0])
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not search for entity with id %q", args[0])
			}

			seedId, err := uuid.Parse(seed.Get("id").String())
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get seed id")
			}

			output := vpr.GetString("output")
			if output == "json" {
				output = "pretty-json"
			}
			printer := cli.GetObjectPrinter(output)

			if cmd.Flags().Changed("record") {
				return d.AppendSeedRecord(seed, ingestionClient.Seeds().Records(seedId), recordId, printer)
			}

			seedExecutionId, err := uuid.Parse(executionId)
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get seed execution id")
			}

			seedExecutionClient := ingestionClient.Seeds().Executions(seedId)

			summarizers := map[string]cli.Summarizer{
				"records": seedExecutionClient.Records(seedExecutionId),
				"jobs":    seedExecutionClient.Jobs(seedExecutionId),
			}

			return d.GetSeedExecution(seedExecutionClient, seedExecutionId, summarizers, details, printer)
		},
		Args: cobra.MaximumNArgs(1),
		Example: `	# Get seed by name
	discovery ingestion seed get "Search seed"

	# Get seeds using filters
	discovery ingestion seed get --filter label=A:A -f type=staging

	# Get all seeds using the configuration in profile "cn"
	discovery ingestion seed get -p cn
	
	# Get a seed record by id
	discovery ingestion seed get 2acd0a61-852c-4f38-af2b-9c84e152873e --record A3HTDEgCa65BFZsac9TInFisvloRlL3M50ijCWNCKx0=
	
	# Get a seed execution by id and with details
	discovery ingestion seed get 2acd0a61-852c-4f38-af2b-9c84e152873e --execution 0f20f984-1854-4741-81ea-30f8b965b007 --details`,
	}

	get.Flags().StringArrayVarP(&filters, "filter", "f", []string{}, `apply filters in the format "filter=key:value". The available filters are:
- Label: The format is label={key}[:{value}], where the value is optional.
- Type: The format is type={type}.`)

	get.Flags().StringVar(&recordId, "record", "", "the id of the record that will be retrieved")
	get.Flags().StringVar(&executionId, "execution", "", "the id of the seed execution that will be retrieved")
	get.Flags().BoolVar(&details, "details", false, "gets more information when getting a seed execution, like the audited changes and record and job summaries")

	get.MarkFlagsMutuallyExclusive("filter", "record", "execution")
	get.MarkFlagsMutuallyExclusive("filter", "record", "details")
	return get
}
