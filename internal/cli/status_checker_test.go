package cli

import (
	"bytes"
	"errors"
	"io"
	"net/url"
	"os"
	"testing"

	discoveryPackage "github.com/pureinsights/discovery-cli/v2/discovery"
	"github.com/pureinsights/discovery-cli/v2/internal/iostreams"
	"github.com/pureinsights/discovery-cli/v2/internal/testutils"
	"github.com/pureinsights/discovery-cli/v2/internal/testutils/mocks"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// errorStatusChecker is a StatusChecker that fails with the given error.
type errorStatusChecker struct {
	err error
}

// StatusCheck returns the configured error.
func (c errorStatusChecker) StatusCheck() (gjson.Result, error) {
	return gjson.Result{}, c.err
}

// Test_discovery_StatusCheck tests the discovery.StatusCheck() function.
func Test_discovery_StatusCheck(t *testing.T) {
	tests := []struct {
		name           string
		client         StatusChecker
		printer        Printer
		product        string
		expectedOutput string
		outWriter      io.Writer
		err            error
	}{
		// Working case
		{
			name:           "StatusCheck correctly prints the status with the pretty printer",
			client:         new(mocks.WorkingStatusChecker),
			printer:        nil,
			expectedOutput: "{\n  \"status\": \"UP\"\n}\n",
			err:            nil,
		},
		{
			name:           "StatusCheck correctly prints the status with JSON ugly printer",
			client:         new(mocks.WorkingStatusChecker),
			printer:        JsonObjectPrinter(false),
			expectedOutput: "{\"status\":\"UP\"}\n",
			err:            nil,
		},

		// Error case
		{
			name:           "StatusCheck prints DOWN and returns error when the product is unreachable",
			client:         new(mocks.FailingStatusChecker),
			printer:        nil,
			expectedOutput: "{\n  \"error\": \"Get \\\"http://localhost:12030/health\\\": dial tcp [::1]:12030: connectex: No connection could be made because the target machine actively refused it.\",\n  \"status\": \"DOWN\"\n}\n",
			product:        "Core",
			err:            NewError(ErrorExitCode, "Could not check the status of Discovery Core."),
		},
		{
			name:           "StatusCheck prints INVALID CONFIGURATION and returns error when the URL is empty",
			client:         errorStatusChecker{&url.Error{Op: "Get", URL: "/health", Err: errors.New("unsupported protocol scheme \"\"")}},
			printer:        JsonObjectPrinter(false),
			expectedOutput: "{\"error\":\"Get \\\"/health\\\": unsupported protocol scheme \\\"\\\"\",\"status\":\"INVALID CONFIGURATION\"}\n",
			product:        "Core",
			err:            NewError(ErrorExitCode, "Could not check the status of Discovery Core."),
		},
		{
			name:           "StatusCheck prints DOWN and returns error when the product responds with an error",
			client:         errorStatusChecker{discoveryPackage.Error{Status: 500, Body: gjson.Parse(`{"error":"Internal server error"}`)}},
			printer:        JsonObjectPrinter(false),
			expectedOutput: "{\"error\":\"status: 500, body: {\\\"error\\\":\\\"Internal server error\\\"}\\n\",\"status\":\"DOWN\"}\n",
			product:        "Core",
			err:            NewError(ErrorExitCode, "Could not check the status of Discovery Core."),
		},
		{
			name:           "StatusCheck prints INVALID CONFIGURATION and returns error when the API key is missing",
			client:         errorStatusChecker{discoveryPackage.Error{Status: 403, Body: gjson.Parse(`{"message":"Forbidden"}`)}},
			printer:        JsonObjectPrinter(false),
			expectedOutput: "{\"error\":\"status: 403, body: {\\\"message\\\":\\\"Forbidden\\\"}\\n\",\"status\":\"INVALID CONFIGURATION\"}\n",
			product:        "Core",
			err:            NewError(ErrorExitCode, "Could not check the status of Discovery Core."),
		},
		{
			name:           "StatusCheck prints INVALID CONFIGURATION and returns error when the URL has an unexpected path",
			client:         errorStatusChecker{discoveryPackage.Error{Status: 404, Body: gjson.Parse(`{"message":["Page Not Found"]}`)}},
			printer:        JsonObjectPrinter(false),
			expectedOutput: "{\"error\":\"status: 404, body: {\\\"message\\\":[\\\"Page Not Found\\\"]}\\n\",\"status\":\"INVALID CONFIGURATION\"}\n",
			product:        "Core",
			err:            NewError(ErrorExitCode, "Could not check the status of Discovery Core."),
		},
		{
			name:      "Printing fails",
			client:    new(mocks.WorkingStatusChecker),
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
			err := d.StatusCheck(tc.client, tc.product, tc.printer)

			if tc.err != nil {
				require.Error(t, err)
				var errStruct Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tc.expectedOutput, buf.String())
		})
	}
}

