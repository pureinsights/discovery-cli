package commands

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/pureinsights/discovery-cli/internal/cli"
	"github.com/pureinsights/discovery-cli/internal/iostreams"
	"github.com/pureinsights/discovery-cli/internal/testutils"
	"github.com/pureinsights/discovery-cli/internal/testutils/mocks"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStatusCheckCommand tests the StatusCheckCommand() function.
func TestStatusCheckCommand(t *testing.T) {
	tests := []struct {
		name           string
		client         cli.StatusChecker
		product        string
		url            string
		apiKey         string
		expectedOutput string
		outWriter      io.Writer
		err            error
	}{
		// Working case
		{
			name:           "StatusCheck correctly prints the status with the pretty printer",
			client:         new(mocks.WorkingStatusChecker),
			url:            "http://localhost:12010",
			apiKey:         "core123",
			product:        "Core",
			expectedOutput: "{\n  \"status\": \"UP\"\n}\n",
			err:            nil,
		},

		// Error case
		{
			name:      "CheckCredentials fails",
			client:    new(mocks.WorkingStatusChecker),
			url:       "",
			apiKey:    "core123",
			product:   "Core",
			outWriter: testutils.ErrWriter{Err: errors.New("write failed")},
			err:       cli.NewError(cli.ErrorExitCode, "The Discovery Core URL is missing for profile \"default\".\nTo set the URL for the Discovery Core API, run any of the following commands:\n      discovery config  --profile \"default\"\n      discovery core config --profile \"default\""),
		},
		{
			name:           "StatusCheck returns error",
			client:         new(mocks.FailingStatusChecker),
			expectedOutput: "",
			url:            "http://localhost:12010",
			apiKey:         "core123",
			product:        "Core",
			err:            cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("Get \"http://localhost:12030/health\": dial tcp [::1]:12030: connectex: No connection could be made because the target machine actively refused it."), "Could not check the status of Discovery Core."),
		},
		{
			name:      "Printing fails",
			client:    new(mocks.WorkingStatusChecker),
			url:       "http://localhost:12010",
			apiKey:    "core123",
			product:   "Core",
			outWriter: testutils.ErrWriter{Err: errors.New("write failed")},
			err:       cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("write failed"), "Could not print JSON object"),
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
			vpr.Set("profile", "default")
			vpr.Set("output", "pretty-json")
			if tc.url != "" {
				vpr.Set("default.core_url", tc.url)
			}
			if tc.apiKey != "" {
				vpr.Set("default.core_key", tc.apiKey)
			}

			d := cli.NewDiscovery(&ios, vpr, "")
			err := StatusCheckCommand(d, tc.client, tc.product, GetCommandConfig("default", "pretty-json", tc.product, "core_url"))

			if tc.err != nil {
				require.Error(t, err)
				var errStruct cli.Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedOutput, buf.String())
			}
		})
	}
}
