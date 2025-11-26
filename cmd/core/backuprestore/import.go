package backuprestore

import (
	"fmt"

	"github.com/pureinsights/pdp-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/pdp-cli/discovery"
	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewImportCommand creates the discovery core import command that imports entities to Discovery Core
func NewImportCommand(d cli.Discovery) *cobra.Command {
	var onConflict string
	var file string
	importCmd := &cobra.Command{
		Use:   "import [subcommands]",
		Short: "Import entities to Discovery Core",
		Long:  fmt.Sprintf(commands.LongImport, "Core"),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
			}

			vpr := d.Config()

			coreClient := discoveryPackage.NewCore(vpr.GetString(profile+".core_url"), vpr.GetString(profile+".core_key"))
			return commands.ImportCommand(d, coreClient.BackupRestore(), file, discoveryPackage.OnConflict(onConflict), commands.GetCommandConfig(profile, vpr.GetString("output"), "Core", "core_url"))
		},
		Args: cobra.NoArgs,
		Example: `
	# Import the entities using profile "cn" and update conflict resolution strategy.
	# The rest of the command's output is omitted.
	discovery core import -p cn --file "entities/core.zip" --on-conflict UPDATE
	{
	"Credential": [
		{
		"id": "3b32e410-2f33-412d-9fb8-17970131921c",
		"status": 200
		},
		{
		"id": "458d245a-6ed2-4c2b-a73f-5540d550a479",
		"status": 200
		},
		{
		"id": "46cb4fff-28be-4901-b059-1dd618e74ee4",
		"status": 200
		},
		...`,
	}

	importCmd.Flags().StringVarP(&file, "file", "f", "", "the file that contains the entities that will be restored")
	importCmd.Flags().StringVar(&onConflict, "on-conflict", string(discoveryPackage.OnConflictFail), "the conflict resolution strategy that will be used")

	importCmd.MarkFlagRequired("file")

	return importCmd
}
