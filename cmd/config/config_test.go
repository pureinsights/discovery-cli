package config

import (
	"bytes"
	"strings"
	"testing"

	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/pureinsights/pdp-cli/internal/iostreams"
	"github.com/pureinsights/pdp-cli/internal/testutils"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

// Test_NewConfigCommand tests the NewConfigCommand() function.
func Test_NewConfigCommand(t *testing.T) {
	in := strings.NewReader(strings.Repeat("\n", 8))
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	ios := iostreams.IOStreams{
		In:  in,
		Out: out,
		Err: errBuf,
	}

	dir := t.TempDir()

	config := map[string]string{
		"profile":          "cn",
		"cn.core_url":      "http://localhost:8080",
		"cn.ingestion_url": "http://localhost:8080",
		"cn.queryflow_url": "http://localhost:8088",
		"cn.staging_url":   "http://localhost:8081",
		"cn.core_key":      "core321",
		"cn.ingestion_key": "ingestion432",
		"cn.queryflow_key": "queryflow123",
		"cn.staging_key":   "staging235",
	}

	vpr := viper.New()
	for k, v := range config {
		vpr.Set(k, v)
	}

	d := cli.NewDiscovery(&ios, vpr, dir)

	configCmd := NewConfigCommand(d)

	configCmd.PersistentFlags().StringP(
		"profile",
		"p",
		d.Config().GetString("profile"),
		"configuration profile to use",
	)

	configCmd.SetArgs([]string{"discovery config", "--profile=cn"})

	err := configCmd.Execute()
	require.NoError(t, err)

	testutils.CompareBytes(t, "newConfigCommand_Out_All", out.Bytes())
	testutils.CompareBytes(t, "newConfigCommand_Err_All", errBuf.Bytes())
}
