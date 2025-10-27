package config

import (
	"fmt"

	"github.com/pureinsights/pdp-cli/cmd/commands"
	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewGetCommand creates the config get command.
func NewGetCommand(d cli.Discovery) *cobra.Command {
	get := &cobra.Command{
		Use:   "get",
		Short: "Print Discovery's configuration",
		Long:  fmt.Sprintf(commands.LongConfig, "Platform"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return commands.PrintConfigCommand(cmd, d.PrintConfigToUser)
		},
	}
	get.Flags().BoolP("sensitive", "s", true, "--sensitive=true")
	return get
}
