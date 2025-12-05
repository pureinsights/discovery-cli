package seeds

import (
	"github.com/pureinsights/discovery-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
)

// NewStartCommand creates the seed start command to start a seed execution
func NewStartCommand(d cli.Discovery) *cobra.Command {
	var scanType string
	var executionProperties string
	start := &cobra.Command{
		Use:   "start",
		Short: "The command that starts a seed execution in Discovery Ingestion.",
		Long:  "start is the command used to start a seed execution in Discovery Ingestion. With the properties flag, the user set the execution properties with which to run the seed. With the scan-type flag, the user can set the scan type of the execution: FULL or INCREMENTAL.",
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
			scan := discoveryPackage.ScanType(scanType)
			propertiesJSON := gjson.Parse(executionProperties)
			printer := cli.GetObjectPrinter(vpr.GetString("output"))
			return d.StartSeed(ingestionClient.Seeds(), args[0], scan, propertiesJSON, printer)
		},
		Args: cobra.ExactArgs(1),
		Example: `# Start a seed seed execution with the properties and scan-type flags
	discovery ingestion seed start --scan-type FULL --properties '{"stagingBucket":"testBucket"}' 0ce1bece-5a01-4d4a-bf92-5ca3cd5327f3`,
	}

	start.Flags().StringVar(&scanType, "scan-type", string(discoveryPackage.ScanFull), "the scan type of the seed execution")
	start.Flags().StringVar(&executionProperties, "properties", "", "the execution properties of the seed execution")

	return start
}
