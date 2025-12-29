package pipelines

import (
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewPipelineCommand creates the pipeline command.
func NewPipelineCommand(d cli.Discovery) *cobra.Command {
	pipeline := &cobra.Command{
		Use:   "pipeline [subcommand] [flags]",
		Short: "The command to interact with Discovery QueryFlow's pipelines.",
	}

	pipeline.AddCommand(NewGetCommand(d))

	return pipeline
}
