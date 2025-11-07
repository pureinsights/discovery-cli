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
	var filters []string
	get := &cobra.Command{
		Use:   "get",
		Short: "The command that obtains seeds from Discovery Ingestion.",
		Long:  fmt.Sprintf(commands.LongGetSearch, "seed", "Ingestion"),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			ingestionClient := discoveryPackage.NewIngestion(vpr.GetString(profile+".ingestion_url"), vpr.GetString(profile+".ingestion_key"))

			if !cmd.Flags().Changed("records") && !cmd.Flags().Changed("record") {
				return commands.SearchCommand(args, d, ingestionClient.Seeds(), commands.GetCommandConfig(profile, vpr.GetString("output"), "Ingestion", "ingestion_url", "ingestion_key"), &filters)
			}

			if len(args) <= 0 {
				return cli.NewError(cli.ErrorExitCode, "Missing the seed")
			}

			printer := cli.GetObjectPrinter(vpr.GetString("output"))

			seed, err := cli.SearchEntity(d, ingestionClient.Seeds(), args[0])
			if err != nil {
				return err
			}

			seedId, err := uuid.Parse(seed.Get("id").String())
			if err != nil {
				return err
			}

			if cmd.Flags().Changed("record") {
				record, err := cmd.Flags().GetString("record")
				if err != nil {
					return err
				}
				return d.AppendSeedRecord(seed, ingestionClient.Seeds().Records(seedId), record, printer)
			}

			return d.AppendSeedRecords(seed, ingestionClient.Seeds().Records(seedId), printer)
		},
		Args: cobra.MaximumNArgs(1),
	}

	get.Flags().StringArrayVarP(&filters, "filter", "f", []string{}, `Apply filters in the format "filter=key:value". The available filters are:
- Label: The format is label={key}[:{value}], where the value is optional.
- Type: The format is type={type}.`)

	get.Flags().Bool("records", false, "Show the records")
	get.Flags().String("record", "", "the id of the record")

	get.MarkFlagsMutuallyExclusive("filter", "records", "record")

	return get
}
