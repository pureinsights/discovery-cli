package commands

import (
	"fmt"

	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/pureinsights/discovery-cli/internal/iostreams"
	"github.com/spf13/cobra"
)

const (
	// LongConfig is the text used in the Long field of a config command.
	LongConfig string = "config is the main command used to interact with Discovery %[1]s's configuration for a profile. This command by itself asks the user to save Discovery %[1]s's configuration for the given profile. The command prints the property to be modified along with its current value. If the property currently being shown is sensitive, its value is obfuscated. To keep the current value, the user must press \"Enter\" without any text, and to set the value as empty, a sole whitespace must be inputted."
	// LongConfigGet is the text used in the Long field of a config get command.
	LongConfigGet string = "get is the command used to obtain s %[1]s's configuration for a given profile. If the API keys are sensitive, the `sensitive` flag can be set to true in order to obfuscate them before printing them out. If a configuration property was not set, it is not displayed."
	// SaveHeader contains the instructions header printed when saving a configuration.
	SaveHeader string = "Editing profile %q. Press Enter to keep the value shown, type a single space to set empty.\n\n"
	// PrintHeader contains the header displayed when printing a configuration.
	PrintHeader string = "Showing the configuration of profile %q:\n\n"
)

// SaveConfigCommand is the generic function to run the commands that save configurations.
func SaveConfigCommand(cmd *cobra.Command, ios *iostreams.IOStreams, config func(string) error) error {
	profile, err := cmd.Flags().GetString("profile")
	if err != nil {
		return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
	}

	_, err = fmt.Fprintf(ios.Out, SaveHeader, profile)
	if err != nil {
		return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not save the configuration")
	}

	return config(profile)
}

// PrintConfigCommand is the generic function to run the commands that print configurations.
func PrintConfigCommand(cmd *cobra.Command, ios *iostreams.IOStreams, config func(string, bool) error) error {
	profile, err := cmd.Flags().GetString("profile")
	if err != nil {
		return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the profile")
	}
	sensitive, err := cmd.Flags().GetBool("sensitive")
	if err != nil {
		return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not get the sensitive flag")
	}

	_, err = fmt.Fprintf(ios.Out, PrintHeader, profile)
	if err != nil {
		return cli.NewErrorWithCause(cli.ErrorExitCode, err, "Could not print the configuration")
	}

	return config(profile, sensitive)
}
