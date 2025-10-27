package cmd

import (
	"bytes"
	"errors"
	"flag"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/pureinsights/pdp-cli/internal/iostreams"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Defines the update flag in the package
var Update = flag.Bool("update", false, "rewrite golden files")

// Test_newRootCommand tests the newRootCommand() function.
func Test_newRootCommand(t *testing.T) {
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
	d := cli.NewDiscovery(&ios, vpr, dir)
	discoveryCmd := newRootCommand(d)

	assert.Equal(t, in, discoveryCmd.InOrStdin())
	assert.Equal(t, out, discoveryCmd.OutOrStdout())
	assert.Equal(t, errBuf, discoveryCmd.ErrOrStderr())

	assert.Equal(t, "default", discoveryCmd.PersistentFlags().Lookup("profile").DefValue)

	// Change flag value to check Viper binding
	discoveryCmd.PersistentFlags().Set("profile", "cn")
	assert.Equal(t, "cn", vpr.GetString("profile"))

	var commandNames []string
	for _, c := range discoveryCmd.Commands() {
		commandNames = append(commandNames, c.Name())
	}

	expectedCommands := []string{"config", "core"}
	for _, c := range expectedCommands {
		require.Contains(t, commandNames, c)
	}
}

// TestRun_SetDiscoveryDirFails tests the Run function when the SetDiscoveryDir() function fails.
func TestRun_SetDiscoveryDirFails(t *testing.T) {
	tmp := t.TempDir()

	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)

	target := filepath.Join(tmp, ".discovery")

	require.NoError(t, os.WriteFile(target, []byte("MkDirAll will fail"), 0o644))
	os.Args = []string{"discovery"}
	exitCode, err := Run()
	require.Error(t, err)
	assert.Equal(t, cli.ErrorExitCode, exitCode)
	var cliError *cli.Error
	require.ErrorAs(t, err, &cliError)
	cause := cliError.Cause
	isFileError := errors.Is(cause, fs.ErrNotExist) || errors.Is(cause, syscall.ENOTDIR)
	assert.True(t, isFileError)
	assert.Equal(t, "Could not create the /.discovery directory", cliError.Message)
	assert.Equal(t, cli.ErrorExitCode, cliError.ExitCode)
}

// TestRun_InitializeConfigFails tests the Run function when the InitializeConfig() function fails.
func TestRun_InitializeConfigFails(t *testing.T) {
	tmp := t.TempDir()

	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)

	require.NoError(t, os.Mkdir(filepath.Join(tmp, ".discovery"), 0o744))

	config := filepath.Join(filepath.Join(tmp, ".discovery"), "config.toml")

	require.NoError(t, os.WriteFile(config, []byte(`
{
  "default": {
    "core_url": "http://localhost:12010"
  },
  "cn": {
    "core_url": "http://discovery.core.cn"
  }
}
`), 0o644))
	os.Args = []string{"discovery"}
	exitCode, err := Run()
	require.Error(t, err)
	assert.Equal(t, cli.ErrorExitCode, exitCode)
	var cliError *cli.Error
	require.ErrorAs(t, err, &cliError)
	errorStruct := cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("While parsing config: toml: invalid character at start of key: {"), "Could not read the configuration file")
	assert.EqualError(t, cliError, errorStruct.Error())
}

// TestRun_ExecuteFails tests when the execution of the CLI results in an error
func TestRun_ExecuteFails(t *testing.T) {
	tmp := t.TempDir()

	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)

	os.Args = []string{"discovery", "--profiles=cn"}
	exitCode, err := Run()
	require.Error(t, err)
	assert.Equal(t, cli.ErrorExitCode, exitCode)
	var cliError *cli.Error
	require.ErrorAs(t, err, &cliError)
	assert.Equal(t, cliError.Message, "")
	assert.EqualError(t, cliError.Cause, "unknown flag: --profiles")
}

// TestRun_Success tests when the Run function works
func TestRun_Success(t *testing.T) {
	tmp := t.TempDir()

	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)

	os.Args = []string{"discovery"}
	exitCode, err := Run()
	require.NoError(t, err)
	assert.Equal(t, cli.SuccessExitCode, exitCode)
}
