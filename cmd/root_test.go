package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/pureinsights/pdp-cli/internal/iostreams"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

	discoveryCmd.PersistentFlags().Lookup("profile")
	assert.Equal(t, "default", discoveryCmd.PersistentFlags().Lookup("profile").DefValue)

	// Change flag value to check Viper binding
	discoveryCmd.PersistentFlags().Set("profile", "cn")
	assert.Equal(t, "cn", vpr.GetString("profile"))
}

func TestRun_SetDiscoveryDirFails(t *testing.T) {
	tmp := t.TempDir()

	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)

	target := filepath.Join(tmp, ".discovery")

	require.NoError(t, os.WriteFile(target, []byte("MkDirAll will fail"), 0o600))
	os.Args = []string{"discovery"}
	exitCode, err := Run()
	require.Error(t, err)
	assert.Equal(t, cli.ErrorExitCode, exitCode)
	cliError, ok := err.(cli.Error)
	if ok {
		assert.Equal(t, cliError.Message, "Could not set up Discovery's directory in User's home directory")
	}
}

func TestRun_InitializeConfigFails(t *testing.T) {
	tmp := t.TempDir()

	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)

	require.NoError(t, os.Mkdir(filepath.Join(tmp, ".discovery"), 0x700))

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
`), 0o600))
	os.Args = []string{"discovery"}
	exitCode, err := Run()
	require.Error(t, err)
	assert.Equal(t, cli.ErrorExitCode, exitCode)
	cliError, ok := err.(cli.Error)
	if ok {
		assert.Equal(t, cliError.Message, "Could not initialize configuration")
	}
}

func TestRun_Success(t *testing.T) {
	tmp := t.TempDir()

	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)

	os.Args = []string{"discovery"}
	exitCode, err := Run()
	require.NoError(t, err)
	assert.Equal(t, cli.SuccessExitCode, exitCode)
}
