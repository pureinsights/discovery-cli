package cli

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/google/uuid"
	discoveryPackage "github.com/pureinsights/pdp-cli/discovery"
	"github.com/pureinsights/pdp-cli/internal/iostreams"
	"github.com/pureinsights/pdp-cli/internal/testutils"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// SearcherIDNotUUID simulates when the searcher returns a result with an ID that is not a UUID.
type SearcherIDNotUUID struct {
	mock.Mock
}

// Search implements the searcher interface
func (s *SearcherIDNotUUID) Search(gjson.Result) ([]gjson.Result, error) {
	return []gjson.Result(nil), discoveryPackage.Error{
		Status: http.StatusNotFound,
		Body:   gjson.Result{},
	}
}

// SearchByName returns a result with an ID that is not a UUID so that the conversion can fail.
func (s *SearcherIDNotUUID) SearchByName(name string) (gjson.Result, error) {
	return gjson.Parse(`{
			"type": "mongo",
			"name": "MongoDB Atlas seed clone",
			"labels": [],
			"active": true,
			"id": "test",
			"creationTimestamp": "2025-09-29T15:50:17Z",
			"lastUpdatedTimestamp": "2025-09-29T15:50:17Z"
		}`), nil
}

// Get implements the Searcher interface
func (s *SearcherIDNotUUID) Get(id uuid.UUID) (gjson.Result, error) {
	return gjson.Result{}, discoveryPackage.Error{
		Status: http.StatusNotFound,
		Body: gjson.Parse(`{
			"status": 404,
			"code": 1003,
			"messages": [
				"Seed not found: 986ce864-af76-4fcb-8b4f-f4e4c6ab0951"
			],
			"timestamp": "2025-10-16T00:15:31.888410500Z"
		}`),
	}
}

// GetAll implements the searcher interface
func (s *SearcherIDNotUUID) GetAll() ([]gjson.Result, error) {
	return []gjson.Result(nil), discoveryPackage.Error{Status: http.StatusUnauthorized, Body: gjson.Parse(`{"error":"unauthorized"}`)}
}

// TestGetSeedId tests the GetSeedId() function.
func TestGetSeedId(t *testing.T) {
	successId, successErr := uuid.Parse("986ce864-af76-4fcb-8b4f-f4e4c6ab0951")
	errorId, errorErr := uuid.Parse("test")
	tests := []struct {
		name     string
		client   Searcher
		expected uuid.UUID
		err      error
	}{
		// Working case
		{
			name:     "GetSeedId works",
			client:   new(WorkingSearcher),
			expected: successId,
			err:      successErr,
		},

		// Error case
		{
			name:     "Cannot convert to UUID fails",
			client:   new(SearcherIDNotUUID),
			expected: errorId,
			err:      errorErr,
		},
		{
			name:     "Search fails",
			client:   new(FailingSearcher),
			expected: uuid.Nil,
			err: discoveryPackage.Error{
				Status: http.StatusBadRequest,
				Body: gjson.Parse(`{
	"status": 400,
	"code": 3002,
	"messages": [
		"Invalid JSON: Unexpected end-of-input:"
	],
	"timestamp": "2025-10-17T17:43:52.817308100Z"
	}`),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buf := &bytes.Buffer{}

			ios := iostreams.IOStreams{
				In:  os.Stdin,
				Out: buf,
				Err: os.Stderr,
			}

			d := NewDiscovery(&ios, viper.New(), "")
			seedId, err := GetSeedId(d, tc.client, "seed")

			if tc.err != nil {
				require.Error(t, err)
				assert.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, seedId)
			}
		})
	}
}

// WorkingSeedController simulates a working IngestionSeedController
type WorkingSeedController struct {
	mock.Mock
	WorkingSearcher
}

// Start returns the result of a new seed execution
func (c *WorkingSeedController) Start(id uuid.UUID, scan discoveryPackage.ScanType, executionProperties gjson.Result) (gjson.Result, error) {
	return gjson.Parse(`{"id":"a056c7fb-0ca1-45f6-97ea-ec849a0701fd","creationTimestamp":"2025-09-04T19:29:41.119013Z","lastUpdatedTimestamp":"2025-09-04T19:29:41.119013Z","triggerType":"MANUAL","status":"CREATED","scanType":"INCREMENTAL","properties":{"stagingBucket":"testBucket"}}`), nil
}

