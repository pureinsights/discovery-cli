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

	// credential.AddCommand(NewGetCommand(d))
	// credential.AddCommand(NewStoreCommand(d))
	// credential.AddCommand(NewDeleteCommand(d))

	return file
}
