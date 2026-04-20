package version

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/pureinsights/discovery-cli/internal/iostreams"
	"github.com/pureinsights/discovery-cli/internal/testutils"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewVersionCommand_versionPrints tests the NewVersionCommand() function when the version can be printed.
func TestNewVersionCommand_versionPrints(t *testing.T) {
	in := strings.NewReader("In Reader")
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	ios := iostreams.IOStreams{
		In:  in,
		Out: out,
		Err: errBuf,
	}

	dir := t.TempDir()
	vpr := viper.New()
	vpr.SetDefault("profile", "default")
	Version = "2.7.1"
	d := cli.NewDiscovery(&ios, vpr, dir)
	versionCmd := NewVersionCommand(d)
	versionCmd.SilenceUsage = true
	versionCmd.SetIn(ios.In)
	versionCmd.SetOut(ios.Out)
	versionCmd.SetErr(ios.Err)
	versionCmd.SetArgs([]string{})
	err := versionCmd.Execute()
	assert.Nil(t, err)
	testutils.CompareBytes(t, "NewVersionCommand_Out_VersionPrints", testutils.Read(t, "NewVersionCommand_Out_VersionPrints"), out.Bytes())
}

// TestNewVersionCommand_versionFails tests the NewVersionCommand() function when the version cannot be printed.
func TestNewVersionCommand_versionFails(t *testing.T) {
	in := strings.NewReader("In Reader")
	out := testutils.ErrWriter{Err: errors.New("write failed")}
	errBuf := &bytes.Buffer{}
	ios := iostreams.IOStreams{
		In:  in,
		Out: out,
		Err: errBuf,
	}

	dir := t.TempDir()
	vpr := viper.New()
	vpr.SetDefault("profile", "default")
	Version = "2.7.1"
	d := cli.NewDiscovery(&ios, vpr, dir)
	versionCmd := NewVersionCommand(d)
	versionCmd.SilenceUsage = true
	versionCmd.SetIn(ios.In)
	versionCmd.SetOut(ios.Out)
	versionCmd.SetErr(ios.Err)

	versionCmd.SetArgs([]string{})
	err := versionCmd.Execute()
	var errStruct cli.Error
	require.ErrorAs(t, err, &errStruct)
	assert.EqualError(t, err, cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("write failed"), "Could not print CLI version").Error())
	testutils.CompareBytes(t, "NewVersionCommand_Err_VersionFails", testutils.Read(t, "NewVersionCommand_Err_VersionFails"), errBuf.Bytes())
}
