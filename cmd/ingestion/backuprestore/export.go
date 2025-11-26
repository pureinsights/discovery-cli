package backuprestore

import (
	"fmt"

	"github.com/pureinsights/pdp-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/pdp-cli/discovery"
	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewExportCommand creates the discovery ingestion export command that exports Discovery Ingestion's entities
func NewExportCommand(d cli.Discovery) *cobra.Command {
	var file string
	export := &cobra.Command{
		Use:   "export [subcommands]",
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
		Example: `	# Import the entities using profile "cn" and ignore conflict resolution strategy.
	# The rest of the command's output is omitted.
	discovery ingestion import -p cn --file "entities/ingestion.zip" --on-conflict IGNORE
	{
	"Pipeline": [
		{
		"id": "0d3f476d-9003-4fc8-b9a9-8ba6ebf9445b",
		"status": 204
		},
		{
		"id": "25012a20-fe60-4ad6-a05c-9abcbfc1dfb1",
		"status": 204
		},
		{
		"id": "36f8ce72-f23d-4768-91e8-58693ff1b272",
		"status": 204
		},
		...`,
	}

	export.Flags().StringVarP(&file, "file", "f", "", "the file that will contain the exported entities")
	return export
}
