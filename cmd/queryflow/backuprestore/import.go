package backuprestore

import (
	"fmt"

	"github.com/pureinsights/discovery-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewImportCommand creates the discovery queryflow import command that imports entities to Discovery QueryFlow.
func NewImportCommand(d cli.Discovery) *cobra.Command {
	var onConflict string
	var file string
	importCmd := &cobra.Command{
		Use:   "import [subcommands]",
		Short: "Import entities to Discovery QueryFlow",
		Long:  fmt.Sprintf(commands.LongImport, "QueryFlow"),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			queryflowClient := discoveryPackage.NewQueryFlow(vpr.GetString(profile+".queryflow_url"), vpr.GetString(profile+".queryflow_key"))
			return commands.ImportCommand(d, queryflowClient.BackupRestore(), file, discoveryPackage.OnConflict(onConflict), commands.GetCommandConfig(profile, vpr.GetString("output"), "QueryFlow", "queryflow_url"))
		},
		Args: cobra.NoArgs,
		Example: `	# Import the entities using profile "cn" and fail conflict resolution strategy.
	discovery queryflow import -p cn --file "entities/queryflow.zip"`,
	}

	importCmd.Flags().StringVarP(&file, "file", "f", "", "the file that contains the entities that will be restored")
	importCmd.Flags().StringVar(&onConflict, "on-conflict", string(discoveryPackage.OnConflictFail), "the conflict resolution strategy that will be used")

	importCmd.MarkFlagRequired("file")

	return importCmd
}
