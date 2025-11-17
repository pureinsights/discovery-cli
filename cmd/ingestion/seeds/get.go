package seeds

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/pureinsights/pdp-cli/cmd/commands"
	"github.com/pureinsights/pdp-cli/discovery"
	discoveryPackage "github.com/pureinsights/pdp-cli/discovery"
	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tidwall/gjson"
)

// NewGetCommand creates the seed get command
func NewGetCommand(d cli.Discovery) *cobra.Command {
	var (
		filters     []string
		recordId    string
		executionId string
		details     bool
		records     bool
	)
	get := &cobra.Command{
		Use:   "get",
		Short: "The command that obtains seeds from Discovery Ingestion.",
		Long:  fmt.Sprintf(commands.LongGetSearch, "seed", "Ingestion"),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			if cmd.Flags().Changed("record") && recordId == "" {
				records = true
			}

			vpr := d.Config()

			ingestionClient := discoveryPackage.NewIngestion(vpr.GetString(profile+".ingestion_url"), vpr.GetString(profile+".ingestion_key"))
			if !cmd.Flags().Changed("record") {
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

			output := d.Config().GetString("output")
			if output == "json" && (cmd.Flags().Changed("record") || cmd.Flags().Changed("seed-execution")) {
				output = "pretty-json"
			}
			printer := cli.GetObjectPrinter(output)

			if cmd.Flags().Changed("record") {
				return d.AppendSeedRecord(seed, ingestionClient.Seeds().Records(seedId), recordId, printer)
			}
			printer := cli.GetObjectPrinter(output)

			seedExecutionId, err := uuid.Parse(executionId)
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get seed id")
			}

			seedExecutionClient := ingestionClient.Seeds().Executions(seedId)

			summarizers := map[string]cli.Summarizer{
				"records": seedExecutionClient.Records(seedExecutionId),
				"jobs":    seedExecutionClient.Jobs(seedExecutionId),
			}

			return d.GetSeedExecution(seedExecutionClient, seedExecutionId, summarizers, details, printer)
		},
		Args: cobra.MaximumNArgs(1),
	}

	get.Flags().StringArrayVarP(&filters, "filter", "f", []string{}, `Apply filters in the format "filter=key:value". The available filters are:
- Label: The format is label={key}[:{value}], where the value is optional.
- Type: The format is type={type}.`)

	get.Flags().StringVar(&recordId, "record", "", "the id of the record that will be retrieved")
	get.Flags().StringVar(&executionId, "seed-execution", "", "the id of the seed exectuion that will be retrieved")
	get.Flags().BoolVar(&details, "details", false, "flag documentation")

	get.MarkFlagsMutuallyExclusive("filter", "record", "seed-execution")
	return get
}
