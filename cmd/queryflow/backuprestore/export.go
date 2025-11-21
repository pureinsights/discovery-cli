package backuprestore

import (
	"fmt"

	"github.com/pureinsights/pdp-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/pdp-cli/discovery"
	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewExportCommand creates the discovery queryflow export command that exports Discovery QueryFlow's entities
func NewExportCommand(d cli.Discovery) *cobra.Command {
	var file string
	export := &cobra.Command{
		Use:   "export [subcommands]",
		Short: "Export all of Discovery QueryFlow's entities",
		Long:  fmt.Sprintf(commands.LongExport, "QueryFlow"),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			queryflowClient := discoveryPackage.NewQueryFlow(vpr.GetString(profile+".queryflow_url"), vpr.GetString(profile+".queryflow_key"))
			return commands.ExportCommand(d, queryflowClient.BackupRestore(), file, commands.GetCommandConfig(profile, vpr.GetString("output"), "QueryFlow", "queryflow_url"))
		},
		Args: cobra.NoArgs,
	}

	export.Flags().StringVarP(&file, "file", "f", "", "the file that will contain the exported entities")
	return export
}
