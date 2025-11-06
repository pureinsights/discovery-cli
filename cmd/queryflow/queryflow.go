package queryflow

import (
	"github.com/pureinsights/pdp-cli/cmd/queryflow/config"
	"github.com/pureinsights/pdp-cli/cmd/queryflow/endpoints"
	"github.com/pureinsights/pdp-cli/cmd/queryflow/processors"
	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewQueryFlowCommand creates the queryFlow command.
func NewQueryFlowCommand(d cli.Discovery) *cobra.Command {
	queryflow := &cobra.Command{
		Use:   "queryflow [subcommand] [flags]",
		Short: "The main command to interact with Discovery's Queryflow",
	}

	queryflow.AddCommand(config.NewConfigCommand(d))
	queryflow.AddCommand(processors.NewProcessorCommand(d))
	queryflow.AddCommand(endpoints.NewEndpointCommand(d))

	return queryflow
}
