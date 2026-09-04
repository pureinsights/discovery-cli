package version

import (
	"bytes"
	"errors"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/pureinsights/discovery-cli/v2/internal/cli"
	"github.com/pureinsights/discovery-cli/v2/internal/iostreams"
	"github.com/pureinsights/discovery-cli/v2/internal/testutils"
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

// Test_computeVersion tests the computeVersion() function.
func Test_computeVersion(t *testing.T) {
	tests := []struct {
		name     string
		current  string
		info     *debug.BuildInfo
		ok       bool
		expected string
	}{
		{
			name:     "There already is a version",
			current:  "v2.12.0",
			info:     &debug.BuildInfo{Main: debug.Module{Version: "v2.11.0"}},
			ok:       true,
			expected: "v2.12.0",
		},
		{
			name:     "The version is empty, so it is obtained from debug.BuildInfo",
			current:  "",
			info:     &debug.BuildInfo{Main: debug.Module{Version: "v2.12.0"}},
			ok:       true,
			expected: "v2.12.0",
		},
		{
			name:     "The version is dev, so it is obtained from debug.BuildInfo",
			current:  "dev",
			info:     &debug.BuildInfo{Main: debug.Module{Version: "v2.12.0"}},
			ok:       true,
			expected: "v2.12.0",
		},
		{
			name:     "debug.ReadBuildInfo is not ok, so the version is unchanged",
			current:  "dev",
			info:     nil,
			ok:       false,
			expected: "dev",
		},
		{
			name:     "debug.ReadBuildInfo is (devel), so the version is unchanged",
			current:  "",
			info:     &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}},
			ok:       true,
			expected: "",
		},
		{
			name:     "debug.ReadBuildInfo is an empty string, so the version is unchanged",
			current:  "dev",
			info:     &debug.BuildInfo{Main: debug.Module{Version: ""}},
			ok:       true,
			expected: "dev",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := computeVersion(tc.current, tc.info, tc.ok)
			assert.Equal(t, tc.expected, got)
		})
	}
}
