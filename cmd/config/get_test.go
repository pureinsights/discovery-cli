package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
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
func Test_NewGetComman3d(t *testing.T) {
	tests := []struct {
		name                string
		config              map[string]string
		writePath           string
		outGolden           string
		errGolden           string
		expectedConfig      map[string]string
		expectedCredentials map[string]string
		err                 error
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
			outGolden: "newConfigCommand_Out_All",
			errGolden: "newConfigCommand_Err_All",
			expectedConfig: map[string]string{
				"profile":          "cn",
				"cn.core_url":      "http://localhost:12010",
				"cn.ingestion_url": "http://localhost:12030",
				"cn.queryflow_url": "http://localhost:12040",
				"cn.staging_url":   "http://localhost:12020",
			},
			expectedCredentials: map[string]string{
				"cn.core_key":      "core321",
				"cn.ingestion_key": "ingestion432",
				"cn.queryflow_key": "queryflow123",
				"cn.staging_key":   "staging235",
			},
			err: nil,
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
			outGolden: "newConfigCommand_Out_NoKeys",
			errGolden: "newConfigCommand_Err_NoKeys",
			expectedConfig: map[string]string{
				"profile":          "cn",
				"cn.core_url":      "http://localhost:12010",
				"cn.ingestion_url": "http://localhost:12030",
				"cn.queryflow_url": "http://localhost:12040",
				"cn.staging_url":   "http://localhost:12020",
			},
			expectedCredentials: map[string]string{},
			err:                 nil,
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
			outGolden:      "newConfigCommand_Out_OnlyKeys",
			errGolden:      "newConfigCommand_Err_OnlyKeys",
			expectedConfig: map[string]string{},
			expectedCredentials: map[string]string{
				"cn.core_key":      "core321",
				"cn.ingestion_key": "ingestion432",
				"cn.queryflow_key": "queryflow123",
				"cn.staging_key":   "staging235",
			},
			err: nil,
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
			outGolden: "newConfigCommand_Out_MultiplePeriods",
			errGolden: "newConfigCommand_Err_MultiplePeriods",
			expectedConfig: map[string]string{
				"profile":          "cn",
				"cn.core_url":      "http://localhost:12010",
				"cn.ingestion_url": "http://localhost:12030",
				"cn.queryflow_url": "http://localhost:12040",
				"cn.staging_url":   "http://localhost:12020",
			},
			expectedCredentials: map[string]string{
				"cn.core_key":            "core321",
				"cn.cn.ingestion_key":    "ingestion432",
				"cn.cn.cn.queryflow_key": "queryflow123",
				"cn.cn.cn.staging_key":   "staging235",
			},
			err: nil,
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
			outGolden: "newConfigCommand_Out_ConfigError",
			errGolden: "newConfigCommand_Err_ConfigError",
			expectedConfig: map[string]string{
				"profile":          "cn",
				"cn.core_url":      "http://localhost:12010",
				"cn.ingestion_url": "http://localhost:12030",
				"cn.queryflow_url": "http://localhost:12040",
				"cn.staging_url":   "http://localhost:12020",
			},
			expectedCredentials: map[string]string{
				"cn.core_key":      "core321",
				"cn.ingestion_key": "ingestion432",
				"cn.queryflow_key": "queryflow123",
				"cn.staging_key":   "staging235",
			},
			err: fmt.Errorf("cannot find the path specified"),
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

			configCmd.PersistentFlags().StringP(
				"profile",
				"p",
				d.Config().GetString("profile"),
				"configuration profile to use",
			)

			configCmd.SetArgs([]string{"discovery config", "--profile=cn"})

			err := configCmd.Execute()
			if tc.err != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.err.Error())
			} else {
				require.NoError(t, err)
			}

			testutils.CompareBytes(t, tc.outGolden, out.Bytes())
			testutils.CompareBytes(t, tc.errGolden, errBuf.Bytes())
		})
	}
}

