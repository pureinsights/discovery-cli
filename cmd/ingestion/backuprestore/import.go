package backuprestore

import (
	"fmt"

	"github.com/pureinsights/discovery-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewImportCommand creates the discovery ingestion import command that imports entities to Discovery Ingestion.
func NewImportCommand(d cli.Discovery) *cobra.Command {
	var onConflict string
	importCmd := &cobra.Command{
		Use:   "import <file>",
		Short: "Import entities to Discovery Ingestion",
		Long:  fmt.Sprintf(commands.LongImport, "Ingestion"),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			ingestionClient := discoveryPackage.NewIngestion(vpr.GetString(profile+".ingestion_url"), vpr.GetString(profile+".ingestion_key"))
			return commands.ImportCommand(d, ingestionClient.BackupRestore(), args[0], discoveryPackage.OnConflict(onConflict), commands.GetCommandConfig(profile, vpr.GetString("output"), "Ingestion", "ingestion_url"))
		},
		Args: cobra.ExactArgs(1),
		Example: `	# Import the entities using profile "cn" and ignore conflict resolution strategy.
	discovery ingestion import -p cn "entities/ingestion.zip" --on-conflict IGNORE`,
	}

	importCmd.Flags().StringVar(&onConflict, "on-conflict", string(discoveryPackage.OnConflictFail), "the conflict resolution strategy that will be used")

	return importCmd
}
