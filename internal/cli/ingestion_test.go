package cli

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"os"
	"testing"

	discoveryPackage "github.com/pureinsights/pdp-cli/discovery"
	"github.com/pureinsights/pdp-cli/internal/iostreams"
	"github.com/pureinsights/pdp-cli/internal/testutils"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

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
