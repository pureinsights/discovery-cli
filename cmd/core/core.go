package core

import (
	"github.com/pureinsights/pdp-cli/cmd/core/config"
	"github.com/pureinsights/pdp-cli/cmd/core/labels"
	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewCoreCommand creates the core command.
func NewCoreCommand(d cli.Discovery) *cobra.Command {
	core := &cobra.Command{
		Use:   "core [subcommand] [flags]",
		Short: "The main command to interact with Discovery's Core",
	}

	core.AddCommand(config.NewConfigCommand(d))
	core.AddCommand(labels.NewLabelCommand(d))

	return core
}
