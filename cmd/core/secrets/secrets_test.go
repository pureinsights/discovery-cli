package secrets

import (
	"bytes"
	"strings"
	"testing"

	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/pureinsights/pdp-cli/internal/iostreams"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

// Test_NewSecretCommand tests the NewCoreCommand() function
func Test_NewSecretCommand(t *testing.T) {
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
	coreCmd := NewSecretCommand(d)

	coreCmd.SetIn(ios.In)
	coreCmd.SetOut(ios.Out)
	coreCmd.SetErr(ios.Err)

	coreCmd.PersistentFlags().StringP(
		"profile",
		"p",
		d.Config().GetString("profile"),
		"configuration profile to use",
	)

	var commandNames []string
	for _, c := range coreCmd.Commands() {
		commandNames = append(commandNames, c.Name())
	}

	expectedCommands := []string{"get"}
	for _, c := range expectedCommands {
		require.Contains(t, commandNames, c)
	}
}
