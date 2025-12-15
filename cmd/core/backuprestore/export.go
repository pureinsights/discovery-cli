package backuprestore

import (
	"fmt"

	"github.com/pureinsights/discovery-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewExportCommand creates the discovery core export command that exports Discovery Core's entities.
func NewExportCommand(d cli.Discovery) *cobra.Command {
	var file string
	export := &cobra.Command{
		Use:   "export [subcommands]",
		Short: "Export all of Discovery Core's entities",
		Long:  fmt.Sprintf(commands.LongExport, "Core"),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			coreClient := discoveryPackage.NewCore(vpr.GetString(profile+".core_url"), vpr.GetString(profile+".core_key"))
			return commands.ExportCommand(d, coreClient.BackupRestore(), file, commands.GetCommandConfig(profile, vpr.GetString("output"), "Core", "core_url"))
		},
		Args: cobra.NoArgs,
		Example: `	# Export the entities to a specific file.
	discovery core export -p cn --file "entities/core.zip`,
	}

	export.Flags().StringVarP(&file, "file", "f", "", "the file that will contain the exported entities")
	return export
}
