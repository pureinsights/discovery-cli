package config

import (
	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewConfigCommand creates the discovery config command that saves the User's configuration
func NewConfigCommand(d cli.Discovery) *cobra.Command {
	config := &cobra.Command{
		Use:   "config [subcommands]",
		Short: "Save Discovery Ingestion's configuration",
		Long:  "config is the command used to interact with Discovery Ingestion's configuration for a profile. This command by itself asks the user to save Ingestion's configuration for the given profile. The command prints the property to be modified along with its current value. If the property currently being shown is sensitive, its value is obfuscated. To keep the current value, the user must press \"Enter\" without any text, and to set the value as empty, a sole whitespace must be inputted.",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, err := cmd.Flags().GetString("profile")
			if err != nil {
				return err
			}
			return d.SaveIngestionConfigFromUser(profile, true)
		},
	}

	return config
}
