package ingestion

import (
	"github.com/pureinsights/discovery-cli/cmd/ingestion/config"
	"github.com/pureinsights/discovery-cli/cmd/ingestion/pipelines"
	"github.com/pureinsights/discovery-cli/cmd/ingestion/processors"
	"github.com/pureinsights/discovery-cli/cmd/ingestion/seeds"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewIngestionCommand creates the ingestion command.
func NewIngestionCommand(d cli.Discovery) *cobra.Command {
	ingestion := &cobra.Command{
		Use:   "ingestion [subcommand] [flags]",
		Short: "The main command to interact with Discovery's Ingestion",
	}

	ingestion.AddCommand(config.NewConfigCommand(d))
	ingestion.AddCommand(processors.NewProcessorCommand(d))
	ingestion.AddCommand(pipelines.NewPipelineCommand(d))
	ingestion.AddCommand(seeds.NewSeedCommand(d))

	return ingestion
}
