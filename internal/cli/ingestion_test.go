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
	return []gjson.Result{}, discoveryPackage.Error{
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
			name:           "StartSeed correctly prints the received object with the pretty printer",
			client:         new(WorkingSeedController),
			printer:        nil,
			expectedOutput: "{\n  \"creationTimestamp\": \"2025-09-04T19:29:41.119013Z\",\n  \"id\": \"a056c7fb-0ca1-45f6-97ea-ec849a0701fd\",\n  \"lastUpdatedTimestamp\": \"2025-09-04T19:29:41.119013Z\",\n  \"properties\": {\n    \"stagingBucket\": \"testBucket\"\n  },\n  \"scanType\": \"INCREMENTAL\",\n  \"status\": \"CREATED\",\n  \"triggerType\": \"MANUAL\"\n}\n",
			err:            nil,
		},
		{
			name:           "StartSeed correctly prints the received object with JSON ugly printer",
			client:         new(WorkingSeedController),
			printer:        JsonObjectPrinter(false),
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
			name:           "HaltSeed correctly prints the stopped executions with the pretty printer",
			client:         new(WorkingSeedController),
			printer:        JsonArrayPrinter(true),
			expectedOutput: "[\n{\n  \"id\": \"a056c7fb-0ca1-45f6-97ea-ec849a0701fd\",\n  \"status\": 202\n},\n{\n  \"id\": \"365d3ce3-4ea6-47a8-ada5-4ab4bedcbb3b\",\n  \"status\": 202\n}\n]\n",
			err:            nil,
		},
		{
			name:           "HaltSeed prints the halted executions with the ugly printer",
			client:         new(WorkingSeedController),
			printer:        nil,
			expectedOutput: "{\"id\":\"a056c7fb-0ca1-45f6-97ea-ec849a0701fd\",\"status\":202}\n{\"id\":\"365d3ce3-4ea6-47a8-ada5-4ab4bedcbb3b\",\"status\":202}\n",
			err:            nil,
		},

		// Error case
		{
			name:           "GetByIdFails",
			client:         new(FailingSeedControllerGetSeedIdFails),
			printer:        nil,
			expectedOutput: "",
			err:            NewErrorWithCause(ErrorExitCode, errors.New("invalid UUID length: 4"), "Could not get seed ID to halt execution."),
		},
		{
			name:           "Halt fails seed not found",
			client:         new(FailingSeedControllerStartFails),
			printer:        nil,
			expectedOutput: "",
			err: NewErrorWithCause(ErrorExitCode, discoveryPackage.Error{
				Status: http.StatusNotFound,
				Body: gjson.Parse(`{
			"status": 404,
			"code": 1003,
			"messages": [
				"Seed not found: 986ce864-af76-4fcb-8b4f-f4e4c6ab0951"
			],
			"timestamp": "2025-10-16T00:15:31.888410500Z"
		}`)}, "Could not halt seed execution for seed with id \"986ce864-af76-4fcb-8b4f-f4e4c6ab0951\""),
		},
		{
			name:      "Printing fails",
			client:    new(WorkingSeedController),
			printer:   nil,
			outWriter: testutils.ErrWriter{Err: errors.New("write failed")},
			err:       NewErrorWithCause(ErrorExitCode, errors.New("write failed"), "Could not print JSON Array"),
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
			err := d.HaltSeed(tc.client, "", tc.printer)

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

// WorkingSeedController simulates a working IngestionSeedController
type WorkingSeedExecutionController struct {
	mock.Mock
	WorkingGetter
}

// Halt returns the results of halting a seed
func (c *WorkingSeedExecutionController) Halt(id uuid.UUID) (gjson.Result, error) {
	return gjson.Parse(`{"acknowledged":true}`), nil
}

// FailingSeedControllerStartFails simulates when starting a seed execution fails.
type FailingSeedExecutionControllerHaltFails struct {
	mock.Mock
	WorkingGetter
}

// Halt returns the results of halting a seed
func (c *FailingSeedExecutionControllerHaltFails) Halt(id uuid.UUID) (gjson.Result, error) {
	return gjson.Result{}, discoveryPackage.Error{Status: http.StatusConflict, Body: gjson.Parse(`{
			"status": 409,
			"code": 4001,
			"messages": [
				"Action HALT cannot be applied to seed execution cc89b714-d00a-4774-9c45-9497b5d9f8ef because of its current status: HALTING"
			],
			"timestamp": "2025-09-03T21:05:21.861757200Z"
			}`)}
}

// Test_discovery_HaltSeedExecution tests the discovery.HaltSeedExecution() function.
func Test_discovery_HaltSeedExecution(t *testing.T) {
	tests := []struct {
		name           string
		client         IngestionSeedExecutionController
		printer        Printer
		expectedOutput string
		outWriter      io.Writer
		err            error
	}{
		// Working case
		{
			name:           "HaltSeedExecution correctly prints the received object with the pretty printer",
			client:         new(WorkingSeedExecutionController),
			printer:        nil,
			expectedOutput: "{\n  \"acknowledged\": true\n}\n",
			err:            nil,
		},
		{
			name:           "StartSeed correctly prints the received object with JSON ugly printer",
			client:         new(WorkingSeedExecutionController),
			printer:        JsonObjectPrinter(false),
			expectedOutput: "{\"acknowledged\":true}\n",
			err:            nil,
		},

		// Error case
		{
			name:           "Halt fails seed not found",
			client:         new(FailingSeedExecutionControllerHaltFails),
			printer:        nil,
			expectedOutput: "",
			err: NewErrorWithCause(ErrorExitCode, discoveryPackage.Error{Status: http.StatusConflict, Body: gjson.Parse(`{
			"status": 409,
			"code": 4001,
			"messages": [
				"Action HALT cannot be applied to seed execution cc89b714-d00a-4774-9c45-9497b5d9f8ef because of its current status: HALTING"
			],
			"timestamp": "2025-09-03T21:05:21.861757200Z"
			}`)}, "Could not halt the seed execution with id \"cc89b714-d00a-4774-9c45-9497b5d9f8ef\""),
		},
		{
			name:      "Printing fails",
			client:    new(WorkingSeedExecutionController),
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

			executionId, err := uuid.Parse("cc89b714-d00a-4774-9c45-9497b5d9f8ef")
			require.NoError(t, err)

			d := NewDiscovery(&ios, viper.New(), "")
			err = d.HaltSeedExecution(tc.client, executionId, tc.printer)

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

// WorkingGetter mocks the RecordGetter interface to always answer a working result
type WorkingRecordGetter struct {
	mock.Mock
}

// Get returns a record as if the request worked successfully.
func (g *WorkingRecordGetter) Get(id string) (gjson.Result, error) {
	return gjson.Parse(`{
  "id": {
    "plain": "4e7c8a47efd829ef7f710d64da661786",
    "hash": "A3HTDEgCa65BFZsac9TInFisvloRlL3M50ijCWNCKx0="
  },
  "creationTimestamp": "2025-09-03T21:02:54Z",
  "lastUpdatedTimestamp": "2025-09-03T21:02:54Z",
  "status": "SUCCESS"
}`), nil
}

// GetAll returns a list of records
func (g *WorkingRecordGetter) GetAll() ([]gjson.Result, error) {
	return gjson.Parse(`[
		{"id":{"plain":"4e7c8a47efd829ef7f710d64da661786","hash":"A3HTDEgCa65BFZsac9TInFisvloRlL3M50ijCWNCKx0="},"creationTimestamp":"2025-09-05T20:13:47Z","lastUpdatedTimestamp":"2025-09-05T20:13:47Z","status":"SUCCESS"},
		{"id":{"plain":"8148e6a7b952a3b2964f706ced8c6885","hash":"IJeF-losyj33EAuqjgGW2G7sT-eE7poejQ5HokerZio="},"creationTimestamp":"2025-09-05T20:13:47Z","lastUpdatedTimestamp":"2025-09-05T20:13:47Z","status":"SUCCESS"},
		{"id":{"plain":"b1e3e4f42c0818b1580e306eb776d4a1","hash":"N2lubqCWTqEEaymQVntpdP5dqKDP-LYk81C_PCr6btQ="},"creationTimestamp":"2025-09-05T20:13:47Z","lastUpdatedTimestamp":"2025-09-05T20:13:47Z","status":"SUCCESS"}
	]`).Array(), nil
}

// FailingRecordGetter mocks the RecordGetter struct to always return an HTTP error.
type FailingRecordGetter struct {
	mock.Mock
}

// Get returns a 404 Not Found
func (g *FailingRecordGetter) Get(id string) (gjson.Result, error) {
	return gjson.Result{}, discoveryPackage.Error{
		Status: http.StatusNotFound,
		Body: gjson.Parse(`{
  "status": 404,
  "code": 1003,
  "messages": [
    "Entity not found: SeedRecordId(seed=Seed(super=AbstractComponentConfigEntity(super=AbstractJsonConfigEntity(super=AbstractTypedConfigEntity(super=AbstractConfigEntity(super=AbstractUpdatableEntity(super=AbstractCoreEntity(id=2acd0a61-852c-4f38-af2b-9c84e152873e), creationTimestamp=2025-08-21T21:52:03Z, lastUpdatedTimestamp=2025-08-21T21:52:03Z), name=Search seed, description=null, active=true), type=staging), config={\"action\":\"scroll\",\"bucket\":\"blogs\"})), properties=null, labels=[], recordOptions=SeedRecordPolicy[timeoutPolicy=TimeoutPolicy[slice=PT1H], errorPolicy=FATAL, outboundPolicy=OutboundPolicy[idPolicy=IdPolicy[generator=null], batchPolicy=BatchPolicy[maxCount=25, flushAfter=PT1M]]], hooks=[], beforeHooksOptions=null, afterHooksOptions=null), recordId=[3, 113, -45, 12, 72, 2, 107, -82, 65, 21, -101, 26, 115, -44, -56, -100, 88, -84, -66, 90, 17, -108, -67, -52, -25, 72, -93, 9, 99, 66, 43, 31])"
  ],
  "timestamp": "2025-11-09T14:42:48.411373100Z"
}`),
	}
}

// GetAll returns 401 unauthorized
func (g *FailingRecordGetter) GetAll() ([]gjson.Result, error) {
	return []gjson.Result(nil), discoveryPackage.Error{Status: http.StatusUnauthorized, Body: gjson.Parse(`{"error":"unauthorized"}`)}
}

// TestAppendSeedRecord tests the AppendSeedRecord function()
func TestAppendSeedRecord(t *testing.T) {
	tests := []struct {
		name           string
		client         RecordGetter
		id             string
		expectedRecord string
		err            error
	}{
		{
			name:   "Getting the ID and setting the record field works",
			client: new(WorkingRecordGetter),
			id:     "A3HTDEgCa65BFZsac9TInFisvloRlL3M50ijCWNCKx0=",
			expectedRecord: `{
  "id": {
    "plain": "4e7c8a47efd829ef7f710d64da661786",
    "hash": "A3HTDEgCa65BFZsac9TInFisvloRlL3M50ijCWNCKx0="
  },
  "creationTimestamp": "2025-09-03T21:02:54Z",
  "lastUpdatedTimestamp": "2025-09-03T21:02:54Z",
  "status": "SUCCESS"
}`,
			err: nil,
		},
		{
			name:           "Getting the ID fails",
			client:         new(FailingRecordGetter),
			id:             "A3HTDEgCa65BFZsac9TInFisvloRlL3M50ijCWNCKx0=",
			expectedRecord: "",
			err: discoveryPackage.Error{
				Status: http.StatusNotFound,
				Body: gjson.Parse(`{
  "status": 404,
  "code": 1003,
  "messages": [
    "Entity not found: SeedRecordId(seed=Seed(super=AbstractComponentConfigEntity(super=AbstractJsonConfigEntity(super=AbstractTypedConfigEntity(super=AbstractConfigEntity(super=AbstractUpdatableEntity(super=AbstractCoreEntity(id=2acd0a61-852c-4f38-af2b-9c84e152873e), creationTimestamp=2025-08-21T21:52:03Z, lastUpdatedTimestamp=2025-08-21T21:52:03Z), name=Search seed, description=null, active=true), type=staging), config={\"action\":\"scroll\",\"bucket\":\"blogs\"})), properties=null, labels=[], recordOptions=SeedRecordPolicy[timeoutPolicy=TimeoutPolicy[slice=PT1H], errorPolicy=FATAL, outboundPolicy=OutboundPolicy[idPolicy=IdPolicy[generator=null], batchPolicy=BatchPolicy[maxCount=25, flushAfter=PT1M]]], hooks=[], beforeHooksOptions=null, afterHooksOptions=null), recordId=[3, 113, -45, 12, 72, 2, 107, -82, 65, 21, -101, 26, 115, -44, -56, -100, 88, -84, -66, 90, 17, -108, -67, -52, -25, 72, -93, 9, 99, 66, 43, 31])"
  ],
  "timestamp": "2025-11-09T14:42:48.411373100Z"
}`),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			seed := gjson.Parse(`{
  "type": "staging",
  "name": "Search seed",
  "labels": [],
  "active": true,
  "id": "2acd0a61-852c-4f38-af2b-9c84e152873e",
  "creationTimestamp": "2025-08-21T21:52:03Z",
  "lastUpdatedTimestamp": "2025-08-21T21:52:03Z",
  "config": {
    "action": "scroll",
    "bucket": "blogs"
  },
  "pipeline": "9a74bf3a-eb2a-4334-b803-c92bf1bc45fe",
  "recordPolicy": {
    "errorPolicy": "FATAL",
    "timeoutPolicy": {
      "slice": "PT1H"
    },
    "outboundPolicy": {
      "idPolicy": {},
      "batchPolicy": {
        "maxCount": 25,
        "flushAfter": "PT1M"
      }
    }
  }
}`)
			result, err := AppendSeedRecord(seed, tc.client, tc.id)
			assert.Equal(t, tc.expectedRecord, result.Get("record").Raw)
			if tc.err != nil {
				require.Error(t, err)
				assert.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Test_discovery_AppendSeedRecord tests the discovery.AppendSeedRecord function
func Test_discovery_AppendSeedRecord(t *testing.T) {
	tests := []struct {
		name           string
		client         RecordGetter
		printer        Printer
		id             string
		expectedOutput string
		outWriter      io.Writer
		err            error
	}{
		// Working cases
		{
			name:           "Getting the ID and printing the result with pretty printer works",
			client:         new(WorkingRecordGetter),
			printer:        nil,
			id:             "A3HTDEgCa65BFZsac9TInFisvloRlL3M50ijCWNCKx0=",
			expectedOutput: "{\n  \"active\": true,\n  \"config\": {\n    \"action\": \"scroll\",\n    \"bucket\": \"blogs\"\n  },\n  \"creationTimestamp\": \"2025-08-21T21:52:03Z\",\n  \"id\": \"2acd0a61-852c-4f38-af2b-9c84e152873e\",\n  \"labels\": [],\n  \"lastUpdatedTimestamp\": \"2025-08-21T21:52:03Z\",\n  \"name\": \"Search seed\",\n  \"pipeline\": \"9a74bf3a-eb2a-4334-b803-c92bf1bc45fe\",\n  \"record\": {\n    \"creationTimestamp\": \"2025-09-03T21:02:54Z\",\n    \"id\": {\n      \"hash\": \"A3HTDEgCa65BFZsac9TInFisvloRlL3M50ijCWNCKx0=\",\n      \"plain\": \"4e7c8a47efd829ef7f710d64da661786\"\n    },\n    \"lastUpdatedTimestamp\": \"2025-09-03T21:02:54Z\",\n    \"status\": \"SUCCESS\"\n  },\n  \"recordPolicy\": {\n    \"errorPolicy\": \"FATAL\",\n    \"outboundPolicy\": {\n      \"batchPolicy\": {\n        \"flushAfter\": \"PT1M\",\n        \"maxCount\": 25\n      },\n      \"idPolicy\": {}\n    },\n    \"timeoutPolicy\": {\n      \"slice\": \"PT1H\"\n    }\n  },\n  \"type\": \"staging\"\n}\n",
			err:            nil,
		},
		{
			name:           "Getting the ID and printing the result with ugly printer works",
			client:         new(WorkingRecordGetter),
			printer:        JsonObjectPrinter(false),
			id:             "A3HTDEgCa65BFZsac9TInFisvloRlL3M50ijCWNCKx0=",
			expectedOutput: "{\"active\":true,\"config\":{\"action\":\"scroll\",\"bucket\":\"blogs\"},\"creationTimestamp\":\"2025-08-21T21:52:03Z\",\"id\":\"2acd0a61-852c-4f38-af2b-9c84e152873e\",\"labels\":[],\"lastUpdatedTimestamp\":\"2025-08-21T21:52:03Z\",\"name\":\"Search seed\",\"pipeline\":\"9a74bf3a-eb2a-4334-b803-c92bf1bc45fe\",\"record\":{\"creationTimestamp\":\"2025-09-03T21:02:54Z\",\"id\":{\"hash\":\"A3HTDEgCa65BFZsac9TInFisvloRlL3M50ijCWNCKx0=\",\"plain\":\"4e7c8a47efd829ef7f710d64da661786\"},\"lastUpdatedTimestamp\":\"2025-09-03T21:02:54Z\",\"status\":\"SUCCESS\"},\"recordPolicy\":{\"errorPolicy\":\"FATAL\",\"outboundPolicy\":{\"batchPolicy\":{\"flushAfter\":\"PT1M\",\"maxCount\":25},\"idPolicy\":{}},\"timeoutPolicy\":{\"slice\":\"PT1H\"}},\"type\":\"staging\"}\n",
			err:            nil,
		},
		// Error cases
		{
			name:           "Getting the ID fails",
			client:         new(FailingRecordGetter),
			id:             "A3HTDEgCa65BFZsac9TInFisvloRlL3M50ijCWNCKx0=",
			expectedOutput: "",
			err: NewErrorWithCause(ErrorExitCode, discoveryPackage.Error{
				Status: http.StatusNotFound,
				Body: gjson.Parse(`{
  "status": 404,
  "code": 1003,
  "messages": [
    "Entity not found: SeedRecordId(seed=Seed(super=AbstractComponentConfigEntity(super=AbstractJsonConfigEntity(super=AbstractTypedConfigEntity(super=AbstractConfigEntity(super=AbstractUpdatableEntity(super=AbstractCoreEntity(id=2acd0a61-852c-4f38-af2b-9c84e152873e), creationTimestamp=2025-08-21T21:52:03Z, lastUpdatedTimestamp=2025-08-21T21:52:03Z), name=Search seed, description=null, active=true), type=staging), config={\"action\":\"scroll\",\"bucket\":\"blogs\"})), properties=null, labels=[], recordOptions=SeedRecordPolicy[timeoutPolicy=TimeoutPolicy[slice=PT1H], errorPolicy=FATAL, outboundPolicy=OutboundPolicy[idPolicy=IdPolicy[generator=null], batchPolicy=BatchPolicy[maxCount=25, flushAfter=PT1M]]], hooks=[], beforeHooksOptions=null, afterHooksOptions=null), recordId=[3, 113, -45, 12, 72, 2, 107, -82, 65, 21, -101, 26, 115, -44, -56, -100, 88, -84, -66, 90, 17, -108, -67, -52, -25, 72, -93, 9, 99, 66, 43, 31])"
  ],
  "timestamp": "2025-11-09T14:42:48.411373100Z"
}`),
			}, "Could not get record with id \"A3HTDEgCa65BFZsac9TInFisvloRlL3M50ijCWNCKx0=\""),
		},
		{
			name:      "Printing fails",
			client:    new(WorkingRecordGetter),
			printer:   nil,
			outWriter: testutils.ErrWriter{Err: errors.New("write failed")},
			err:       NewErrorWithCause(ErrorExitCode, errors.New("write failed"), "Could not print JSON object"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			seed := gjson.Parse(`{
  "type": "staging",
  "name": "Search seed",
  "labels": [],
  "active": true,
  "id": "2acd0a61-852c-4f38-af2b-9c84e152873e",
  "creationTimestamp": "2025-08-21T21:52:03Z",
  "lastUpdatedTimestamp": "2025-08-21T21:52:03Z",
  "config": {
    "action": "scroll",
    "bucket": "blogs"
  },
  "pipeline": "9a74bf3a-eb2a-4334-b803-c92bf1bc45fe",
  "recordPolicy": {
    "errorPolicy": "FATAL",
    "timeoutPolicy": {
      "slice": "PT1H"
    },
    "outboundPolicy": {
      "idPolicy": {},
      "batchPolicy": {
        "maxCount": 25,
        "flushAfter": "PT1M"
      }
    }
  }
}`)

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
			err := d.AppendSeedRecord(seed, tc.client, tc.id, tc.printer)

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

// TestAppendSeedRecords tests the AppendSeedRecords() function.
func TestAppendSeedRecords(t *testing.T) {
	tests := []struct {
		name            string
		client          RecordGetter
		expectedRecords string
		err             error
	}{
		// Working case
		{
			name:   "Getting the records and setting the record field works",
			client: new(WorkingRecordGetter),
			expectedRecords: `[
{"id":{"plain":"4e7c8a47efd829ef7f710d64da661786","hash":"A3HTDEgCa65BFZsac9TInFisvloRlL3M50ijCWNCKx0="},"creationTimestamp":"2025-09-05T20:13:47Z","lastUpdatedTimestamp":"2025-09-05T20:13:47Z","status":"SUCCESS"},
{"id":{"plain":"8148e6a7b952a3b2964f706ced8c6885","hash":"IJeF-losyj33EAuqjgGW2G7sT-eE7poejQ5HokerZio="},"creationTimestamp":"2025-09-05T20:13:47Z","lastUpdatedTimestamp":"2025-09-05T20:13:47Z","status":"SUCCESS"},
{"id":{"plain":"b1e3e4f42c0818b1580e306eb776d4a1","hash":"N2lubqCWTqEEaymQVntpdP5dqKDP-LYk81C_PCr6btQ="},"creationTimestamp":"2025-09-05T20:13:47Z","lastUpdatedTimestamp":"2025-09-05T20:13:47Z","status":"SUCCESS"}
]`,
			err: nil,
		},
		// Error case
		{
			name:            "Getting the records fails",
			client:          new(FailingRecordGetter),
			expectedRecords: "",
			err:             discoveryPackage.Error{Status: http.StatusUnauthorized, Body: gjson.Parse(`{"error":"unauthorized"}`)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			seed := gjson.Parse(`{
  "type": "staging",
  "name": "Search seed",
  "labels": [],
  "active": true,
  "id": "2acd0a61-852c-4f38-af2b-9c84e152873e",
  "creationTimestamp": "2025-08-21T21:52:03Z",
  "lastUpdatedTimestamp": "2025-08-21T21:52:03Z",
  "config": {
    "action": "scroll",
    "bucket": "blogs"
  },
  "pipeline": "9a74bf3a-eb2a-4334-b803-c92bf1bc45fe",
  "recordPolicy": {
    "errorPolicy": "FATAL",
    "timeoutPolicy": {
      "slice": "PT1H"
    },
    "outboundPolicy": {
      "idPolicy": {},
      "batchPolicy": {
        "maxCount": 25,
        "flushAfter": "PT1M"
      }
    }
  }
}`)
			result, err := AppendSeedRecords(seed, tc.client)

			if tc.err != nil {
				require.Error(t, err)
				assert.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedRecords, result.Get("records").Raw)
			}
		})
	}
}

// Test_discovery_AppendSeedRecords tests the discovery.AppendSeedRecords() function
func Test_discovery_AppendSeedRecords(t *testing.T) {
	tests := []struct {
		name           string
		client         RecordGetter
		printer        Printer
		expectedOutput string
		outWriter      io.Writer
		err            error
	}{
		// Working cases
		{
			name:    "Getting the records and printing the result with pretty printer works",
			client:  new(WorkingRecordGetter),
			printer: JsonObjectPrinter(true),
			expectedOutput: `{
  "active": true,
  "config": {
    "action": "scroll",
    "bucket": "blogs"
  },
  "creationTimestamp": "2025-08-21T21:52:03Z",
  "id": "2acd0a61-852c-4f38-af2b-9c84e152873e",
  "labels": [],
  "lastUpdatedTimestamp": "2025-08-21T21:52:03Z",
  "name": "Search seed",
  "pipeline": "9a74bf3a-eb2a-4334-b803-c92bf1bc45fe",
  "recordPolicy": {
    "errorPolicy": "FATAL",
    "outboundPolicy": {
      "batchPolicy": {
        "flushAfter": "PT1M",
        "maxCount": 25
      },
      "idPolicy": {}
    },
    "timeoutPolicy": {
      "slice": "PT1H"
    }
  },
  "records": [
    {
      "creationTimestamp": "2025-09-05T20:13:47Z",
      "id": {
        "hash": "A3HTDEgCa65BFZsac9TInFisvloRlL3M50ijCWNCKx0=",
        "plain": "4e7c8a47efd829ef7f710d64da661786"
      },
      "lastUpdatedTimestamp": "2025-09-05T20:13:47Z",
      "status": "SUCCESS"
    },
    {
      "creationTimestamp": "2025-09-05T20:13:47Z",
      "id": {
        "hash": "IJeF-losyj33EAuqjgGW2G7sT-eE7poejQ5HokerZio=",
        "plain": "8148e6a7b952a3b2964f706ced8c6885"
      },
      "lastUpdatedTimestamp": "2025-09-05T20:13:47Z",
      "status": "SUCCESS"
    },
    {
      "creationTimestamp": "2025-09-05T20:13:47Z",
      "id": {
        "hash": "N2lubqCWTqEEaymQVntpdP5dqKDP-LYk81C_PCr6btQ=",
        "plain": "b1e3e4f42c0818b1580e306eb776d4a1"
      },
      "lastUpdatedTimestamp": "2025-09-05T20:13:47Z",
      "status": "SUCCESS"
    }
  ],
  "type": "staging"
}
`,
			err: nil,
		},
		{
			name:           "Getting the records and printing the result with ugly printer works",
			printer:        JsonObjectPrinter(false),
			client:         new(WorkingRecordGetter),
			expectedOutput: "{\"active\":true,\"config\":{\"action\":\"scroll\",\"bucket\":\"blogs\"},\"creationTimestamp\":\"2025-08-21T21:52:03Z\",\"id\":\"2acd0a61-852c-4f38-af2b-9c84e152873e\",\"labels\":[],\"lastUpdatedTimestamp\":\"2025-08-21T21:52:03Z\",\"name\":\"Search seed\",\"pipeline\":\"9a74bf3a-eb2a-4334-b803-c92bf1bc45fe\",\"recordPolicy\":{\"errorPolicy\":\"FATAL\",\"outboundPolicy\":{\"batchPolicy\":{\"flushAfter\":\"PT1M\",\"maxCount\":25},\"idPolicy\":{}},\"timeoutPolicy\":{\"slice\":\"PT1H\"}},\"records\":[{\"creationTimestamp\":\"2025-09-05T20:13:47Z\",\"id\":{\"hash\":\"A3HTDEgCa65BFZsac9TInFisvloRlL3M50ijCWNCKx0=\",\"plain\":\"4e7c8a47efd829ef7f710d64da661786\"},\"lastUpdatedTimestamp\":\"2025-09-05T20:13:47Z\",\"status\":\"SUCCESS\"},{\"creationTimestamp\":\"2025-09-05T20:13:47Z\",\"id\":{\"hash\":\"IJeF-losyj33EAuqjgGW2G7sT-eE7poejQ5HokerZio=\",\"plain\":\"8148e6a7b952a3b2964f706ced8c6885\"},\"lastUpdatedTimestamp\":\"2025-09-05T20:13:47Z\",\"status\":\"SUCCESS\"},{\"creationTimestamp\":\"2025-09-05T20:13:47Z\",\"id\":{\"hash\":\"N2lubqCWTqEEaymQVntpdP5dqKDP-LYk81C_PCr6btQ=\",\"plain\":\"b1e3e4f42c0818b1580e306eb776d4a1\"},\"lastUpdatedTimestamp\":\"2025-09-05T20:13:47Z\",\"status\":\"SUCCESS\"}],\"type\":\"staging\"}\n",
			err:            nil,
		},
		// Error cases
		{
			name:           "Getting the records fails",
			client:         new(FailingRecordGetter),
			expectedOutput: "",
			err:            NewErrorWithCause(ErrorExitCode, discoveryPackage.Error{Status: http.StatusUnauthorized, Body: gjson.Parse(`{"error":"unauthorized"}`)}, "Could not get records"),
		},
		{
			name:      "Printing fails",
			client:    new(WorkingRecordGetter),
			printer:   nil,
			outWriter: testutils.ErrWriter{Err: errors.New("write failed")},
			err:       NewErrorWithCause(ErrorExitCode, errors.New("write failed"), "Could not print JSON object"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			seed := gjson.Parse(`{
  "type": "staging",
  "name": "Search seed",
  "labels": [],
  "active": true,
  "id": "2acd0a61-852c-4f38-af2b-9c84e152873e",
  "creationTimestamp": "2025-08-21T21:52:03Z",
  "lastUpdatedTimestamp": "2025-08-21T21:52:03Z",
  "config": {
    "action": "scroll",
    "bucket": "blogs"
  },
  "pipeline": "9a74bf3a-eb2a-4334-b803-c92bf1bc45fe",
  "recordPolicy": {
    "errorPolicy": "FATAL",
    "timeoutPolicy": {
      "slice": "PT1H"
    },
    "outboundPolicy": {
      "idPolicy": {},
      "batchPolicy": {
        "maxCount": 25,
        "flushAfter": "PT1M"
      }
    }
  }
}`)

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
			err := d.AppendSeedRecords(seed, tc.client, tc.printer)

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
