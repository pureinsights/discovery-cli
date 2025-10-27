package commands

import (
	"bytes"
	"errors"
	"flag"
	"strings"
	"testing"

	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/pureinsights/pdp-cli/internal/iostreams"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Defines the update flag in the package
var Update = flag.Bool("update", false, "rewrite golden files")

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
			return SaveConfigCommand(cmd, d.SaveConfigFromUser)
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
			return SaveConfigCommand(cmd, d.SaveConfigFromUser)
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
