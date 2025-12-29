package staging

import (
	"github.com/pureinsights/discovery-cli/cmd/staging/config"
	"github.com/pureinsights/discovery-cli/cmd/staging/statuscheck"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewStagingCommand creates the staging command.
func NewStagingCommand(d cli.Discovery) *cobra.Command {
	staging := &cobra.Command{
		Use:   "staging [subcommand] [flags]",
		Short: "The main command to interact with Discovery's Staging",
	}

	staging.AddCommand(config.NewConfigCommand(d))
	staging.AddCommand(statuscheck.NewStatusCommand(d))

	return staging
}
