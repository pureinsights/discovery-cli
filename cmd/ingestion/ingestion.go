package ingestion

import (
	"github.com/pureinsights/pdp-cli/cmd/core/config"
	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewIngestionCommand creates the ingestion command.
func NewIngestionCommand(d cli.Discovery) *cobra.Command {
	core := &cobra.Command{
		Use:   "ingestion [subcommand] [flags]",
		Short: "The main command to interact with Discovery's Ingestion",
	}

	core.AddCommand(config.NewConfigCommand(d))

	return core
}
