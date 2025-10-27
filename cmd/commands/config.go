package commands

import (
	"fmt"

	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/spf13/cobra"
)

const (
	LongConfig string = "config is the main command used to interact with Discovery %[1]s's configuration for a profile. This command by itself asks the user to save Discovery %[1]s's configuration for the given profile. The command prints the property to be modified along with its current value. If the property currently being shown is sensitive, its value is obfuscated. To keep the current value, the user must press \"Enter\" without any text, and to set the value as empty, a sole whitespace must be inputted."

	// SaveHeader contains the instructions header printed when saving a configuration.
	SaveHeader string = "Editing profile %q. Press Enter to keep the value shown, type a single space to set empty.\n\n"
	// PrintHeader contains the header displayed when printing a configuration.
	PrintHeader string = "Showing the configuration of profile %q:\n\n"
)

// SaveConfigCommand is the generic function to run the commands that save configurations.
func SaveConfigCommand(cmd *cobra.Command, config func(string) error) error {
	profile, err := cmd.Flags().GetString("profile")
	if err != nil {
		return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
	}
	return config(profile)
}

// SaveComponentConfigCommand is the generic function to run the commands that save configurations.
func SaveComponentConfigCommand(cmd *cobra.Command, config func(string, bool) error) error {
	profile, err := cmd.Flags().GetString("profile")
	if err != nil {
		return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
	}
	return config(profile, true)
}

// PrintConfigCommand is the generic function to run the commands that print configurations.
func PrintConfigCommand(cmd *cobra.Command, config func(string, bool) error) error {
	profile, err := cmd.Flags().GetString("profile")
	if err != nil {
		return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
	}
	sensitive, err := cmd.Flags().GetBool("sensitive")
	if err != nil {
		return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the sensitive flag")
	}

	ios := d.IOStreams()
	var err error
	if standalone {
		fmt.Fprintf(ios.Out, PrintHeader, profile)
	}
	return config(profile, sensitive)
}
