package core

import (
	"bytes"
	"flag"
	"slices"
	"strings"
	"testing"

	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/pureinsights/pdp-cli/internal/iostreams"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

var Update = flag.Bool("update", false, "rewrite golden files")

// Test_NewCoreCommand tests the NewCoreCommand() function
func Test_NewCoreCommand(t *testing.T) {
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
	coreCmd := NewCoreCommand(d)

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
		if !slices.Contains([]string{"help", "completion"}, c.Name()) {
			commandNames = append(commandNames, c.Name())
		}
	}

	expectedCommands := []string{"config", "label", "secret", "credential", "server"}
	assert.Equal(t, expectedCommands, commandNames)
}
