package config

import (
	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/spf13/cobra"
)

// getCommandExecute is an auxiliary function was defined to be able to test the Get Command's RunE field.
func getCommandExecute(cmd *cobra.Command, d cli.Discovery) error {
	profile, err := cmd.Flags().GetString("profile")
	if err != nil {
		return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
	}
	sensitive, err := cmd.Flags().GetBool("sensitive")
	if err != nil {
		return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the sensitive flag")
	}

	return d.PrintConfigToUser(profile, sensitive)
}

// NewGetCommand creates the config get command.
func NewGetCommand(d cli.Discovery) *cobra.Command {
	get := &cobra.Command{
		Use:   "get",
		Short: "Print Discovery's configuration",
		Long:  "get is the command used to obtain Discovery's configuration for a given profile. If the API keys are sensitive, the `sensitive` flag can be set to true in order to obfuscate them before printing them out. If a configuration property was not set, it is not displayed.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return getCommandExecute(cmd, d)
		},
	}
	get.Flags().BoolP("sensitive", "s", true, "--sensitive=true")
	return get
}
