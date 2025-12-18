package cli

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/google/uuid"
	discoveryPackage "github.com/pureinsights/discovery-cli/discovery"
	"github.com/pureinsights/discovery-cli/internal/iostreams"
	"github.com/pureinsights/discovery-cli/internal/testutils"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// WorkingServerPinger simulates when a ping to a server worked.
type WorkingServerPinger struct {
	mock.Mock
}

// Ping returns the response of a working ping.
func (s *WorkingServerPinger) Ping(uuid.UUID) (gjson.Result, error) {
	return gjson.Parse(`{
  "acknowledged": true
}`), nil
}

// SearchByName returns a server result.
func (s *WorkingServerPinger) SearchByName(name string) (gjson.Result, error) {
	return gjson.Parse(`{
  "type": "mongo",
  "name": "MongoDB Atlas server",
  "labels": [],
  "active": true,
  "id": "21029da3-041c-43b5-a67e-870251f2f6a6",
  "creationTimestamp": "2025-11-20T00:06:05Z",
  "lastUpdatedTimestamp": "2025-11-20T00:06:05Z",
  "config": {
    "servers": [
      "mongodb+srv://cluster0.dleud.mongodb.net/"
    ],
    "connection": {
      "readTimeout": "30s",
      "connectTimeout": "1m"
    },
    "credentialId": "9ababe08-0b74-4672-bb7c-e7a8227d6d4c"
  }
}`), nil
}

// Search implements the searcher interface.
func (s *WorkingServerPinger) Search(gjson.Result) ([]gjson.Result, error) {
	return []gjson.Result(nil), discoveryPackage.Error{
		Status: http.StatusNotFound,
		Body:   gjson.Result{},
	}
}

// Get implements the Searcher interface.
func (s *WorkingServerPinger) Get(id uuid.UUID) (gjson.Result, error) {
	return gjson.Result{}, discoveryPackage.Error{
		Status: http.StatusNotFound,
		Body: gjson.Parse(`{
			"status": 404,
			"code": 1003,
			"messages": [
				"Server not found: 21029da3-041c-43b5-a67e-870251f2f6a6"
			],
			"timestamp": "2025-10-16T00:15:31.888410500Z"
		}`),
	}
}

// GetAll implements the searcher interface.
func (s *WorkingServerPinger) GetAll() ([]gjson.Result, error) {
	return []gjson.Result(nil), discoveryPackage.Error{Status: http.StatusUnauthorized, Body: gjson.Parse(`{"error":"unauthorized"}`)}
}

// FailingServerPinger simulates when a server that does not exist was pinged.
type FailingServerPingerServerNotFound struct {
	mock.Mock
}

// Ping returns the response of a failing ping.
func (s *FailingServerPingerServerNotFound) Ping(uuid.UUID) (gjson.Result, error) {
	return gjson.Result{}, discoveryPackage.Error{Status: http.StatusUnprocessableEntity, Body: gjson.Parse(`{
	"status": 422,
	"code": 4001,
	"messages": [
		"Client of type openai cannot be validated."
	],
	"timestamp": "2025-10-23T22:35:38.345647200Z"
	}`)}
}

// SearchByName returns a not found error.
func (s *FailingServerPingerServerNotFound) SearchByName(name string) (gjson.Result, error) {
	return gjson.Result{}, discoveryPackage.Error{
		Status: http.StatusNotFound,
		Body: gjson.Parse(`{
			"status": 404,
			"code": 1003,
			"messages": [
				"Server not found: 21029da3-041c-43b5-a67e-870251f2f6a6"
			],
			"timestamp": "2025-10-16T00:15:31.888410500Z"
		}`),
	}
}

// Search implements the searcher interface.
func (s *FailingServerPingerServerNotFound) Search(gjson.Result) ([]gjson.Result, error) {
	return []gjson.Result(nil), discoveryPackage.Error{
		Status: http.StatusNotFound,
		Body:   gjson.Result{},
	}
}

