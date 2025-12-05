package config

import (
	"fmt"

	"github.com/pureinsights/discovery-cli/cmd/commands"
	"github.com/pureinsights/discovery-cliry-cliry-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewGetCommand creates the config get command.
func NewGetCommand(d cli.Discovery) *cobra.Command {
	get := &cobra.Command{
		Use:   "get",
		Short: "Print Discovery Ingestion's configuration",
		Long:  fmt.Sprintf(commands.LongConfigGet, "Ingestion"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return commands.PrintConfigCommand(cmd, d.IOStreams(), d.PrintIngestionConfigToUser)
		},
	}
	get.Flags().BoolP("sensitive", "s", true, "this flag obfuscates sensitive values before showing them to the user.")
	return get
}
