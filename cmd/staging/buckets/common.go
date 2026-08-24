package buckets

import (
	"github.com/pureinsights/discovery-cli/cmd/commands"
	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/spf13/cobra"
)

// Fetches profile flag, verifies Staging credentials, and builds the output printer
func prepareStagingCommand(d cli.Discovery, cmd *cobra.Command) (string, cli.Printer, error) {
	profile, err := cmd.Flags().GetString("profile")
	if err != nil {
		return "", nil, cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
	}

	if err := commands.CheckCredentials(d, profile, "Staging", "staging_url"); err != nil {
		return "", nil, err
	}

	return profile, cli.GetObjectPrinter(d.Config().GetString("output")), nil
}
