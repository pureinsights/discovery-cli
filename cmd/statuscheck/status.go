package statuscheck

import (
	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewStatusCommand creates the discovery status command that gets the status of every Discovery product.
func NewStatusCommand(d cli.Discovery) *cobra.Command {
	export := &cobra.Command{
		Use:   "status",
		Short: "Check if all of Discovery's products are online",
		Long:  "status is the command used to check the status of every Discovery product. If a product is healthy, it should return a JSON with an \"UP\" status field, which is added to a results JSON that matches the product to the received status response.",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			coreClient := discoveryPackage.NewCore(vpr.GetString(profile+".core_url"), vpr.GetString(profile+".core_key")).StatusChecker()
			ingestionClient := discoveryPackage.NewIngestion(vpr.GetString(profile+".ingestion_url"), vpr.GetString(profile+".ingestion_key")).StatusChecker()
			queryflowClient := discoveryPackage.NewQueryFlow(vpr.GetString(profile+".queryflow_url"), vpr.GetString(profile+".queryflow_key")).StatusChecker()
			stagingClient := discoveryPackage.NewStaging(vpr.GetString(profile+".staging_url"), vpr.GetString(profile+".staging_key")).StatusChecker()

			clients := []cli.StatusCheckClientEntry{
				{Name: "core", Client: coreClient},
				{Name: "ingestion", Client: ingestionClient},
				{Name: "queryflow", Client: queryflowClient},
				{Name: "staging", Client: stagingClient},
			}

			printer := cli.GetObjectPrinter(d.Config().GetString("output"))
			return d.StatusCheckOfClients(clients, printer)
		},
		Args: cobra.NoArgs,
		Example: `	# Check the status of Discovery
	discovery status`,
	}

	return export
}
