package seeds

import (
	"github.com/pureinsights/pdp-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/pdp-cli/discovery"
	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
)

// NewGetCommand creates the seed get command
func NewStartCommand(d cli.Discovery) *cobra.Command {
	var scanType string
	var executionProperties string
	start := &cobra.Command{
		Use:   "start",
		Short: "The command that start a seed's  execution in Discovery Ingestion.",
		Long:  "start is the command used to start a seed execution in Discovery Ingestion. With the data flag, the user can send a single JSON configuration or an array to upsert multiple %[1]ss. With the file flag, the user can also send the address of a file that contains the JSON configurations. The data and file flags are required, but mutually exclusive.",
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
			scan := discoveryPackage.ScanType(scanType)
			propertiesJSON := gjson.Parse(executionProperties)
			printer := cli.GetObjectPrinter(vpr.GetString("output"))
			err = d.StartSeed(ingestionClient.Seeds(), args[0], scan, propertiesJSON, printer)
			return err
		},
		Args: cobra.ExactArgs(1),
	}

	start.Flags().StringVar(&scanType, "scan-type", string(discoveryPackage.ScanFull), "The scan type of the seed execution")
	start.Flags().StringVar(&executionProperties, "properties", "", "The execution properties of the seed execution")

	return start
}
