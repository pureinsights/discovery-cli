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

// TestConvertJSONArrayToString tests the ConvertJSONArrayToString() function.
func TestConvertJSONArrayToString(t *testing.T) {
	tests := []struct {
		name     string
		array    []gjson.Result
		expected string
	}{
		{
			name:  "Array with some elements.",
			array: gjson.Parse(`[{"id": 1},{"id": 2},{"id": 3}]`).Array(),
			expected: `[
{"id": 1},
{"id": 2},
{"id": 3}
]`,
		},
		{
			name:  "Array with no elements.",
			array: gjson.Result{}.Array(),
			expected: `[
]`,
		}, {
			name:  "Array with one element.",
			array: gjson.Parse(`[{"id": 1}]`).Array(),
			expected: `[
{"id": 1}
]`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := ConvertJSONArrayToString(tc.array)
			assert.Equal(t, tc.expected, actual)
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
			printer:        JsonObjectPrinter(true),
			id:             "A3HTDEgCa65BFZsac9TInFisvloRlL3M50ijCWNCKx0=",
			expectedOutput: "{\n  \"active\": true,\n  \"config\": {\n    \"action\": \"scroll\",\n    \"bucket\": \"blogs\"\n  },\n  \"creationTimestamp\": \"2025-08-21T21:52:03Z\",\n  \"id\": \"2acd0a61-852c-4f38-af2b-9c84e152873e\",\n  \"labels\": [],\n  \"lastUpdatedTimestamp\": \"2025-08-21T21:52:03Z\",\n  \"name\": \"Search seed\",\n  \"pipeline\": \"9a74bf3a-eb2a-4334-b803-c92bf1bc45fe\",\n  \"record\": {\n    \"creationTimestamp\": \"2025-09-03T21:02:54Z\",\n    \"id\": {\n      \"hash\": \"A3HTDEgCa65BFZsac9TInFisvloRlL3M50ijCWNCKx0=\",\n      \"plain\": \"4e7c8a47efd829ef7f710d64da661786\"\n    },\n    \"lastUpdatedTimestamp\": \"2025-09-03T21:02:54Z\",\n    \"status\": \"SUCCESS\"\n  },\n  \"recordPolicy\": {\n    \"errorPolicy\": \"FATAL\",\n    \"outboundPolicy\": {\n      \"batchPolicy\": {\n        \"flushAfter\": \"PT1M\",\n        \"maxCount\": 25\n      },\n      \"idPolicy\": {}\n    },\n    \"timeoutPolicy\": {\n      \"slice\": \"PT1H\"\n    }\n  },\n  \"type\": \"staging\"\n}\n",
			err:            nil,
		},
		{
			name:           "Getting the ID and printing the result with ugly printer works",
			client:         new(WorkingRecordGetter),
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

// Test_discovery_AppendSeedRecord tests the discovery.AppendSeedRecord function
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

// WorkingSeedExecutionGetter mocks a working seed execution getter.
type WorkingSeedExecutionGetter struct {
	mock.Mock
}

// Get returns a seed execution.
func (g *WorkingSeedExecutionGetter) Get(id uuid.UUID) (gjson.Result, error) {
	return gjson.Parse(`{
  "id": "f85a5e19-8ed9-4f8c-9e2e-e1d5484612f3",
  "creationTimestamp": "2025-10-10T19:48:31Z",
  "lastUpdatedTimestamp": "2025-10-10T19:48:31Z",
  "triggerType": "MANUAL",
  "status": "RUNNING",
  "scanType": "FULL",
  "properties": {
    "stagingBucket": "testBucket"
  },
  "stages": ["BEFORE_HOOKS","INGEST"]
}`), nil
}

// GetAll implements the interface.
func (g *WorkingSeedExecutionGetter) GetAll() ([]gjson.Result, error) {
	return []gjson.Result{}, nil
}

// Audit returns real audited changes.
func (g *WorkingSeedExecutionGetter) Audit(id uuid.UUID) ([]gjson.Result, error) {
	return gjson.Parse(`[
	{"timestamp":"2025-09-05T20:09:22.543Z","status":"CREATED","stages":[]},
	{"timestamp":"2025-09-05T20:09:26.621Z","status":"RUNNING","stages":[]},
	{"timestamp":"2025-09-05T20:09:37.592Z","status":"RUNNING","stages":["BEFORE_HOOKS"]},
	{"timestamp":"2025-09-05T20:13:26.602Z","status":"RUNNING","stages":["BEFORE_HOOKS","INGEST"]}
]`).Array(), nil
}

// FailingSeedExecutionGetterGetExecutionFails mocks when getting a seed execution fails.
type FailingSeedExecutionGetterGetExecutionFails struct {
	mock.Mock
}

// Get returns seed execution not found.
func (g *FailingSeedExecutionGetterGetExecutionFails) Get(id uuid.UUID) (gjson.Result, error) {
	return gjson.Result{}, discoveryPackage.Error{Status: http.StatusNotFound, Body: gjson.Parse(`{
  "status": 404,
  "code": 1003,
  "messages": [
    "Seed execution not found: f85a5e19-8ed9-4f8c-9e2e-e1d5484612f2"
  ],
  "timestamp": "2025-11-17T19:32:01.555127800Z"
}`)}
}

// GetAll implements the interface.
func (g *FailingSeedExecutionGetterGetExecutionFails) GetAll() ([]gjson.Result, error) {
	return []gjson.Result{}, nil
}

// Audit implements the interface.
func (g *FailingSeedExecutionGetterGetExecutionFails) Audit(id uuid.UUID) ([]gjson.Result, error) {
	return []gjson.Result{}, discoveryPackage.Error{Status: http.StatusUnauthorized, Body: gjson.Parse(`{"error":"unauthorized"}`)}
}

// FFailingSeedExecutionGetterAuditFails mocks when getting the audit fails.
type FailingSeedExecutionGetterAuditFails struct {
	mock.Mock
}

// Get returns a seed execution.
func (g *FailingSeedExecutionGetterAuditFails) Get(id uuid.UUID) (gjson.Result, error) {
	return gjson.Parse(`{
  "id": "f85a5e19-8ed9-4f8c-9e2e-e1d5484612f3",
  "creationTimestamp": "2025-10-10T19:48:31Z",
  "lastUpdatedTimestamp": "2025-10-10T19:48:31Z",
  "triggerType": "MANUAL",
  "status": "RUNNING",
  "scanType": "FULL",
  "properties": {
    "stagingBucket": "testBucket"
  },
  "stages": ["BEFORE_HOOKS","INGEST"]
}`), nil
}

// GetAll implements the interface.
func (g *FailingSeedExecutionGetterAuditFails) GetAll() ([]gjson.Result, error) {
	return []gjson.Result{}, nil
}

// Audit returns an error.
func (g *FailingSeedExecutionGetterAuditFails) Audit(id uuid.UUID) ([]gjson.Result, error) {
	return []gjson.Result{}, discoveryPackage.Error{Status: http.StatusUnauthorized, Body: gjson.Parse(`{"error":"unauthorized"}`)}
}

// WorkingRecordSummarizer mocks when getting the record summary works.
type WorkingRecordSummarizer struct {
	mock.Mock
}

// Summarize returns a real result.
func (s *WorkingRecordSummarizer) Summarize() (gjson.Result, error) {
	return gjson.Parse(`{"PROCESSING":4,"DONE": 4}`), nil
}

// NoContentRecordSummarizer mocks when the summarize does not return anything.
type NoContentRecordSummarizer struct {
	mock.Mock
}

// NoContentRecordSummarizer returns an empty JSON.
func (s *NoContentRecordSummarizer) Summarize() (gjson.Result, error) {
	return gjson.Parse(``), nil
}

// WorkingJobSummarizer mocks when getting the job summary works.
type WorkingJobSummarizer struct {
	mock.Mock
}

// Summarizer returns real results.
func (s *WorkingJobSummarizer) Summarize() (gjson.Result, error) {
	return gjson.Parse(`{"DONE":5,"RUNNING":3}`), nil
}

// FailingJobSummarizer mocks when getting the job summary fails.
type FailingJobSummarizer struct {
	mock.Mock
}

// Summarize returns an error.
func (s *FailingJobSummarizer) Summarize() (gjson.Result, error) {
	return gjson.Result{}, discoveryPackage.Error{Status: http.StatusNotFound, Body: gjson.Parse(`{
  "status": 404,
  "code": 1003,
  "messages": [
    "Seed execution not found: f85a5e19-8ed9-4f8c-9e2e-e1d5484612f2"
  ],
  "timestamp": "2025-11-17T19:32:01.555127800Z"
}`)}
}

// TestAppendSeedExecutionDetails tests the AppendSeedExecutionDetails() function.
func TestAppendSeedExecutionDetails(t *testing.T) {
	tests := []struct {
		name              string
		client            SeedExecutionGetter
		summarizers       map[string]Summarizer
		expectedExecution gjson.Result
		err               error
	}{
		// Working case
		{
			name:   "Auditing works and summarizers return results",
			client: new(WorkingSeedExecutionGetter),
			summarizers: map[string]Summarizer{
				"records": new(WorkingRecordSummarizer),
				"jobs":    new(WorkingJobSummarizer),
			},
			expectedExecution: gjson.Parse("{\n  \"id\": \"f85a5e19-8ed9-4f8c-9e2e-e1d5484612f3\",\n  \"creationTimestamp\": \"2025-10-10T19:48:31Z\",\n  \"lastUpdatedTimestamp\": \"2025-10-10T19:48:31Z\",\n  \"triggerType\": \"MANUAL\",\n  \"status\": \"RUNNING\",\n  \"scanType\": \"FULL\",\n  \"properties\": {\n    \"stagingBucket\": \"testBucket\"\n  },\n  \"stages\": [\"BEFORE_HOOKS\",\"INGEST\"]\n,\"audit\":[\n{\"timestamp\":\"2025-09-05T20:09:22.543Z\",\"status\":\"CREATED\",\"stages\":[]},\n{\"timestamp\":\"2025-09-05T20:09:26.621Z\",\"status\":\"RUNNING\",\"stages\":[]},\n{\"timestamp\":\"2025-09-05T20:09:37.592Z\",\"status\":\"RUNNING\",\"stages\":[\"BEFORE_HOOKS\"]},\n{\"timestamp\":\"2025-09-05T20:13:26.602Z\",\"status\":\"RUNNING\",\"stages\":[\"BEFORE_HOOKS\",\"INGEST\"]}\n],\"records\":{\"PROCESSING\":4,\"DONE\": 4},\"jobs\":{\"DONE\":5,\"RUNNING\":3}}"),
			err:               nil,
		},
		{
			name:   "Auditing works and one summarizer returns no content",
			client: new(WorkingSeedExecutionGetter),
			summarizers: map[string]Summarizer{
				"records": new(NoContentRecordSummarizer),
				"jobs":    new(WorkingJobSummarizer),
			},
			expectedExecution: gjson.Parse("{\n  \"id\": \"f85a5e19-8ed9-4f8c-9e2e-e1d5484612f3\",\n  \"creationTimestamp\": \"2025-10-10T19:48:31Z\",\n  \"lastUpdatedTimestamp\": \"2025-10-10T19:48:31Z\",\n  \"triggerType\": \"MANUAL\",\n  \"status\": \"RUNNING\",\n  \"scanType\": \"FULL\",\n  \"properties\": {\n    \"stagingBucket\": \"testBucket\"\n  },\n  \"stages\": [\"BEFORE_HOOKS\",\"INGEST\"]\n,\"audit\":[\n{\"timestamp\":\"2025-09-05T20:09:22.543Z\",\"status\":\"CREATED\",\"stages\":[]},\n{\"timestamp\":\"2025-09-05T20:09:26.621Z\",\"status\":\"RUNNING\",\"stages\":[]},\n{\"timestamp\":\"2025-09-05T20:09:37.592Z\",\"status\":\"RUNNING\",\"stages\":[\"BEFORE_HOOKS\"]},\n{\"timestamp\":\"2025-09-05T20:13:26.602Z\",\"status\":\"RUNNING\",\"stages\":[\"BEFORE_HOOKS\",\"INGEST\"]}\n],\"records\":{},\"jobs\":{\"DONE\":5,\"RUNNING\":3}}"),
			err:               nil,
		},
		// Error case
		{
			name:              "Getting the audit fails",
			client:            new(FailingSeedExecutionGetterAuditFails),
			expectedExecution: gjson.Result{},
			summarizers: map[string]Summarizer{
				"records": new(NoContentRecordSummarizer),
				"jobs":    new(WorkingJobSummarizer),
			},
			err: discoveryPackage.Error{Status: http.StatusUnauthorized, Body: gjson.Parse(`{"error":"unauthorized"}`)},
		},
		{
			name:   "Auditing works but a summarizer fails",
			client: new(WorkingSeedExecutionGetter),
			summarizers: map[string]Summarizer{
				"records": new(WorkingRecordSummarizer),
				"jobs":    new(FailingJobSummarizer),
			},
			expectedExecution: gjson.Result{},
			err: discoveryPackage.Error{Status: http.StatusNotFound, Body: gjson.Parse(`{
  "status": 404,
  "code": 1003,
  "messages": [
    "Seed execution not found: f85a5e19-8ed9-4f8c-9e2e-e1d5484612f2"
  ],
  "timestamp": "2025-11-17T19:32:01.555127800Z"
}`)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			seedExecution := gjson.Parse(`{
  "id": "f85a5e19-8ed9-4f8c-9e2e-e1d5484612f3",
  "creationTimestamp": "2025-10-10T19:48:31Z",
  "lastUpdatedTimestamp": "2025-10-10T19:48:31Z",
  "triggerType": "MANUAL",
  "status": "RUNNING",
  "scanType": "FULL",
  "properties": {
    "stagingBucket": "testBucket"
  },
  "stages": ["BEFORE_HOOKS","INGEST"]
}`)
			executionId, err := uuid.Parse("f85a5e19-8ed9-4f8c-9e2e-e1d5484612f3")
			require.NoError(t, err)
			result, err := AppendSeedExecutionDetails(seedExecution, executionId, tc.client, tc.summarizers)
			assert.Equal(t, tc.expectedExecution, result)
			if tc.err != nil {
				require.Error(t, err)
				assert.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestGetSeedExecution tests the GetSeedExecution() function.
func TestGetSeedExecution(t *testing.T) {
	tests := []struct {
		name           string
		client         SeedExecutionGetter
		summarizers    map[string]Summarizer
		details        bool
		expectedOutput string
		printer        Printer
		outWriter      io.Writer
		err            error
	}{
		// Working case
		{
			name:   "GetSeedExecution returns a working seed execution, appends the details, and correctly prints the result with the given printer",
			client: new(WorkingSeedExecutionGetter),
			summarizers: map[string]Summarizer{
				"records": new(WorkingRecordSummarizer),
				"jobs":    new(WorkingJobSummarizer),
			},
			details:        true,
			expectedOutput: "{\"audit\":[{\"stages\":[],\"status\":\"CREATED\",\"timestamp\":\"2025-09-05T20:09:22.543Z\"},{\"stages\":[],\"status\":\"RUNNING\",\"timestamp\":\"2025-09-05T20:09:26.621Z\"},{\"stages\":[\"BEFORE_HOOKS\"],\"status\":\"RUNNING\",\"timestamp\":\"2025-09-05T20:09:37.592Z\"},{\"stages\":[\"BEFORE_HOOKS\",\"INGEST\"],\"status\":\"RUNNING\",\"timestamp\":\"2025-09-05T20:13:26.602Z\"}],\"creationTimestamp\":\"2025-10-10T19:48:31Z\",\"id\":\"f85a5e19-8ed9-4f8c-9e2e-e1d5484612f3\",\"jobs\":{\"DONE\":5,\"RUNNING\":3},\"lastUpdatedTimestamp\":\"2025-10-10T19:48:31Z\",\"properties\":{\"stagingBucket\":\"testBucket\"},\"records\":{\"DONE\":4,\"PROCESSING\":4},\"scanType\":\"FULL\",\"stages\":[\"BEFORE_HOOKS\",\"INGEST\"],\"status\":\"RUNNING\",\"triggerType\":\"MANUAL\"}\n",
			printer:        JsonObjectPrinter(false),
			err:            nil,
		},
		{
			name:   "GetSeedExecution prints a seed execution with no details with the pretty printer",
			client: new(WorkingSeedExecutionGetter),
			summarizers: map[string]Summarizer{
				"records": new(WorkingRecordSummarizer),
				"jobs":    new(WorkingJobSummarizer),
			},
			details:        false,
			expectedOutput: "{\n  \"creationTimestamp\": \"2025-10-10T19:48:31Z\",\n  \"id\": \"f85a5e19-8ed9-4f8c-9e2e-e1d5484612f3\",\n  \"lastUpdatedTimestamp\": \"2025-10-10T19:48:31Z\",\n  \"properties\": {\n    \"stagingBucket\": \"testBucket\"\n  },\n  \"scanType\": \"FULL\",\n  \"stages\": [\n    \"BEFORE_HOOKS\",\n    \"INGEST\"\n  ],\n  \"status\": \"RUNNING\",\n  \"triggerType\": \"MANUAL\"\n}\n",
			printer:        nil,
			err:            nil,
		},
		// Error case
		{
			name:   "Getting the seed execution fails",
			client: new(FailingSeedExecutionGetterGetExecutionFails),
			summarizers: map[string]Summarizer{
				"records": new(WorkingRecordSummarizer),
				"jobs":    new(WorkingJobSummarizer),
			},
			details: true,
			err: NewErrorWithCause(ErrorExitCode, discoveryPackage.Error{Status: http.StatusNotFound, Body: gjson.Parse(`{
  "status": 404,
  "code": 1003,
  "messages": [
    "Seed execution not found: f85a5e19-8ed9-4f8c-9e2e-e1d5484612f2"
  ],
  "timestamp": "2025-11-17T19:32:01.555127800Z"
}`)}, "Could not get seed execution with id \"f85a5e19-8ed9-4f8c-9e2e-e1d5484612f3\""),
		},
		{
			name:   "Getting the seed execution audited changes fails",
			client: new(FailingSeedExecutionGetterAuditFails),
			summarizers: map[string]Summarizer{
				"records": new(WorkingRecordSummarizer),
				"jobs":    new(WorkingJobSummarizer),
			},
			details: true,
			err:     NewErrorWithCause(ErrorExitCode, discoveryPackage.Error{Status: http.StatusUnauthorized, Body: gjson.Parse(`{"error":"unauthorized"}`)}, "Could not get details for seed execution with id \"f85a5e19-8ed9-4f8c-9e2e-e1d5484612f3\""),
		},
		{
			name:   "Printing fails",
			client: new(WorkingSeedExecutionGetter),
			summarizers: map[string]Summarizer{
				"records": new(WorkingRecordSummarizer),
				"jobs":    new(WorkingJobSummarizer),
			},
			details:        true,
			expectedOutput: "{\n  \"id\": \"f85a5e19-8ed9-4f8c-9e2e-e1d5484612f3\",\n  \"creationTimestamp\": \"2025-10-10T19:48:31Z\",\n  \"lastUpdatedTimestamp\": \"2025-10-10T19:48:31Z\",\n  \"triggerType\": \"MANUAL\",\n  \"status\": \"RUNNING\",\n  \"scanType\": \"FULL\",\n  \"properties\": {\n    \"stagingBucket\": \"testBucket\"\n  },\n  \"stages\": [\"BEFORE_HOOKS\",\"INGEST\"]\n,\"audit\":[\n{\"timestamp\":\"2025-09-05T20:09:22.543Z\",\"status\":\"CREATED\",\"stages\":[]},\n{\"timestamp\":\"2025-09-05T20:09:26.621Z\",\"status\":\"RUNNING\",\"stages\":[]},\n{\"timestamp\":\"2025-09-05T20:09:37.592Z\",\"status\":\"RUNNING\",\"stages\":[\"BEFORE_HOOKS\"]},\n{\"timestamp\":\"2025-09-05T20:13:26.602Z\",\"status\":\"RUNNING\",\"stages\":[\"BEFORE_HOOKS\",\"INGEST\"]}\n],\"records\":{\"PROCESSING\":4,\"DONE\": 4},\"jobs\":{\"DONE\":5,\"RUNNING\":3}}",
			printer:        JsonObjectPrinter(false),
			outWriter:      testutils.ErrWriter{Err: errors.New("write failed")},
			err:            NewErrorWithCause(ErrorExitCode, errors.New("write failed"), "Could not print JSON object"),
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
			executionId, err := uuid.Parse("f85a5e19-8ed9-4f8c-9e2e-e1d5484612f3")
			require.NoError(t, err)

			err = d.GetSeedExecution(tc.client, executionId, tc.summarizers, tc.details, tc.printer)

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
