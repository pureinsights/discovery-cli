package seed_schedules

import (
	"bytes"
	"slices"
	"strings"
	"testing"

	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/pureinsights/discovery-cli/internal/iostreams"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// Test_NewSeedScheduleCommand tests the NewSeedScheduleCommand() function.
func Test_NewSeedScheduleCommand(t *testing.T) {
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
	seedScheduleCmd := NewSeedScheduleCommand(d)

	seedScheduleCmd.SetIn(ios.In)
	seedScheduleCmd.SetOut(ios.Out)
	seedScheduleCmd.SetErr(ios.Err)

	seedScheduleCmd.PersistentFlags().StringP(
		"profile",
		"p",
		d.Config().GetString("profile"),
		"configuration profile to use",
	)

	var commandNames []string
	for _, c := range seedScheduleCmd.Commands() {
		if !slices.Contains([]string{"help", "completion"}, c.Name()) {
			commandNames = append(commandNames, c.Name())
		}
	}

	expectedCommands := []string{"get"}
	assert.Equal(t, expectedCommands, commandNames)
}
