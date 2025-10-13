package queryflow

import (
	"bytes"
	"flag"
	"strings"
	"testing"

	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/pureinsights/pdp-cli/internal/iostreams"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

var Update = flag.Bool("update", false, "rewrite golden files")

// Test_NewQueryFlowCommand tests the NewQueryFlowCommand() function
func Test_NewQueryFlowCommand(t *testing.T) {
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
	queryflowCmd := NewQueryFlowCommand(d)

	queryflowCmd.SetIn(ios.In)
	queryflowCmd.SetOut(ios.Out)
	queryflowCmd.SetErr(ios.Err)

	queryflowCmd.PersistentFlags().StringP(
		"profile",
		"p",
		d.Config().GetString("profile"),
		"configuration profile to use",
	)

	var commandNames []string
	for _, c := range queryflowCmd.Commands() {
		commandNames = append(commandNames, c.Name())
	}

	expectedCommands := []string{"config"}
	for _, c := range expectedCommands {
		require.Contains(t, commandNames, c)
	}
}
