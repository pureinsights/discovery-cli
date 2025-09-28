package config

import (
	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/spf13/cobra"
)

func NewGetCommand(d cli.Discovery) *cobra.Command {
	get := &cobra.Command{
		Use:   "get",
		Short: "Print Discovery's configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return err
			}
			sensitive, err := cmd.Flags().GetBool("sensitive")
			if err != nil {
				return err
			}
			return d.PrintConfigToUser(profile, sensitive)
		},
	}
	get.Flags().BoolP("sensitive", "s", true, "--sensitive=true")
	return get
}
