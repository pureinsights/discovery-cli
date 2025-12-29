package staging

import (
	"bytes"
	"flag"
	"slices"
	"strings"
	"testing"

	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/pureinsights/discovery-cli/internal/iostreams"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

var Update = flag.Bool("update", false, "rewrite golden files")

// TestNewStagingCommand tests the NewStagingCommand() function.
func TestNewStagingCommand(t *testing.T) {
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
	stagingCmd := NewStagingCommand(d)

	stagingCmd.SetIn(ios.In)
	stagingCmd.SetOut(ios.Out)
	stagingCmd.SetErr(ios.Err)

	stagingCmd.PersistentFlags().StringP(
		"profile",
		"p",
		d.Config().GetString("profile"),
		"configuration profile to use",
	)

	var commandNames []string
	for _, c := range stagingCmd.Commands() {
		if !slices.Contains([]string{"help", "completion"}, c.Name()) {
			commandNames = append(commandNames, c.Name())
		}
	}

	expectedCommands := []string{"bucket", "config", "status"}
	assert.Equal(t, expectedCommands, commandNames)
}
