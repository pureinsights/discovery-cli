package commands

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/pureinsights/pdp-cli/internal/iostreams"
	"github.com/pureinsights/pdp-cli/internal/testutils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSaveConfigCommand_WithProfile tests the SaveConfigCommand() function with the profile flag.
func TestSaveConfigCommand_WithProfile(t *testing.T) {
	in := strings.NewReader(strings.Repeat("\n", 8))
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	ios := iostreams.IOStreams{
		In:  in,
		Out: out,
		Err: errBuf,
	}

	d := cli.NewDiscovery(&ios, viper.New(), t.TempDir())

	configCmd := &cobra.Command{
		Use:   "config [subcommands]",
		Short: "Save Discovery's configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return SaveConfigCommand(cmd, &ios, d.SaveConfigFromUser)
		},
	}

	configCmd.PersistentFlags().StringP(
		"profile",
		"p",
		d.Config().GetString("profile"),
		"configuration profile to use",
	)

	configCmd.SetIn(ios.In)
	configCmd.SetOut(ios.Out)
	configCmd.SetErr(ios.Err)

	configCmd.SetArgs([]string{})

	err := configCmd.Execute()
	require.NoError(t, err)
}

// TestSaveConfigCommand_NoProfile tests the SaveConfigCommand() function with no profile flag.
func TestSaveConfigCommand_NoProfile(t *testing.T) {
	in := strings.NewReader(strings.Repeat("\n", 8))
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	ios := iostreams.IOStreams{
		In:  in,
		Out: out,
		Err: errBuf,
	}

	d := cli.NewDiscovery(&ios, viper.New(), t.TempDir())

	configCmd := &cobra.Command{
		Use:   "config [subcommands]",
		Short: "Save Discovery's configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return SaveConfigCommand(cmd, d.IOStreams(), d.SaveConfigFromUser)
		},
	}

	configCmd.SetIn(ios.In)
	configCmd.SetOut(ios.Out)
	configCmd.SetErr(ios.Err)

	configCmd.SetArgs([]string{})

	err := configCmd.Execute()
	require.Error(t, err)
	var errStruct cli.Error
	require.ErrorAs(t, err, &errStruct)
	assert.EqualError(t, errStruct, cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("flag accessed but not defined: profile"), "Could not get the profile").Error())
}

// TestSaveConfigCommand_ErrorPrintingHeader tests the SaveConfigCommand() function when printing the header fails.
func TestSaveConfigCommand_ErrorPrintingHeader(t *testing.T) {
	in := strings.NewReader(strings.Repeat("\n", 8))
	errBuf := &bytes.Buffer{}
	ios := iostreams.IOStreams{
		In:  in,
		Out: testutils.ErrWriter{Err: errors.New("write failed")},
		Err: errBuf,
	}

	d := cli.NewDiscovery(&ios, viper.New(), t.TempDir())

	configCmd := &cobra.Command{
		Use:   "config [subcommands]",
		Short: "Save Discovery's configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return SaveConfigCommand(cmd, &ios, d.SaveConfigFromUser)
		},
	}

	configCmd.PersistentFlags().StringP(
		"profile",
		"p",
		d.Config().GetString("profile"),
		"configuration profile to use",
	)

	configCmd.SetIn(ios.In)
	configCmd.SetOut(ios.Out)
	configCmd.SetErr(ios.Err)

	configCmd.SetArgs([]string{})

	err := configCmd.Execute()
	require.Error(t, err)
	var errStruct cli.Error
	require.ErrorAs(t, err, &errStruct)
	assert.EqualError(t, errStruct, cli.NewErrorWithCause(cli.ErrorExitCode, fmt.Errorf("write failed"), "Could not save the configuration").Error())
}

// TestPrintConfigCommand_WithProfileAndSensitive tests the PrintConfig() function with the profile and sensitive flags.
func TestPrintConfigCommand_WithProfileAndSensitive(t *testing.T) {
	in := strings.NewReader(strings.Repeat("\n", 8))
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	ios := iostreams.IOStreams{
		In:  in,
		Out: out,
		Err: errBuf,
	}

	d := cli.NewDiscovery(&ios, viper.New(), t.TempDir())

	configCmd := &cobra.Command{
		Use:   "config [subcommands]",
		Short: "Save Discovery's configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return PrintConfigCommand(cmd, d.PrintConfigToUser)
		},
	}

	configCmd.PersistentFlags().StringP(
		"profile",
		"p",
		d.Config().GetString("profile"),
		"configuration profile to use",
	)
	configCmd.Flags().BoolP("sensitive", "s", true, "--sensitive=true")

	configCmd.SetIn(ios.In)
	configCmd.SetOut(ios.Out)
	configCmd.SetErr(ios.Err)

	configCmd.SetArgs([]string{})

	err := configCmd.Execute()
	require.NoError(t, err)
}

// TestPrintConfigCommand_NoProfile tests the PrintConfig() function with no profile flag.
func TestPrintConfigCommand_NoProfile(t *testing.T) {
	in := strings.NewReader(strings.Repeat("\n", 8))
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	ios := iostreams.IOStreams{
		In:  in,
		Out: out,
		Err: errBuf,
	}

	d := cli.NewDiscovery(&ios, viper.New(), t.TempDir())

	configCmd := &cobra.Command{
		Use:   "config [subcommands]",
		Short: "Save Discovery's configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return PrintConfigCommand(cmd, d.PrintConfigToUser)
		},
	}

	configCmd.SetIn(ios.In)
	configCmd.SetOut(ios.Out)
	configCmd.SetErr(ios.Err)

	configCmd.SetArgs([]string{})

	err := configCmd.Execute()
	require.Error(t, err)
	var errStruct cli.Error
	require.ErrorAs(t, err, &errStruct)
	assert.EqualError(t, errStruct, cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("flag accessed but not defined: profile"), "Could not get the profile").Error())
}

// TestPrintConfigCommand_NoSensitive tests the PrintConfig() function with no sensitive flag.
func TestPrintConfigCommand_NoSensitive(t *testing.T) {
	in := strings.NewReader(strings.Repeat("\n", 8))
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	ios := iostreams.IOStreams{
		In:  in,
		Out: out,
		Err: errBuf,
	}

	d := cli.NewDiscovery(&ios, viper.New(), t.TempDir())

	configCmd := &cobra.Command{
		Use:   "config [subcommands]",
		Short: "Save Discovery's configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return PrintConfigCommand(cmd, d.PrintConfigToUser)
		},
	}

	configCmd.PersistentFlags().StringP(
		"profile",
		"p",
		d.Config().GetString("profile"),
		"configuration profile to use",
	)

	configCmd.SetIn(ios.In)
	configCmd.SetOut(ios.Out)
	configCmd.SetErr(ios.Err)

	configCmd.SetArgs([]string{})

	err := configCmd.Execute()
	require.Error(t, err)
	var errStruct cli.Error
	require.ErrorAs(t, err, &errStruct)
	assert.EqualError(t, errStruct, cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("flag accessed but not defined: sensitive"), "Could not get the sensitive flag").Error())
}
