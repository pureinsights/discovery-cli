package backuprestore

import (
	"fmt"

	"github.com/pureinsights/pdp-cli/cmd/commands"
	discoveryPackage "github.com/pureinsights/pdp-cli/discovery"
	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewImportCommand creates the discovery import command that imports entitites to Discovery Core, Ingestion, and QueryFlow
func NewImportCommand(d cli.Discovery) *cobra.Command {
	var onConflict string
	var file string
	importCmd := &cobra.Command{
		Use:   "import [subcommands]",
		Short: "import entities to all of Discovery's products",
		Long:  fmt.Sprintf(commands.LongImport, "Core"),
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
	}

	importCmd.Flags().StringVarP(&file, "file", "f", "", "the file that contains the entities that will be restored")
	importCmd.Flags().StringVar(&onConflict, "on-conflict", string(discoveryPackage.OnConflictFail), "the conflict resolution strategy that will be used")

	importCmd.MarkFlagRequired("file")

	return importCmd
}
