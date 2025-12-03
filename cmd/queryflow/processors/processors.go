package processors

import (
	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewProcessorCommand creates the processor command.
func NewProcessorCommand(d cli.Discovery) *cobra.Command {
	processor := &cobra.Command{
		Use:   "processor [subcommand] [flags]",
		Short: "The command to interact with Discovery QueryFlow's processors.",
	}

	processor.AddCommand(NewGetCommand(d))
	processor.AddCommand(NewStoreCommand(d))
    processor.AddCommand(NewDeleteCommand(d))

	return processor
}

	