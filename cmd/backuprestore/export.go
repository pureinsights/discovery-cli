package backuprestore

import (
	"github.com/pureinsights/discovery-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewExportCommand creates the discovery export command that exports all Discovery's entities.
func NewExportCommand(d cli.Discovery) *cobra.Command {
	var file string
	export := &cobra.Command{
		Use:   "export [subcommands]",
		Short: "Export all of Discovery's entities",
		Long:  "export is the command used to backup all of Discovery's entities at once. With the --file flag, the user can send the specific file in which to save the configurations. If not, they will be saved in a zip file in the current directory. The resulting zip file contains three zip files containing the entities of Discovery Core, Ingestion, and QueryFlow. If an export fails, the error is reported in the returned JSON.",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			err = commands.CheckCredentials(d, profile, "Core", "core_url")
			if err != nil {
				return err
			}

			err = commands.CheckCredentials(d, profile, "Ingestion", "ingestion_url")
			if err != nil {
				return err
			}

			err = commands.CheckCredentials(d, profile, "QueryFlow", "queryflow_url")
			if err != nil {
				return err
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

			printer := cli.GetObjectPrinter(d.Config().GetString("output"))
			return d.ExportEntitiesFromClients(clients, file, printer)
		},
		Args: cobra.NoArgs,
		Example: `	# Export the entities to a specific file.
	discovery export -p cn --file "entities/discovery.zip"`,
	}

	export.Flags().StringVarP(&file, "file", "f", "", "the file in which the information of the entities is going to be saved")
	return export
}