// Test_discovery_StatusCheckOfClients tests the discoveryStatusCheckOfClients() function.
func Test_discovery_StatusCheckOfClients(t *testing.T) {
	tests := []struct {
		name            string
		clients         []StatusCheckClientEntry
		printer         Printer
		expectedOutput  string
		outWriter       io.Writer
		err             error
		partialExitCode ExitCode
	}{
		// Working cases
		{
			name:            "StatusCheckOfClients correctly prints with the pretty printer when one of the status checks fails",
			clients:         []StatusCheckClientEntry{{Name: "core", Client: new(mocks.WorkingStatusChecker)}, {Name: "ingestion", Client: new(mocks.FailingStatusChecker)}, {Name: "queryflow", Client: new(mocks.WorkingStatusChecker)}, {Name: "staging", Client: new(mocks.WorkingStatusChecker)}},
			printer:         nil,
			expectedOutput:  "{\n  \"core\": {\n    \"status\": \"UP\"\n  },\n  \"ingestion\": {\n    \"error\": \"Get \\\"http://localhost:12030/health\\\": dial tcp [::1]:12030: connectex: No connection could be made because the target machine actively refused it.\",\n    \"status\": \"DOWN\"\n  },\n  \"queryflow\": {\n    \"status\": \"UP\"\n  },\n  \"staging\": {\n    \"status\": \"UP\"\n  }\n}\n",
			err:             nil,
			partialExitCode: ErrorExitCode,
		},
		{
			name:            "StatusCheckOfClients correctly prints the results with the ugly printer when one of the status checks fails",
			clients:         []StatusCheckClientEntry{{Name: "core", Client: new(mocks.WorkingStatusChecker)}, {Name: "ingestion", Client: new(mocks.WorkingStatusChecker)}, {Name: "queryflow", Client: new(mocks.FailingStatusChecker)}, {Name: "staging", Client: new(mocks.WorkingStatusChecker)}},
			printer:         JsonObjectPrinter(false),
			expectedOutput:  "{\"core\":{\"status\":\"UP\"},\"ingestion\":{\"status\":\"UP\"},\"queryflow\":{\"error\":\"Get \\\"http://localhost:12030/health\\\": dial tcp [::1]:12030: connectex: No connection could be made because the target machine actively refused it.\",\"status\":\"DOWN\"},\"staging\":{\"status\":\"UP\"}}\n",
			err:             nil,
			partialExitCode: ErrorExitCode,
		},
		{
			name: "StatusCheckOfClients prints INVALID CONFIGURATION or DOWN with the error depending on the failure",
			clients: []StatusCheckClientEntry{
				{Name: "core", Client: errorStatusChecker{&url.Error{Op: "Get", URL: "/health", Err: errors.New("unsupported protocol scheme \"\"")}}},
				{Name: "ingestion", Client: errorStatusChecker{&url.Error{Op: "Get", URL: "localhost:12030/health", Err: errors.New("unsupported protocol scheme \"localhost\"")}}},
				{Name: "queryflow", Client: errorStatusChecker{discoveryPackage.Error{Status: 503, Body: gjson.Parse(`{"status":"DOWN"}`)}}},
				{Name: "staging", Client: errorStatusChecker{discoveryPackage.Error{Status: 500}}},
			},
			printer:         JsonObjectPrinter(false),
			expectedOutput:  "{\"core\":{\"error\":\"Get \\\"/health\\\": unsupported protocol scheme \\\"\\\"\",\"status\":\"INVALID CONFIGURATION\"},\"ingestion\":{\"error\":\"Get \\\"localhost:12030/health\\\": unsupported protocol scheme \\\"localhost\\\"\",\"status\":\"INVALID CONFIGURATION\"},\"queryflow\":{\"error\":\"status: 503, body: {\\\"status\\\":\\\"DOWN\\\"}\\n\",\"status\":\"DOWN\"},\"staging\":{\"error\":\"status: 500, body: \\n\",\"status\":\"DOWN\"}}\n",
			err:             nil,
			partialExitCode: ErrorExitCode,
		},
		// Error cases
		{
			name:            "A working status has an invalid field name.",
			clients:         []StatusCheckClientEntry{{Name: "", Client: new(mocks.WorkingStatusChecker)}, {Name: "ingestion", Client: new(mocks.FailingStatusChecker)}, {Name: "queryflow", Client: new(mocks.WorkingStatusChecker)}, {Name: "staging", Client: new(mocks.WorkingStatusChecker)}},
			printer:         nil,
			expectedOutput:  "",
			err:             NewErrorWithCause(ErrorExitCode, errors.New("path cannot be empty"), "Could not get the status of the Discovery products"),
			partialExitCode: 0,
		},
		{
			name:            "A failing status has an invalid field name.",
			clients:         []StatusCheckClientEntry{{Name: "core", Client: new(mocks.WorkingStatusChecker)}, {Name: "", Client: new(mocks.FailingStatusChecker)}, {Name: "queryflow", Client: new(mocks.WorkingStatusChecker)}, {Name: "staging", Client: new(mocks.WorkingStatusChecker)}},
			printer:         nil,
			expectedOutput:  "",
			err:             NewErrorWithCause(ErrorExitCode, errors.New("path cannot be empty"), "Could not get the status of the Discovery products"),
			partialExitCode: ErrorExitCode,
		},
		{
			name:            "Printing fails",
			clients:         []StatusCheckClientEntry{{Name: "core", Client: new(mocks.WorkingStatusChecker)}, {Name: "ingestion", Client: new(mocks.FailingStatusChecker)}, {Name: "queryflow", Client: new(mocks.WorkingStatusChecker)}, {Name: "staging", Client: new(mocks.WorkingStatusChecker)}},
			printer:         nil,
			outWriter:       testutils.ErrWriter{Err: errors.New("write failed")},
			err:             NewErrorWithCause(ErrorExitCode, errors.New("write failed"), "Could not print JSON object"),
			partialExitCode: ErrorExitCode,
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
			err := d.StatusCheckOfClients(tc.clients, tc.printer)

			if tc.err != nil {
				var errStruct Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedOutput, buf.String())
			}
			assert.Equal(t, tc.partialExitCode, d.PartialExitCode())
		})
	}
}
