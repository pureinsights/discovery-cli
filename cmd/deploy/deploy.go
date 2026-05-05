package deploy

import (
	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewDeployCommand creates the discovery deploy command that deploys entities to Discovery Core, Ingestion, and QueryFlow.
func NewDeployCommand(d cli.Discovery) *cobra.Command {
	deployCmd := &cobra.Command{
		Use:   "deploy <folder>",
		Short: "deploy entities in a directory to all of Discovery's products",
		Long: `deploy is a command to restore entities into Discovery if the user does not have the required export file. This command receives a directory or folder that must have a specific, but simple structure:
--> core
|---> server
|---> credential
|---> files
--> ingestion
|---> pipeline
|---> processor
|---> seed
|---> seedSchedule
--> queryflow
|---> endpoint
|---> pipeline
|---> processor
Inside these directories, the entities can be themselves divided into other subdirectories if desired, but the Discovery product directories must have this structure. The files folder is optional. If present, it uploads those files into Discovery Core's object storage. The command reads each entity's JSON configuration and creates the zip files needed to import them into Core, Ingestion, and QueryFlow. The entities do not need to exist yet in Discovery in order to store them. If a Discovery product's entities do not show up in the results JSON, then they could not be read.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			core := discoveryPackage.NewCore(vpr.GetString(profile+".core_url"), vpr.GetString(profile+".core_key"))
			coreClient := core.BackupRestore()
			ingestionClient := discoveryPackage.NewIngestion(vpr.GetString(profile+".ingestion_url"), vpr.GetString(profile+".ingestion_key")).BackupRestore()
			queryflowClient := discoveryPackage.NewQueryFlow(vpr.GetString(profile+".queryflow_url"), vpr.GetString(profile+".queryflow_key")).BackupRestore()

			clients := []cli.BackupRestoreClientEntry{
				{Name: "core", Client: coreClient},
				{Name: "ingestion", Client: ingestionClient},
				{Name: "queryflow", Client: queryflowClient},
			}

			printer := cli.GetObjectPrinter(d.Config().GetString("output"))
			return d.Deploy(core.Files(), clients, args[0], printer)
		},
		Args: cobra.ExactArgs(1),
		Example: `	# Deploy the entities to the Discovery products using profile "cn".
	discovery deploy -p cn "entities_folder"`,
	}

	return deployCmd
}
