package core

import (
	"github.com/pureinsights/pdp-cli/cmd/core/config"
	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/spf13/cobra"
)

func NewCoreCommand(d cli.Discovery) *cobra.Command {
	core := &cobra.Command{
		Use:   "core [subcommand] [flags]",
		Short: "The main command to interact with Discovery's Core",
	}

	core.AddCommand(config.NewConfigCommand(d))

	return core
}
