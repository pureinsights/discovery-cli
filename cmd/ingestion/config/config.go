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
		Short: "Save Discovery Ingestion's configuration",
		Long:  fmt.Sprintf(commands.LongConfig, "Ingestion"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return commands.SaveConfigCommand(cmd, d.IOStreams(), d.SaveIngestionConfigFromUser)
		},
		Example: `	# Ask the user for the configuration of profile "cn"
	discovery ingestion config -p cn
	Editing profile "cn". Press Enter to keep the value shown, type a single space to set empty.

	Ingestion URL [http://discovery.ingestion.cn]: https://discovery.ingestion.cn
	Ingestion API Key [*************.ingestion.cn]: `,
	}

	config.AddCommand(NewGetCommand(d))
	return config
}
