package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/pureinsights/pdp-cli/internal/cli"
	"github.com/pureinsights/pdp-cli/internal/iostreams"
	"github.com/pureinsights/pdp-cli/internal/testutils"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewGetCommand_WithProfileAndSensitiveFlags tests the NewGetCommand() function when there are profile and sensitive flags
func TestNewGetCommand_WithProfileAndSensitiveFlags(t *testing.T) {
	tests := []struct {
		name      string
		profile   string
		sensitive bool
		config    map[string]string
		outGolden string
		errGolden string
		outBytes  []byte
		errBytes  []byte
		outWriter io.Writer
		err       error
	}{
		{
			name:      "Print all values not sensitive",
			profile:   "cn",
			sensitive: false,
			config: map[string]string{
				"profile":          "cn",
				"cn.core_url":      "http://localhost:12010",
				"cn.core_key":      "discovery.key.core.cn",
				"cn.ingestion_url": "http://localhost:12020",
				"cn.ingestion_key": "discovery.key.ingestion.cn",
				"cn.queryflow_url": "http://localhost:12030",
				"cn.queryflow_key": "discovery.key.queryflow.cn",
				"cn.staging_url":   "http://localhost:12040",
				"cn.staging_key":   "discovery.key.staging.cn",
			},
			outGolden: "NewGetCommand_Out_AllNotSensitive",
			errGolden: "NewGetCommand_Err_AllNotSensitive",
			outBytes:  testutils.Read(t, "NewGetCommand_Out_AllNotSensitive"),
			errBytes:  []byte(nil),
			outWriter: nil,
			err:       nil,
		},
		{
			name:      "Print all values sensitive",
			profile:   "cn",
			sensitive: true,
			config: map[string]string{
				"profile":          "cn",
				"cn.core_url":      "http://localhost:12010",
				"cn.core_key":      "discovery.key.core.cn",
				"cn.ingestion_url": "http://localhost:12020",
				"cn.ingestion_key": "discovery.key.ingestion.cn",
				"cn.queryflow_url": "http://localhost:12030",
				"cn.queryflow_key": "discovery.key.queryflow.cn",
				"cn.staging_url":   "http://localhost:12040",
				"cn.staging_key":   "discovery.key.staging.cn",
			},
			outGolden: "NewGetCommand_Out_AllSensitive",
			errGolden: "NewGetCommand_Err_AllSensitive",
			outBytes:  testutils.Read(t, "NewGetCommand_Out_AllSensitive"),
			errBytes:  []byte(nil),
			outWriter: nil,
			err:       nil,
		},
		{
			name:      "Print some values",
			profile:   "default",
			sensitive: false,
			config: map[string]string{
				"profile":             "default",
				"default.staging_url": "http://localhost:12020",
			},
			outGolden: "NewGetCommand_Out_SomeValuesNotSensitive",
			errGolden: "NewGetCommand_Err_SomeValuesNotSensitive",
			outBytes:  testutils.Read(t, "NewGetCommand_Out_SomeValuesNotSensitive"),
			errBytes:  []byte(nil),
			outWriter: nil,
			err:       nil,
		},
		{
			name:      "Print Fail on Printing Staging Config",
			profile:   "cn",
			sensitive: false,
			config: map[string]string{
				"profile":          "cn",
				"cn.core_url":      "http://localhost:12010",
				"cn.core_key":      "discovery.key.core.cn",
				"cn.ingestion_url": "http://localhost:12020",
				"cn.ingestion_key": "discovery.key.ingestion.cn",
				"cn.queryflow_url": "http://localhost:12030",
				"cn.queryflow_key": "discovery.key.queryflow.cn",
				"cn.staging_url":   "http://localhost:12040",
				"cn.staging_key":   "discovery.key.staging.cn",
			},
			outGolden: "NewGetCommand_Out_FailPrintingStaging",
			errGolden: "NewGetCommand_Err_FailPrintingStaging",
			outBytes:  testutils.Read(t, "NewGetCommand_Out_FailPrintingStaging"),
			errBytes:  testutils.Read(t, "NewGetCommand_Err_FailPrintingStaging"),
			outWriter: &testutils.FailOnNWriter{Writer: &bytes.Buffer{}, N: 2},
			err:       cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("write failed"), "Could not print Staging's URL"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			in := strings.NewReader("")
			out := &bytes.Buffer{}
			var outWriter io.Writer
			if tc.outWriter != nil {
				failOnNWriter, ok := tc.outWriter.(*testutils.FailOnNWriter)
				if ok {
					failOnNWriter.Writer = out
					outWriter = failOnNWriter
				} else {
					outWriter = tc.outWriter
				}
			} else {
				outWriter = out
			}
			errBuf := &bytes.Buffer{}
			ios := iostreams.IOStreams{
				In:  in,
				Out: outWriter,
				Err: errBuf,
			}

			vpr := viper.New()
			for k, v := range tc.config {
				vpr.Set(k, v)
			}

			d := cli.NewDiscovery(&ios, vpr, t.TempDir())

			getCmd := NewGetCommand(d)

			getCmd.SetIn(ios.In)
			getCmd.SetOut(ios.Out)
			getCmd.SetErr(ios.Err)

			getCmd.PersistentFlags().StringP(
				"profile",
				"p",
				d.Config().GetString("profile"),
				"configuration profile to use",
			)

			getCmd.SetArgs([]string{fmt.Sprintf("--profile=%s", tc.profile), fmt.Sprintf("--sensitive=%t", tc.sensitive)})

			err := getCmd.Execute()
			if tc.err != nil {
				var errStruct cli.Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
				testutils.CompareBytes(t, tc.errGolden, tc.errBytes, errBuf.Bytes())
			} else {
				require.NoError(t, err)
			}

			testutils.CompareBytes(t, tc.outGolden, tc.outBytes, out.Bytes())
		})
	}
}

// TestNewGetCommand_NoProfileFlag test the NewGetCommand() function when there is no profile flag
func TestNewGetCommand_NoProfileFlag(t *testing.T) {
	in := strings.NewReader("")
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

	getCmd := NewGetCommand(d)

	getCmd.SetIn(ios.In)
	getCmd.SetOut(ios.Out)
	getCmd.SetErr(ios.Err)

	getCmd.SetArgs([]string{})

	err := getCmd.Execute()
	require.Error(t, err)
	var errStruct cli.Error
	require.ErrorAs(t, err, &errStruct)
	assert.EqualError(t, errStruct, cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("flag accessed but not defined: profile"), "Could not get the profile").Error())

	testutils.CompareBytes(t, "NewGetCommand_Out_NoProfile", testutils.Read(t, "NewGetCommand_Out_NoProfile"), out.Bytes())
	testutils.CompareBytes(t, "NewGetCommand_Err_NoProfile", testutils.Read(t, "NewGetCommand_Err_NoProfile"), errBuf.Bytes())
}
