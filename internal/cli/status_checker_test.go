package cli

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/pureinsights/discovery-cli/internal/iostreams"
	"github.com/pureinsights/discovery-cli/internal/testutils"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// WorkingStatusChecker mocks the results of a StatusChecker that does a request to an online product.
type WorkingStatusChecker struct {
	mock.Mock
}

// StatusCheck returns the response of an online Discovery product.
func (g *WorkingStatusChecker) StatusCheck() (gjson.Result, error) {
	return gjson.Parse(`{
    "status": "UP"
}`), nil
}

// WorkingStatusChecker mocks the results of a StatusChecker that does a request to an offline product.
type FailingStatusChecker struct {
	mock.Mock
}

// StatusCheck returns the error of an offline Discovery product.
func (g *FailingStatusChecker) StatusCheck() (gjson.Result, error) {
	return gjson.Result{}, errors.New("Get \"http://localhost:12030/health\": dial tcp [::1]:12030: connectex: No connection could be made because the target machine actively refused it.")
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
			client:         new(WorkingStatusChecker),
			printer:        nil,
			expectedOutput: "{\n  \"status\": \"UP\"\n}\n",
			err:            nil,
		},
		{
			name:           "StatusCheck correctly prints the status with JSON ugly printer",
			client:         new(WorkingStatusChecker),
			printer:        JsonObjectPrinter(false),
			expectedOutput: "{\"status\":\"UP\"}\n",
			err:            nil,
		},

		// Error case
		{
			name:           "StatusCheck returns error",
			client:         new(FailingStatusChecker),
			printer:        nil,
			expectedOutput: "",
			product:        "Core",
			err:            NewErrorWithCause(ErrorExitCode, errors.New("Get \"http://localhost:12030/health\": dial tcp [::1]:12030: connectex: No connection could be made because the target machine actively refused it."), "Could not check the status of Discovery Core."),
		},
		{
			name:      "Printing fails",
			client:    new(WorkingStatusChecker),
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
				require.Equal(t, tc.expectedOutput, buf.String())
			}
		})
	}
}

// Test_discovery_StatusCheckOfClients tests the discoveryStatusCheckOfClients() function.
func Test_discovery_StatusCheckOfClients(t *testing.T) {
	tests := []struct {
		name           string
		clients        []StatusCheckClientEntry
		printer        Printer
		expectedOutput string
		outWriter      io.Writer
		err            error
	}{
		// Working cases
		{
			name:           "StatusCheckOfClients correctly prints with the pretty printer when one of the status checks fails",
			clients:        []StatusCheckClientEntry{{Name: "core", Client: new(WorkingStatusChecker)}, {Name: "ingestion", Client: new(FailingStatusChecker)}, {Name: "queryflow", Client: new(WorkingStatusChecker)}, {Name: "staging", Client: new(WorkingStatusChecker)}},
			printer:        nil,
			expectedOutput: "",
			err:            nil,
		},
		{
			name:           "StatusCheckOfClients correctly prints the results with the ugly printer when one of the status checks fails",
			clients:        []StatusCheckClientEntry{{Name: "core", Client: new(WorkingStatusChecker)}, {Name: "ingestion", Client: new(WorkingStatusChecker)}, {Name: "queryflow", Client: new(FailingStatusChecker)}, {Name: "staging", Client: new(WorkingStatusChecker)}},
			printer:        JsonObjectPrinter(false),
			expectedOutput: "",
			err:            nil,
		},
		// Error cases
		{
			name:           "A working status has an invalid field name.",
			clients:        []StatusCheckClientEntry{{Name: "", Client: new(WorkingStatusChecker)}, {Name: "ingestion", Client: new(FailingStatusChecker)}, {Name: "queryflow", Client: new(WorkingStatusChecker)}, {Name: "staging", Client: new(WorkingStatusChecker)}},
			printer:        nil,
			expectedOutput: "",
			err:            NewErrorWithCause(ErrorExitCode, errors.New("path cannot be empty"), "Could not get the status of the Discovery products"),
		},
		{
			name:           "A failing status has an invalid field name.",
			clients:        []StatusCheckClientEntry{{Name: "core", Client: new(WorkingStatusChecker)}, {Name: "", Client: new(FailingStatusChecker)}, {Name: "queryflow", Client: new(WorkingStatusChecker)}, {Name: "staging", Client: new(WorkingStatusChecker)}},
			printer:        nil,
			expectedOutput: "",
			err:            NewErrorWithCause(ErrorExitCode, errors.New("path cannot be empty"), "Could not get the status of the Discovery products"),
		},
		{
			name:      "Printing fails",
			clients:   []StatusCheckClientEntry{{Name: "core", Client: new(WorkingStatusChecker)}, {Name: "ingestion", Client: new(FailingStatusChecker)}, {Name: "queryflow", Client: new(WorkingStatusChecker)}, {Name: "staging", Client: new(WorkingStatusChecker)}},
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
			err := d.StatusCheckOfClients(tc.clients, tc.printer)

			if tc.err != nil {
				var errStruct Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
