package backuprestore

import (
	"fmt"

	"github.com/pureinsights/discovery-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewImportCommand creates the discovery core import command that imports entities to Discovery Core.
func NewImportCommand(d cli.Discovery) *cobra.Command {
	var onConflict string
	importCmd := &cobra.Command{
		Use:   "import <file> [subcommands]",
		Short: "Import entities to Discovery Core",
		Long:  fmt.Sprintf(commands.LongImport, "Core"),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			coreClient := discoveryPackage.NewCore(vpr.GetString(profile+".core_url"), vpr.GetString(profile+".core_key"))
			return commands.ImportCommand(d, coreClient.BackupRestore(), args[0], discoveryPackage.OnConflict(onConflict), commands.GetCommandConfig(profile, vpr.GetString("output"), "Core", "core_url"))
		},
		Args: cobra.ExactArgs(1),
		Example: `	# Import the entities using profile "cn" and update conflict resolution strategy.
	discovery core import -p cn "entities/core.zip" --on-conflict UPDATE`,
	}

	importCmd.Flags().StringVar(&onConflict, "on-conflict", string(discoveryPackage.OnConflictFail), "the conflict resolution strategy that will be used")

	return importCmd
}