// Get implements the Searcher interface.
func (s *FailingServerPingerServerNotFound) Get(id uuid.UUID) (gjson.Result, error) {
	return gjson.Result{}, discoveryPackage.Error{
		Status: http.StatusNotFound,
		Body: gjson.Parse(`{
			"status": 404,
			"code": 1003,
			"messages": [
				"Server not found: 21029da3-041c-43b5-a67e-870251f2f6a6"
			],
			"timestamp": "2025-10-16T00:15:31.888410500Z"
		}`),
	}
}

// GetAll implements the searcher interface.
func (s *FailingServerPingerServerNotFound) GetAll() ([]gjson.Result, error) {
	return []gjson.Result(nil), discoveryPackage.Error{Status: http.StatusUnauthorized, Body: gjson.Parse(`{"error":"unauthorized"}`)}
}

// FailingServerPinger simulates when a ping to a server fails.
type FailingServerPingerPingFailed struct {
	mock.Mock
}

// Ping returns the response of a failing ping.
func (s *FailingServerPingerPingFailed) Ping(uuid.UUID) (gjson.Result, error) {
	return gjson.Result{}, discoveryPackage.Error{Status: http.StatusUnprocessableEntity, Body: gjson.Parse(`{
	"status": 422,
	"code": 4001,
	"messages": [
		"Client of type openai cannot be validated."
	],
	"timestamp": "2025-10-23T22:35:38.345647200Z"
	}`)}
}

// SearchByName returns a result of a server.
func (s *FailingServerPingerPingFailed) SearchByName(name string) (gjson.Result, error) {
	return gjson.Parse(`{
  "type": "mongo",
  "name": "MongoDB Atlas server",
  "labels": [],
  "active": true,
  "id": "21029da3-041c-43b5-a67e-870251f2f6a6",
  "creationTimestamp": "2025-11-20T00:06:05Z",
  "lastUpdatedTimestamp": "2025-11-20T00:06:05Z",
  "config": {
    "servers": [
      "mongodb+srv://cluster0.dleud.mongodb.net/"
    ],
    "connection": {
      "readTimeout": "30s",
      "connectTimeout": "1m"
    },
    "credentialId": "9ababe08-0b74-4672-bb7c-e7a8227d6d4c"
  }
}`), nil
}

// Search implements the searcher interface.
func (s *FailingServerPingerPingFailed) Search(gjson.Result) ([]gjson.Result, error) {
	return []gjson.Result(nil), discoveryPackage.Error{
		Status: http.StatusNotFound,
		Body:   gjson.Result{},
	}
}

// Get implements the Searcher interface.
func (s *FailingServerPingerPingFailed) Get(id uuid.UUID) (gjson.Result, error) {
	return gjson.Result{}, discoveryPackage.Error{
		Status: http.StatusNotFound,
		Body: gjson.Parse(`{
			"status": 404,
			"code": 1003,
			"messages": [
				"Server not found: 21029da3-041c-43b5-a67e-870251f2f6a6"
			],
			"timestamp": "2025-10-16T00:15:31.888410500Z"
		}`),
	}
}

// GetAll implements the searcher interface.
func (s *FailingServerPingerPingFailed) GetAll() ([]gjson.Result, error) {
	return []gjson.Result(nil), discoveryPackage.Error{Status: http.StatusUnauthorized, Body: gjson.Parse(`{"error":"unauthorized"}`)}
}

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
			client:         new(WorkingServerPinger),
			printer:        nil,
			expectedOutput: "{\n  \"acknowledged\": true\n}\n",
			err:            nil,
		},
		{
			name:           "StartSeed correctly prints the received object with JSON ugly printer",
			client:         new(WorkingServerPinger),
			printer:        JsonObjectPrinter(false),
			expectedOutput: "{\"acknowledged\":true}\n",
			err:            nil,
		},

		// Error case
		{
			name:           "GetEntityById fails",
			client:         new(FailingServerPingerServerNotFound),
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
			client:         new(FailingServerPingerPingFailed),
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
			client:    new(WorkingServerPinger),
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
