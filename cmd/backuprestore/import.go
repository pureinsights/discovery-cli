package backuprestore

import (
	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewImportCommand creates the discovery import command that imports entitites to Discovery Core, Ingestion, and QueryFlow
func NewImportCommand(d cli.Discovery) *cobra.Command {
	var onConflict string
	var file string
	importCmd := &cobra.Command{
		Use:   "import [subcommands]",
		Short: "import entities to all of Discovery's products",
		Long:  "import is the command used to restore entities to all of Discovery's products at once. With the --file flag, the user must send the specific file that has the entities' configuration. This file is a compressed zip file that contains the zip files product by the /export endpoint in a Discovery product. It should have at most three zip files: one for Core, one for Ingestion, and a final one for QueryFlow. The export file for a Discovery product has the format productName-*. For example, the Core can be called core-export-20251112T1629.zip and the one for Ingestion can be called ingestion-export-20251110T1607.zip. The sent file does not need to contain the export files for all of Discovery's products. This command can restore entities to one, two, or all products. With the on-conflict flag, the user can send the conflict resolution strategy in case there are duplicate entities.",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			coreClient := discoveryPackage.NewCore(vpr.GetString(profile+".core_url"), vpr.GetString(profile+".core_key")).BackupRestore()
			ingestionClient := discoveryPackage.NewIngestion(vpr.GetString(profile+".ingestion_url"), vpr.GetString(profile+".ingestion_key")).BackupRestore()
			queryflowClient := discoveryPackage.NewQueryFlow(vpr.GetString(profile+".queryflow_url"), vpr.GetString(profile+".queryflow_key")).BackupRestore()

			clients := []cli.BackupRestoreClientEntry{
				{Name: "core", Client: coreClient},
				{Name: "ingestion", Client: ingestionClient},
				{Name: "queryflow", Client: queryflowClient},
			}

			output := d.Config().GetString("output")
			if output == "json" {
				output = "pretty-json"
			}
			printer := cli.GetObjectPrinter(output)
			return d.ImportEntitiesToClients(clients, file, discoveryPackage.OnConflict(onConflict), printer)
		},
		Args: cobra.NoArgs,
		Example: `	# Import the entities to Discovery Core and Ingestion using profile "cn" and ignore conflict resolution strategy.
	discovery import -p cn --file "entities/discovery.zip" --on-conflict IGNORE`,
	}

	importCmd.Flags().StringVarP(&file, "file", "f", "", "the file that contains the files with the exported entities of the Discovery products.")
	importCmd.Flags().StringVar(&onConflict, "on-conflict", string(discoveryPackage.OnConflictFail), "the conflict resolution strategy that will be used")

	importCmd.MarkFlagRequired("file")

	return importCmd
}
