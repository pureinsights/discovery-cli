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
		Long:  fmt.Sprintf(commands.LongConfigGet, "Platform"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return commands.PrintConfigCommand(cmd, d.IOStreams(), d.PrintConfigToUser)
		},
		Example: `	# Print the configuration of the "cn" profile with obfuscated API keys.
	discovery config get -p cn`,
	}
	get.Flags().BoolP("sensitive", "s", true, "this flag obfuscates sensitive values before showing them to the user.")
	return get
}
