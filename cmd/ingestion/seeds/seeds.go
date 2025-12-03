package seeds

import (
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewSeedCommand creates the seed command.
func NewSeedCommand(d cli.Discovery) *cobra.Command {
	seed := &cobra.Command{
		Use:   "seed [subcommand] [flags]",
		Short: "The command to interact with Discovery Ingestion's seeds.",
	}

	seed.AddCommand(NewStoreCommand(d))
	seed.AddCommand(NewGetCommand(d))
	seed.AddCommand(NewStartCommand(d))
	seed.AddCommand(NewHaltCommand(d))
    seed.AddCommand(NewDeleteCommand(d))

	return seed
}
	