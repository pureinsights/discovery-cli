package commands

import (
	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/cli"
)

const (
	// LongExport is the message used in the Long field of the Export commands.
	LongExport string = "export is the command used to backup Discovery %s's entities. With the --file flag, the user can send the specific file in which to save the configurations. If not, they will be saved in a zip file in the current directory."
	// LongImport is the message used in the Long field of the Import commands.
	LongImport string = "import is the command used to restore Discovery %s's entities. With the --output-file flag, the user must send the specific file that has the entities' configuration. With the --on-conflict flag, the user can send the conflict resolution strategy in case there are duplicate entities."
)

// ExportCommand is the function that executes the export operation.
func ExportCommand(d cli.Discovery, client cli.BackupRestore, file string, config commandConfig) error {
	err := CheckCredentials(d, config.profile, config.componentName, config.url)
	if err != nil {
		return err
	}

	printer := cli.GetObjectPrinter(config.output)

	return d.ExportEntitiesFromClient(client, file, printer)
}

// ImportCommand is the function that executes the import operation.
func ImportCommand(d cli.Discovery, client cli.BackupRestore, file string, onConflict discoveryPackage.OnConflict, config commandConfig) error {
	err := CheckCredentials(d, config.profile, config.componentName, config.url)
	if err != nil {
		return err
	}

	printer := cli.GetObjectPrinter(config.output)

	return d.ImportEntitiesToClient(client, file, onConflict, printer)
}
