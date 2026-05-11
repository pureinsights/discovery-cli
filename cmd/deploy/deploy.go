package deploy

import (
	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewDeployCommand creates the discovery deploy command that imports entities in a directory to Discovery Core, Ingestion, and QueryFlow.
func NewDeployCommand(d cli.Discovery) *cobra.Command {
	deployCmd := &cobra.Command{
		Use:   "deploy <path>",
		Short: "deploy entities in a directory to all of Discovery's products",
		Long: `deploy is a command that restores entities into Discovery if the user does not have the required export zip file. This command receives a directory or folder that must have a specific, but simple structure:
directory
├── core
│   ├── server
│   ├── credential
│   └── file
│
├── ingestion
│   ├── pipeline
│   ├── processor
│   ├── seed
│   └── seed-schedule
│
└── queryflow	
    ├── pipeline
    ├── processor
    └── entrypoint
	    └── endpoint

The entity directories have JSON files with the configurations that will be imported. Inside these directories, entities can be further divided into subdirectories if desired, but the Discovery product directories must have this structure. The file folder is optional. If present, it uploads those files into Discovery Core's object storage. The command will fail if any file upload is unsuccessful. The file folder can be in any of the product directories, it does not need to be in Core's directory. The command reads each entity's JSON configuration and creates the zip files needed to import them into Core, Ingestion, and QueryFlow. The entities do not need to exist yet in Discovery in order to store them. Entities that already exist are updated. If a Discovery product's entities do not show up in the results JSON, then they could not be read or are not included in the directory.`,
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
	discovery deploy -p cn "discovery entities"`,
	}

	return deployCmd
}
