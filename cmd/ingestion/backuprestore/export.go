package backuprestore

import (
	"fmt"

	"github.com/pureinsights/discovery-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewExportCommand creates the discovery ingestion export command that exports Discovery Ingestion's entities.
func NewExportCommand(d cli.Discovery) *cobra.Command {
	var file string
	export := &cobra.Command{
		Use:   "export",
		Short: "Export all of Discovery Ingestion's entities",
		Long:  fmt.Sprintf(commands.LongExport, "Ingestion"),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			ingestionClient := discoveryPackage.NewIngestion(vpr.GetString(profile+".ingestion_url"), vpr.GetString(profile+".ingestion_key"))
			return commands.ExportCommand(d, ingestionClient.BackupRestore(), file, commands.GetCommandConfig(profile, vpr.GetString("output"), "Ingestion", "ingestion_url"))
		},
		Args: cobra.NoArgs,
		Example: `	# Export the entities to a specific file.
	discovery ingestion export -p cn --file "entities/ingestion.zip"`,
	}

	export.Flags().StringVarP(&file, "file", "f", "", "the file that will contain the exported entities")
	return export
}
