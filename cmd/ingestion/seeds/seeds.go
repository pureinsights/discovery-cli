package seeds

import (
	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewSeedCommand creates the seed command.
func NewSeedCommand(d cli.Discovery) *cobra.Command {
	seed := &cobra.Command{
		Use:   "seed [subcommand] [flags]",
		Short: "The command to interact with Discovery Ingestion's seeds.",
	}

	seed.AddCommand(NewStoreCommand(d))

	return seed
}