// Halt returns the results of halting a seed
func (c *WorkingSeedController) Halt(id uuid.UUID) ([]gjson.Result, error) {
	return gjson.Parse(`[{"id":"a056c7fb-0ca1-45f6-97ea-ec849a0701fd","status":202}, {"id":"365d3ce3-4ea6-47a8-ada5-4ab4bedcbb3b","status":202}]`).Array(), nil
}

// FailingSeedControllerGetSeedIdFails simulates a failing IngestionSeedController when GetSeedId fails.
type FailingSeedControllerGetSeedIdFails struct {
	mock.Mock
	SearcherIDNotUUID
}

// Start implements the interface.
func (c *FailingSeedControllerGetSeedIdFails) Start(id uuid.UUID, scan discoveryPackage.ScanType, executionProperties gjson.Result) (gjson.Result, error) {
	return gjson.Parse(`{"id":"a056c7fb-0ca1-45f6-97ea-ec849a0701fd","creationTimestamp":"2025-09-04T19:29:41.119013Z","lastUpdatedTimestamp":"2025-09-04T19:29:41.119013Z","triggerType":"MANUAL","status":"CREATED","scanType":"INCREMENTAL","properties":{"stagingBucket":"testBucket"}}`), nil
}

// Halt implements the interface
func (c *FailingSeedControllerGetSeedIdFails) Halt(id uuid.UUID) ([]gjson.Result, error) {
	return gjson.Parse(`[{"id":"a056c7fb-0ca1-45f6-97ea-ec849a0701fd","status":202}, {"id":"365d3ce3-4ea6-47a8-ada5-4ab4bedcbb3b","status":202}]`).Array(), nil
}

// FailingSeedControllerStartFails simulates when starting a seed execution fails.
type FailingSeedControllerStartFails struct {
	mock.Mock
	WorkingSearcher
}

// Start mocks a failing seed execution response.
func (c *FailingSeedControllerStartFails) Start(id uuid.UUID, scan discoveryPackage.ScanType, executionProperties gjson.Result) (gjson.Result, error) {
	return gjson.Result{}, discoveryPackage.Error{Status: http.StatusConflict, Body: gjson.Parse(`{
			"status": 409,
			"code": 4001,
			"messages": [
				"The seed has 1 executions: 0c309dbb-0402-4710-8659-2c75f5d649b6"
			],
			"timestamp": "2025-09-04T20:17:00.116546400Z"
			}`)}
}

// Halt implements the IngestionSeedController interface
func (c *FailingSeedControllerStartFails) Halt(id uuid.UUID) ([]gjson.Result, error) {
	return gjson.Parse(`[{"id":"a056c7fb-0ca1-45f6-97ea-ec849a0701fd","status":202}, {"id":"365d3ce3-4ea6-47a8-ada5-4ab4bedcbb3b","status":202}]`).Array(), nil
}

