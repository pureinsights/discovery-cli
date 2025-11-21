package pipelines

import (
	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewPipelineCommand creates the pipeline command.
func NewPipelineCommand(d cli.Discovery) *cobra.Command {
	pipeline := &cobra.Command{
		Use:   "pipeline [subcommand] [flags]",
		Short: "The command to interact with Discovery Ingestion's pipelines.",
	}

	pipeline.AddCommand(NewStoreCommand(d))
	pipeline.AddCommand(NewGetCommand(d))

	return pipeline
}
