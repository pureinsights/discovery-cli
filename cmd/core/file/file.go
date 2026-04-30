package file

import (
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewFileCommand creates the file command.
func NewFileCommand(d cli.Discovery) *cobra.Command {
	file := &cobra.Command{
		Use:   "file [subcommand] [flags]",
		Short: "The command to interact with Discovery Core's files.",
	}

	file.AddCommand(NewGetCommand(d))
	file.AddCommand(NewDownloadCommand(d))
	file.AddCommand(NewStoreCommand(d))
	// credential.AddCommand(NewDeleteCommand(d))

	return file
}