// Test_discovery_StartSeed tests the discovery.StartSeed() function.
func Test_discovery_StartSeed(t *testing.T) {
	tests := []struct {
		name           string
		client         IngestionSeedController
		printer        Printer
		expectedOutput string
		outWriter      io.Writer
		err            error
	}{
		// Working case
		{
			name:           "StartSeed correctly prints the received object with the sent printer",
			client:         new(WorkingSeedController),
			printer:        JsonObjectPrinter(true),
			expectedOutput: "{\n  \"creationTimestamp\": \"2025-09-04T19:29:41.119013Z\",\n  \"id\": \"a056c7fb-0ca1-45f6-97ea-ec849a0701fd\",\n  \"lastUpdatedTimestamp\": \"2025-09-04T19:29:41.119013Z\",\n  \"properties\": {\n    \"stagingBucket\": \"testBucket\"\n  },\n  \"scanType\": \"INCREMENTAL\",\n  \"status\": \"CREATED\",\n  \"triggerType\": \"MANUAL\"\n}\n",
			err:            nil,
		},
		{
			name:           "StartSeed correctly prints the received object with JSON ugly printer",
			client:         new(WorkingSeedController),
			printer:        nil,
			expectedOutput: "{\"creationTimestamp\":\"2025-09-04T19:29:41.119013Z\",\"id\":\"a056c7fb-0ca1-45f6-97ea-ec849a0701fd\",\"lastUpdatedTimestamp\":\"2025-09-04T19:29:41.119013Z\",\"properties\":{\"stagingBucket\":\"testBucket\"},\"scanType\":\"INCREMENTAL\",\"status\":\"CREATED\",\"triggerType\":\"MANUAL\"}\n",
			err:            nil,
		},

		// Error case
		{
			name:           "GetByIdFails",
			client:         new(FailingSeedControllerGetSeedIdFails),
			printer:        nil,
			expectedOutput: "",
			err:            NewErrorWithCause(ErrorExitCode, errors.New("invalid UUID length: 4"), "Could not get seed ID to start execution."),
		},
		{
			name:           "Start fails because of conflict",
			client:         new(FailingSeedControllerStartFails),
			printer:        nil,
			expectedOutput: "",
			err: NewErrorWithCause(ErrorExitCode, discoveryPackage.Error{Status: http.StatusConflict, Body: gjson.Parse(`{
			"status": 409,
			"code": 4001,
			"messages": [
				"The seed has 1 executions: 0c309dbb-0402-4710-8659-2c75f5d649b6"
			],
			"timestamp": "2025-09-04T20:17:00.116546400Z"
			}`)}, "Could not start seed execution for seed with id \"986ce864-af76-4fcb-8b4f-f4e4c6ab0951\""),
		},
		{
			name:      "Printing fails",
			client:    new(WorkingSeedController),
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
			err := d.StartSeed(tc.client, "", discoveryPackage.ScanFull, gjson.Result{}, tc.printer)

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

// Test_discovery_HaltSeed tests the discovery.HaltSeed() function.
func Test_discovery_HaltSeed(t *testing.T) {
	tests := []struct {
		name           string
		client         IngestionSeedController
		printer        Printer
		expectedOutput string
		outWriter      io.Writer
		err            error
	}{
		// Working case
		{
			name:           "HaltSeed correctly prints the stopped executions with the sent printer",
			client:         new(WorkingSeedController),
			printer:        JsonArrayPrinter(true),
			expectedOutput: "{\n  \"creationTimestamp\": \"2025-09-04T19:29:41.119013Z\",\n  \"id\": \"a056c7fb-0ca1-45f6-97ea-ec849a0701fd\",\n  \"lastUpdatedTimestamp\": \"2025-09-04T19:29:41.119013Z\",\n  \"properties\": {\n    \"stagingBucket\": \"testBucket\"\n  },\n  \"scanType\": \"INCREMENTAL\",\n  \"status\": \"CREATED\",\n  \"triggerType\": \"MANUAL\"\n}\n",
			err:            nil,
		},
		{
			name:           "HaltSeed prints the halted executions with the ugly printer",
			client:         new(WorkingSeedController),
			printer:        nil,
			expectedOutput: "{\"creationTimestamp\":\"2025-09-04T19:29:41.119013Z\",\"id\":\"a056c7fb-0ca1-45f6-97ea-ec849a0701fd\",\"lastUpdatedTimestamp\":\"2025-09-04T19:29:41.119013Z\",\"properties\":{\"stagingBucket\":\"testBucket\"},\"scanType\":\"INCREMENTAL\",\"status\":\"CREATED\",\"triggerType\":\"MANUAL\"}\n",
			err:            nil,
		},

		// Error case
		{
			name:           "GetByIdFails",
			client:         new(FailingSeedControllerGetSeedIdFails),
			printer:        nil,
			expectedOutput: "",
			err:            NewErrorWithCause(ErrorExitCode, errors.New("invalid UUID length: 4"), "Could not get seed ID to start execution."),
		},
		{
			name:           "Start fails because of conflict",
			client:         new(FailingSeedControllerStartFails),
			printer:        nil,
			expectedOutput: "",
			err: NewErrorWithCause(ErrorExitCode, discoveryPackage.Error{Status: http.StatusConflict, Body: gjson.Parse(`{
			"status": 409,
			"code": 4001,
			"messages": [
				"The seed has 1 executions: 0c309dbb-0402-4710-8659-2c75f5d649b6"
			],
			"timestamp": "2025-09-04T20:17:00.116546400Z"
			}`)}, "Could not start seed execution for seed with id \"986ce864-af76-4fcb-8b4f-f4e4c6ab0951\""),
		},
		{
			name:      "Printing fails",
			client:    new(WorkingSeedController),
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
			err := d.StartSeed(tc.client, "", discoveryPackage.ScanFull, gjson.Result{}, tc.printer)

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
