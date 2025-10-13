package config

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/pureinsights/pdp-cli/internal/iostreams"
	"github.com/pureinsights/pdp-cli/internal/testutils"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test_NewConfigCommand tests the NewConfigCommand() function.
func Test_NewConfigCommand_ProfileFlag(t *testing.T) {
	tests := []struct {
		name      string
		config    map[string]string
		writePath string
		outGolden string
		errGolden string
		err       error
	}{
		// Working cases
		{
			name:      "Every value exists",
			writePath: t.TempDir(),
			config: map[string]string{
				"profile":          "cn",
				"cn.core_url":      "http://localhost:12010",
				"cn.ingestion_url": "http://localhost:12030",
				"cn.queryflow_url": "http://localhost:12040",
				"cn.staging_url":   "http://localhost:12020",
				"cn.core_key":      "core321",
				"cn.ingestion_key": "ingestion432",
				"cn.queryflow_key": "queryflow123",
				"cn.staging_key":   "staging235",
			},
			outGolden: "NewConfigCommand_Out_All",
			errGolden: "NewConfigCommand_Err_All",
			err:       nil,
		},
		{
			name:      "No keys exist",
			writePath: t.TempDir(),
			config: map[string]string{
				"profile":          "cn",
				"cn.core_url":      "http://localhost:12010",
				"cn.ingestion_url": "http://localhost:12030",
				"cn.queryflow_url": "http://localhost:12040",
				"cn.staging_url":   "http://localhost:12020",
			},
			outGolden: "NewConfigCommand_Out_NoKeys",
			errGolden: "NewConfigCommand_Err_NoKeys",
			err:       nil,
		},
		{
			name:      "Only keys exist",
			writePath: t.TempDir(),
			config: map[string]string{
				"cn.core_key":      "core321",
				"cn.ingestion_key": "ingestion432",
				"cn.queryflow_key": "queryflow123",
				"cn.staging_key":   "staging235",
			},
			outGolden: "NewConfigCommand_Out_OnlyKeys",
			errGolden: "NewConfigCommand_Err_OnlyKeys",
			err:       nil,
		},
		{
			name:      "There are keys with multiple periods in their viper keys",
			writePath: t.TempDir(),
			config: map[string]string{
				"profile":                "cn",
				"cn.core_url":            "http://localhost:12010",
				"cn.ingestion_url":       "http://localhost:12030",
				"cn.queryflow_url":       "http://localhost:12040",
				"cn.staging_url":         "http://localhost:12020",
				"cn.core_key":            "core321",
				"cn.cn.ingestion_key":    "ingestion432",
				"cn.cn.cn.queryflow_key": "queryflow123",
				"cn.cn.cn.staging_key":   "staging235",
			},
			outGolden: "NewConfigCommand_Out_MultiplePeriods",
			errGolden: "NewConfigCommand_Err_MultiplePeriods",
			err:       nil,
		},

		// Error cases
		{
			name:      "Writing to config.toml fails",
			writePath: "doesnotexist",
			config: map[string]string{
				"profile":          "cn",
				"cn.core_url":      "http://localhost:12010",
				"cn.ingestion_url": "http://localhost:12030",
				"cn.queryflow_url": "http://localhost:12040",
				"cn.staging_url":   "http://localhost:12020",
				"cn.core_key":      "core321",
				"cn.ingestion_key": "ingestion432",
				"cn.queryflow_key": "queryflow123",
				"cn.staging_key":   "staging235",
			},
			outGolden: "NewConfigCommand_Out_ConfigError",
			errGolden: "NewConfigCommand_Err_ConfigError",
			err:       cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("open doesnotexist\\config.toml: The system cannot find the path specified."), "Failed to save Core's configuration"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			in := strings.NewReader(strings.Repeat("\n", 8))
			out := &bytes.Buffer{}
			errBuf := &bytes.Buffer{}
			ios := iostreams.IOStreams{
				In:  in,
				Out: out,
				Err: errBuf,
			}

			vpr := viper.New()
			for k, v := range tc.config {
				vpr.Set(k, v)
			}

			d := cli.NewDiscovery(&ios, vpr, tc.writePath)

			configCmd := NewConfigCommand(d)

			configCmd.SetIn(ios.In)
			configCmd.SetOut(ios.Out)
			configCmd.SetErr(ios.Err)

			configCmd.PersistentFlags().StringP(
				"profile",
				"p",
				d.Config().GetString("profile"),
				"configuration profile to use",
			)

			configCmd.SetArgs([]string{"--profile=cn"})

			err := configCmd.Execute()
			if tc.err != nil {
				var errStruct cli.Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
			}

			testutils.CompareBytes(t, tc.outGolden, out.Bytes())
			testutils.CompareBytes(t, tc.errGolden, errBuf.Bytes())

			var commandNames []string
			for _, c := range configCmd.Commands() {
				commandNames = append(commandNames, c.Name())
			}

			expectedCommands := []string{"get"}
			for _, c := range expectedCommands {
				require.Contains(t, commandNames, c)
			}
		})
	}
}

// Test_NewConfigCommand_NoProfileFlag tests the config command when there is no profile flag defined.
func Test_NewConfigCommand_NoProfileFlag(t *testing.T) {
	in := strings.NewReader(strings.Repeat("\n", 8))
	out := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	ios := iostreams.IOStreams{
		In:  in,
		Out: out,
		Err: errBuf,
	}

	config := map[string]string{
		"profile":          "cn",
		"cn.core_url":      "http://localhost:12010",
		"cn.ingestion_url": "http://localhost:12030",
		"cn.queryflow_url": "http://localhost:12040",
		"cn.staging_url":   "http://localhost:12020",
		"cn.core_key":      "core321",
		"cn.ingestion_key": "ingestion432",
		"cn.queryflow_key": "queryflow123",
		"cn.staging_key":   "staging235",
	}

	vpr := viper.New()
	for k, v := range config {
		vpr.Set(k, v)
	}

	d := cli.NewDiscovery(&ios, vpr, t.TempDir())

	configCmd := NewConfigCommand(d)

	configCmd.SetIn(ios.In)
	configCmd.SetOut(ios.Out)
	configCmd.SetErr(ios.Err)

	configCmd.SetArgs([]string{})

	err := configCmd.Execute()
	require.Error(t, err)
	assert.Equal(t, "flag accessed but not defined: profile", err.Error())

	testutils.CompareBytes(t, "NewConfigCommand_Out_NoProfile", out.Bytes())
	testutils.CompareBytes(t, "NewConfigCommand_Err_NoProfile", errBuf.Bytes())
}