func Test_NewGetCommand(t *testing.T) {
	tests := []struct {
		name           string
		profile        string
		sensitive      bool
		config         map[string]string
		outGolden      string
		errGolden      string
		expectedOutput string
		outWriter      io.Writer
		err            error
	}{
		{
			name:      "Print all values not sensitive",
			profile:   "cn",
			sensitive: false,
			config: map[string]string{
				"cn.core_url":      "http://localhost:12010",
				"cn.core_key":      "discovery.key.core.cn",
				"cn.ingestion_url": "http://localhost:12020",
				"cn.ingestion_key": "discovery.key.ingestion.cn",
				"cn.queryflow_url": "http://localhost:12030",
				"cn.queryflow_key": "discovery.key.queryflow.cn",
				"cn.staging_url":   "http://localhost:12040",
				"cn.staging_key":   "discovery.key.staging.cn",
			},
			expectedOutput: fmt.Sprintf("%s: %q\n%s: %q\n%s: %q\n%s: %q\n%s: %q\n%s: %q\n%s: %q\n%s: %q\n", "Core URL", "http://localhost:12010", "Core API Key", "discovery.key.core.cn", "Ingestion URL", "http://localhost:12020", "Ingestion API Key", "discovery.key.ingestion.cn", "QueryFlow URL", "http://localhost:12030", "QueryFlow API Key", "discovery.key.queryflow.cn", "Staging URL", "http://localhost:12040", "Staging API Key", "discovery.key.staging.cn"),
			outWriter:      nil,
			err:            nil,
		},
		{
			name:      "Print all values sensitive",
			profile:   "cn",
			sensitive: true,
			config: map[string]string{
				"cn.core_url":      "http://localhost:12010",
				"cn.core_key":      "discovery.key.core.cn",
				"cn.ingestion_url": "http://localhost:12020",
				"cn.ingestion_key": "discovery.key.ingestion.cn",
				"cn.queryflow_url": "http://localhost:12030",
				"cn.queryflow_key": "discovery.key.queryflow.cn",
				"cn.staging_url":   "http://localhost:12040",
				"cn.staging_key":   "discovery.key.staging.cn",
			},
			expectedOutput: fmt.Sprintf("%s: %q\n%s: %q\n%s: %q\n%s: %q\n%s: %q\n%s: %q\n%s: %q\n%s: %q\n", "Core URL", "http://localhost:12010", "Core API Key", obfuscate("discovery.key.core.cn"), "Ingestion URL", "http://localhost:12020", "Ingestion API Key", obfuscate("discovery.key.ingestion.cn"), "QueryFlow URL", "http://localhost:12030", "QueryFlow API Key", obfuscate("discovery.key.queryflow.cn"), "Staging URL", "http://localhost:12040", "Staging API Key", obfuscate("discovery.key.staging.cn")),
			outWriter:      nil,
			err:            nil,
		},
		{
			name:      "Print some values",
			profile:   "cn",
			sensitive: false,
			config: map[string]string{
				"cn.core_url":      "http://localhost:12010",
				"cn.core_key":      "discovery.key.core.cn",
				"cn.ingestion_url": "http://localhost:12020",
				"cn.queryflow_key": "discovery.key.queryflow.cn",
			},
			expectedOutput: fmt.Sprintf("%s: %q\n%s: %q\n%s: %q\n%s: %q\n", "Core URL", "http://localhost:12010", "Core API Key", "discovery.key.core.cn", "Ingestion URL", "http://localhost:12020", "QueryFlow API Key", "discovery.key.queryflow.cn"),
			outWriter:      nil,
			err:            nil,
		},
		{
			name:      "Print Fail on Printing Instructions",
			profile:   "cn",
			sensitive: false,
			config: map[string]string{
				"cn.core_url":      "http://localhost:12010",
				"cn.core_key":      "discovery.key.core.cn",
				"cn.ingestion_url": "http://localhost:12020",
				"cn.ingestion_key": "discovery.key.ingestion.cn",
				"cn.queryflow_url": "http://localhost:12030",
				"cn.queryflow_key": "discovery.key.queryflow.cn",
				"cn.staging_url":   "http://localhost:12040",
				"cn.staging_key":   "discovery.key.staging.cn",
			},
			expectedOutput: fmt.Sprintf("%s: %q\n%s: %q\n%s: %q\n%s: %q\n%s: %q\n%s: %q\n%s: %q\n%s: %q\n", "Core URL", "http://localhost:12010", "Core API Key", "discovery.key.core.cn", "Ingestion URL", "http://localhost:12020", "Ingestion API Key", "discovery.key.ingestion.cn", "QueryFlow URL", "http://localhost:12030", "QueryFlow API Key", "discovery.key.queryflow.cn", "Staging URL", "http://localhost:12040", "Staging API Key", "discovery.key.staging.cn"),
			outWriter:      errWriter{err: errors.New("write failed")},
			err:            errors.New("write failed"),
		},
		{
			name:      "Print Fail on Printing Core Config",
			profile:   "cn",
			sensitive: false,
			config: map[string]string{
				"cn.core_url":      "http://localhost:12010",
				"cn.core_key":      "discovery.key.core.cn",
				"cn.ingestion_url": "http://localhost:12020",
				"cn.ingestion_key": "discovery.key.ingestion.cn",
				"cn.queryflow_url": "http://localhost:12030",
				"cn.queryflow_key": "discovery.key.queryflow.cn",
				"cn.staging_url":   "http://localhost:12040",
				"cn.staging_key":   "discovery.key.staging.cn",
			},
			expectedOutput: fmt.Sprintf("%s: %q\n%s: %q\n%s: %q\n%s: %q\n%s: %q\n%s: %q\n%s: %q\n%s: %q\n", "Core URL", "http://localhost:12010", "Core API Key", "discovery.key.core.cn", "Ingestion URL", "http://localhost:12020", "Ingestion API Key", "discovery.key.ingestion.cn", "QueryFlow URL", "http://localhost:12030", "QueryFlow API Key", "discovery.key.queryflow.cn", "Staging URL", "http://localhost:12040", "Staging API Key", "discovery.key.staging.cn"),
			outWriter:      &failOnNWriter{Writer: &bytes.Buffer{}, N: 2},
			err:            errors.New("write failed"),
		},
		{
			name:      "Print Fail on Printing Ingestion Config",
			profile:   "cn",
			sensitive: false,
			config: map[string]string{
				"cn.core_url":      "http://localhost:12010",
				"cn.core_key":      "discovery.key.core.cn",
				"cn.ingestion_url": "http://localhost:12020",
				"cn.ingestion_key": "discovery.key.ingestion.cn",
				"cn.queryflow_url": "http://localhost:12030",
				"cn.queryflow_key": "discovery.key.queryflow.cn",
				"cn.staging_url":   "http://localhost:12040",
				"cn.staging_key":   "discovery.key.staging.cn",
			},
			expectedOutput: fmt.Sprintf("%s: %q\n%s: %q\n%s: %q\n%s: %q\n%s: %q\n%s: %q\n%s: %q\n%s: %q\n", "Core URL", "http://localhost:12010", "Core API Key", "discovery.key.core.cn", "Ingestion URL", "http://localhost:12020", "Ingestion API Key", "discovery.key.ingestion.cn", "QueryFlow URL", "http://localhost:12030", "QueryFlow API Key", "discovery.key.queryflow.cn", "Staging URL", "http://localhost:12040", "Staging API Key", "discovery.key.staging.cn"),
			outWriter:      &failOnNWriter{Writer: &bytes.Buffer{}, N: 4},
			err:            errors.New("write failed"),
		},
		{
			name:      "Print Fail on Printing QueryFlow Config",
			profile:   "cn",
			sensitive: false,
			config: map[string]string{
				"cn.core_url":      "http://localhost:12010",
				"cn.core_key":      "discovery.key.core.cn",
				"cn.ingestion_url": "http://localhost:12020",
				"cn.ingestion_key": "discovery.key.ingestion.cn",
				"cn.queryflow_url": "http://localhost:12030",
				"cn.queryflow_key": "discovery.key.queryflow.cn",
				"cn.staging_url":   "http://localhost:12040",
				"cn.staging_key":   "discovery.key.staging.cn",
			},
			expectedOutput: fmt.Sprintf("%s: %q\n%s: %q\n%s: %q\n%s: %q\n%s: %q\n%s: %q\n%s: %q\n%s: %q\n", "Core URL", "http://localhost:12010", "Core API Key", "discovery.key.core.cn", "Ingestion URL", "http://localhost:12020", "Ingestion API Key", "discovery.key.ingestion.cn", "QueryFlow URL", "http://localhost:12030", "QueryFlow API Key", "discovery.key.queryflow.cn", "Staging URL", "http://localhost:12040", "Staging API Key", "discovery.key.staging.cn"),
			outWriter:      &failOnNWriter{Writer: &bytes.Buffer{}, N: 6},
			err:            errors.New("write failed"),
		},
		{
			name:      "Print Fail on Printing Staging Config",
			profile:   "cn",
			sensitive: false,
			config: map[string]string{
				"cn.core_url":      "http://localhost:12010",
				"cn.core_key":      "discovery.key.core.cn",
				"cn.ingestion_url": "http://localhost:12020",
				"cn.ingestion_key": "discovery.key.ingestion.cn",
				"cn.queryflow_url": "http://localhost:12030",
				"cn.queryflow_key": "discovery.key.queryflow.cn",
				"cn.staging_url":   "http://localhost:12040",
				"cn.staging_key":   "discovery.key.staging.cn",
			},
			expectedOutput: fmt.Sprintf("%s: %q\n%s: %q\n%s: %q\n%s: %q\n%s: %q\n%s: %q\n%s: %q\n%s: %q\n", "Core URL", "http://localhost:12010", "Core API Key", "discovery.key.core.cn", "Ingestion URL", "http://localhost:12020", "Ingestion API Key", "discovery.key.ingestion.cn", "QueryFlow URL", "http://localhost:12030", "QueryFlow API Key", "discovery.key.queryflow.cn", "Staging URL", "http://localhost:12040", "Staging API Key", "discovery.key.staging.cn"),
			outWriter:      &failOnNWriter{Writer: &bytes.Buffer{}, N: 8},
			err:            errors.New("write failed"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			var out io.Writer
			if tc.outWriter != nil {
				out = tc.outWriter
			} else {
				out = buf
			}

			ios := iostreams.IOStreams{
				In:  os.Stdin,
				Out: out,
				Err: os.Stderr,
			}

			vpr := viper.New()
			for k, v := range tc.config {
				vpr.Set(k, v)
			}

			d := NewDiscovery(&ios, vpr, "")

			err := d.PrintConfigToUser(tc.profile, tc.sensitive)

			if tc.err != nil {
				require.Error(t, err)
				require.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				require.Contains(t, buf.String(), fmt.Sprintf("Showing the configuration of profile %q:\n\n", tc.profile))
				require.Contains(t, buf.String(), tc.expectedOutput)
			}
		})
	}
}
