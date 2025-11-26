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
		Short: "Print Discovery Core's configuration",
		Long:  fmt.Sprintf(commands.LongConfigGet, "Core"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return commands.PrintConfigCommand(cmd, d.IOStreams(), d.PrintCoreConfigToUser)
		},
		Example: `
	# Print the configuration of the "cn" profile with obfuscated API keys.
	discovery core config get -p cn
	Showing the configuration of profile "cn":

	Core URL: "https://discovery.core.cn"
	Core API Key: "*************.core.cn"`,
	}
	get.Flags().BoolP("sensitive", "s", true, "this flag obfuscates sensitive values before showing them to the user.")
	return get
}
