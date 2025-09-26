package config

import (
	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/spf13/cobra"
)

func NewConfigCommand(d cli.Discovery) *cobra.Command {
	return &cobra.Command{
		Use:   "config [subcommands]",
		Short: "Save Discovery's configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return d.SaveConfigFromUser(d.Config().GetString("profile"))
		},
	}
}
