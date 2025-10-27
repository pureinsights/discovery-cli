package config

import (
	"fmt"

	"github.com/pureinsights/pdp-cli/cmd/commands"
	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewConfigCommand creates the discovery config command that saves the User's configuration
func NewConfigCommand(d cli.Discovery) *cobra.Command {
	config := &cobra.Command{
		Use:   "config [subcommands]",
		Short: "Save Discovery's configuration",
		Long:  fmt.Sprintf(commands.LongConfig, "Platform"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return commands.SaveConfigCommand(cmd, d.SaveConfigFromUser)
		},
	}
	config.AddCommand(NewGetCommand(d))
	return config
}
