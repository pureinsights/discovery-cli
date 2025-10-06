package config

import (
	"bytes"
	"errors"
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

func Test_NewGetCommand(t *testing.T) {
	tests := []struct {
		name      string
		profile   string
		sensitive bool
		config    map[string]string
		outGolden string
		errGolden string
		outWriter io.Writer
		err       error
	}{
		// {
		// 	name:      "Print all values not sensitive",
		// 	profile:   "cn",
		// 	sensitive: false,
		// 	config: map[string]string{
		// 		"profile":          "cn",
		// 		"cn.core_url":      "http://localhost:12010",
		// 		"cn.core_key":      "discovery.key.core.cn",
		// 		"cn.ingestion_url": "http://localhost:12020",
		// 		"cn.ingestion_key": "discovery.key.ingestion.cn",
		// 		"cn.queryflow_url": "http://localhost:12030",
		// 		"cn.queryflow_key": "discovery.key.queryflow.cn",
		// 		"cn.staging_url":   "http://localhost:12040",
		// 		"cn.staging_key":   "discovery.key.staging.cn",
		// 	},
		// 	outGolden: "NewGetCommand_Out_AllNotSensitive",
		// 	errGolden: "NewGetCommand_Err_AllNotSensitive",
		// 	outWriter: nil,
		// 	err:       nil,
		// },
		// {
		// 	name:      "Print all values sensitive",
		// 	profile:   "cn",
		// 	sensitive: true,
		// 	config: map[string]string{
		// 		"profile":          "cn",
		// 		"cn.core_url":      "http://localhost:12010",
		// 		"cn.core_key":      "discovery.key.core.cn",
		// 		"cn.ingestion_url": "http://localhost:12020",
		// 		"cn.ingestion_key": "discovery.key.ingestion.cn",
		// 		"cn.queryflow_url": "http://localhost:12030",
		// 		"cn.queryflow_key": "discovery.key.queryflow.cn",
		// 		"cn.staging_url":   "http://localhost:12040",
		// 		"cn.staging_key":   "discovery.key.staging.cn",
		// 	},
		// 	outGolden: "NewGetCommand_Out_AllSensitive",
		// 	errGolden: "NewGetCommand_Err_AllSensitive",
		// 	outWriter: nil,
		// 	err:       nil,
		// },
		// {
		// 	name:      "Print some values",
		// 	profile:   "cn",
		// 	sensitive: false,
		// 	config: map[string]string{
		// 		"profile":          "cn",
		// 		"cn.core_url":      "http://localhost:12010",
		// 		"cn.core_key":      "discovery.key.core.cn",
		// 		"cn.ingestion_url": "http://localhost:12020",
		// 		"cn.queryflow_key": "discovery.key.queryflow.cn",
		// 	},
		// 	outWriter: nil,
		// 	err:       nil,
		// },
		{
			name:      "Print Fail on Printing Instructions",
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
			outGolden: "NewGetCommand_Out_FailPrintingInstructions",
			errGolden: "NewGetCommand_Err_FailPrintingInstructions",
			outWriter: testutils.ErrWriter{Err: errors.New("write failed")},
			err:       errors.New("write failed"),
		},
		{
			name:      "Print Fail on Printing Core Config",
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
			outGolden: "NewGetCommand_Out_FailPrintingCore",
			errGolden: "NewGetCommand_Err_FailPrintingCore",
			outWriter: &testutils.FailOnNWriter{Writer: &bytes.Buffer{}, N: 2},
			err:       errors.New("write failed"),
		},
		{
			name:      "Print Fail on Printing Ingestion Config",
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
			outGolden: "NewGetCommand_Out_FailPrintingIngestion",
			errGolden: "NewGetCommand_Err_FailPrintingIngestion",
			outWriter: &testutils.FailOnNWriter{Writer: &bytes.Buffer{}, N: 4},
			err:       errors.New("write failed"),
		},
		{
			name:      "Print Fail on Printing QueryFlow Config",
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
			outGolden: "NewGetCommand_Out_FailPrintingQueryFlow",
			errGolden: "NewGetCommand_Err_FailPrintingQueryFlow",
			outWriter: &testutils.FailOnNWriter{Writer: &bytes.Buffer{}, N: 6},
			err:       errors.New("write failed"),
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
			outWriter: &testutils.FailOnNWriter{Writer: &bytes.Buffer{}, N: 8},
			err:       errors.New("write failed"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			in := strings.NewReader(strings.Repeat("\n", 8))
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

			getCmd.SetArgs([]string{"discovery config get", "--profile=cn"})

			err := getCmd.Execute()
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
