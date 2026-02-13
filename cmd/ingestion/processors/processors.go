package processors

import (
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewProcessorCommand creates the processor command.
func NewProcessorCommand(d cli.Discovery) *cobra.Command {
	processor := &cobra.Command{
		Use:   "processor [subcommand] [flags]",
		Short: "The command to interact with Discovery Ingestion's processors.",
	}

	processor.AddCommand(NewStoreCommand(d))
	processor.AddCommand(NewGetCommand(d))

	return processor
}
