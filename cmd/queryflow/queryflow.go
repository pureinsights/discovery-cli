package queryflow

import (
	"github.com/pureinsights/discovery-cli/cmd/queryflow/config"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewQueryFlowCommand creates the queryFlow command.
func NewQueryFlowCommand(d cli.Discovery) *cobra.Command {
	queryflow := &cobra.Command{
		Use:   "queryflow [subcommand] [flags]",
		Short: "The main command to interact with Discovery's Queryflow",
	}

	queryflow.AddCommand(config.NewConfigCommand(d))

	return queryflow
}
