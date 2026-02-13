package cli

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"os"
	"testing"

	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/iostreams"
	"github.com/pureinsights/discovery-cli/internal/testutils"
	"github.com/pureinsights/discovery-cli/internal/testutils/mocks"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// Test_discovery_PingServer tests the discovery.PingServer() function.
func Test_discovery_PingServer(t *testing.T) {
	tests := []struct {
		name           string
		client         ServerPinger
		printer        Printer
		expectedOutput string
		outWriter      io.Writer
		err            error
	}{
		// Working case
		{
			name:           "StartSeed correctly prints the received object with the pretty printer",
			client:         new(mocks.WorkingServerPinger),
			printer:        nil,
			expectedOutput: "{\n  \"acknowledged\": true\n}\n",
			err:            nil,
		},
		{
			name:           "StartSeed correctly prints the received object with JSON ugly printer",
			client:         new(mocks.WorkingServerPinger),
			printer:        JsonObjectPrinter(false),
			expectedOutput: "{\"acknowledged\":true}\n",
			err:            nil,
		},

		// Error case
		{
			name:           "GetEntityById fails",
			client:         new(mocks.FailingServerPingerServerNotFound),
			printer:        nil,
			expectedOutput: "",
			err: NewErrorWithCause(ErrorExitCode, discoveryPackage.Error{
				Status: http.StatusNotFound,
				Body: gjson.Parse(`{
			"status": 404,
			"code": 1003,
			"messages": [
				"Server not found: 21029da3-041c-43b5-a67e-870251f2f6a6"
			],
			"timestamp": "2025-10-16T00:15:31.888410500Z"
		}`),
			}, "Could not get server ID."),
		},
		{
			name:           "Ping fails",
			client:         new(mocks.FailingServerPingerPingFailed),
			printer:        nil,
			expectedOutput: "",
			err: NewErrorWithCause(ErrorExitCode, discoveryPackage.Error{Status: http.StatusUnprocessableEntity, Body: gjson.Parse(`{
	"status": 422,
	"code": 4001,
	"messages": [
		"Client of type openai cannot be validated."
	],
	"timestamp": "2025-10-23T22:35:38.345647200Z"
	}`)}, "Could not ping server with id \"21029da3-041c-43b5-a67e-870251f2f6a6\""),
		},
		{
			name:      "Printing fails",
			client:    new(mocks.WorkingServerPinger),
			printer:   nil,
			outWriter: testutils.ErrWriter{Err: errors.New("write failed")},
			err:       NewErrorWithCause(ErrorExitCode, errors.New("write failed"), "Could not print JSON object"),
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

			d := NewDiscovery(&ios, viper.New(), "")
			err := d.PingServer(tc.client, "Mongo Atlas server", tc.printer)

			if tc.err != nil {
				require.Error(t, err)
				var errStruct Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedOutput, buf.String())
			}
		})
	}
}
