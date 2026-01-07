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
	"github.com/pureinsights/discovery-cli/internal/testutils/mocks"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// TestGetEntityId tests the GetEntityId() function.
func TestGetEntityId(t *testing.T) {
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
			name:     "GetEntityId works",
			client:   new(mocks.WorkingSearcher),
			expected: successId,
			err:      successErr,
		},

		// Error case
		{
			name:     "Cannot convert to UUID fails",
			client:   new(mocks.SearcherIDNotUUID),
			expected: errorId,
			err:      errorErr,
		},
		{
			name:     "Search fails",
			client:   new(mocks.FailingSearcher),
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
			seedId, err := GetEntityId(d, tc.client, "seed")

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

// Test_discovery_GetEntity tests the discovery.GetEntity() function.
func Test_discovery_GetEntity(t *testing.T) {
	tests := []struct {
		name           string
		client         Getter
		printer        Printer
		expectedOutput string
		outWriter      io.Writer
		err            error
	}{
		// Working case
		{
			name:           "GetEntity correctly prints an object with the pretty printer",
			client:         new(mocks.WorkingGetter),
			printer:        nil,
			expectedOutput: "{\n  \"active\": true,\n  \"creationTimestamp\": \"2025-08-14T18:02:38Z\",\n  \"id\": \"5f125024-1e5e-4591-9fee-365dc20eeeed\",\n  \"labels\": [],\n  \"lastUpdatedTimestamp\": \"2025-08-18T20:55:43Z\",\n  \"name\": \"MongoDB text processor\",\n  \"type\": \"mongo\"\n}\n",
			err:            nil,
		},
		{
			name:           "GetEntity correctly prints an object with JSON ugly printer",
			client:         new(mocks.WorkingGetter),
			printer:        JsonObjectPrinter(false),
			expectedOutput: "{\"active\":true,\"creationTimestamp\":\"2025-08-14T18:02:38Z\",\"id\":\"5f125024-1e5e-4591-9fee-365dc20eeeed\",\"labels\":[],\"lastUpdatedTimestamp\":\"2025-08-18T20:55:43Z\",\"name\":\"MongoDB text processor\",\"type\":\"mongo\"}\n",
			err:            nil,
		},

		// Error case
		{
			name:           "Get returns 404 Not Found",
			client:         new(mocks.FailingGetter),
			printer:        nil,
			expectedOutput: "",
			err: NewErrorWithCause(ErrorExitCode, discoveryPackage.Error{
				Status: http.StatusNotFound,
				Body: gjson.Parse(`{
			"status": 404,
			"code": 1003,
			"messages": [
				"Secret not found: 5f125024-1e5e-4591-9fee-365dc20eeeed"
			],
			"timestamp": "2025-10-16T00:15:31.888410500Z"
		}`),
			}, "Could not get entity with id \"5f125024-1e5e-4591-9fee-365dc20eeeed\""),
		},
		{
			name:      "Printing fails",
			client:    new(mocks.WorkingGetter),
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

			id, err := uuid.Parse("5f125024-1e5e-4591-9fee-365dc20eeeed")
			require.NoError(t, err)
			d := NewDiscovery(&ios, viper.New(), "")
			err = d.GetEntity(tc.client, id, tc.printer)

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

// GetEntities obtains all of the entities using the given client and then prints out the result using the received printer or the JSON array printer.
func Test_discovery_GetEntities(t *testing.T) {
	tests := []struct {
		name           string
		client         Getter
		printer        Printer
		expectedOutput string
		outWriter      io.Writer
		err            error
	}{
		// Working case
		{
			name:           "GetEntities correctly prints an array with the pretty printer",
			client:         new(mocks.WorkingGetter),
			printer:        JsonArrayPrinter(true),
			expectedOutput: "[\n{\n  \"active\": true,\n  \"creationTimestamp\": \"2025-08-21T17:57:16Z\",\n  \"id\": \"3393f6d9-94c1-4b70-ba02-5f582727d998\",\n  \"labels\": [],\n  \"lastUpdatedTimestamp\": \"2025-08-21T17:57:16Z\",\n  \"name\": \"MongoDB text processor 4\",\n  \"type\": \"mongo\"\n},\n{\n  \"active\": true,\n  \"creationTimestamp\": \"2025-08-14T18:02:38Z\",\n  \"id\": \"5f125024-1e5e-4591-9fee-365dc20eeeed\",\n  \"labels\": [],\n  \"lastUpdatedTimestamp\": \"2025-08-18T20:55:43Z\",\n  \"name\": \"MongoDB text processor\",\n  \"type\": \"mongo\"\n},\n{\n  \"active\": true,\n  \"creationTimestamp\": \"2025-08-14T18:02:38Z\",\n  \"id\": \"86e7f920-a4e4-4b64-be84-5437a7673db8\",\n  \"labels\": [],\n  \"lastUpdatedTimestamp\": \"2025-08-14T18:02:38Z\",\n  \"name\": \"Script processor\",\n  \"type\": \"script\"\n}\n]\n",
			err:            nil,
		},
		{
			name:           "GetEntities correctly prints an array with JSON ugly printer",
			client:         new(mocks.WorkingGetter),
			printer:        nil,
			expectedOutput: "{\"active\":true,\"creationTimestamp\":\"2025-08-21T17:57:16Z\",\"id\":\"3393f6d9-94c1-4b70-ba02-5f582727d998\",\"labels\":[],\"lastUpdatedTimestamp\":\"2025-08-21T17:57:16Z\",\"name\":\"MongoDB text processor 4\",\"type\":\"mongo\"}\n{\"active\":true,\"creationTimestamp\":\"2025-08-14T18:02:38Z\",\"id\":\"5f125024-1e5e-4591-9fee-365dc20eeeed\",\"labels\":[],\"lastUpdatedTimestamp\":\"2025-08-18T20:55:43Z\",\"name\":\"MongoDB text processor\",\"type\":\"mongo\"}\n{\"active\":true,\"creationTimestamp\":\"2025-08-14T18:02:38Z\",\"id\":\"86e7f920-a4e4-4b64-be84-5437a7673db8\",\"labels\":[],\"lastUpdatedTimestamp\":\"2025-08-14T18:02:38Z\",\"name\":\"Script processor\",\"type\":\"script\"}\n",
			err:            nil,
		},

		// Error case
		{
			name:           "GetAll returns 401 Unauthorized",
			client:         new(mocks.FailingGetter),
			printer:        nil,
			expectedOutput: "",
			err:            NewErrorWithCause(ErrorExitCode, discoveryPackage.Error{Status: http.StatusUnauthorized, Body: gjson.Parse(`{"error":"unauthorized"}`)}, "Could not get all entities"),
		},
		{
			name:      "Printing fails",
			client:    new(mocks.WorkingGetter),
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
			err := d.GetEntities(tc.client, tc.printer)

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

// Test_searchEntity tests the discovery.searchEntity() function.
func Test_searchEntity(t *testing.T) {
	tests := []struct {
		name      string
		client    Searcher
		id        string
		expected  gjson.Result
		outWriter io.Writer
		err       error
	}{
		{
			name:   "Search by name works for name",
			client: new(mocks.WorkingSearcher),
			id:     "MongoDB Atlas server",
			expected: gjson.Parse(`{
		"type": "mongo",
		"name": "MongoDB Atlas server",
		"labels": [],
		"active": true,
		"id": "986ce864-af76-4fcb-8b4f-f4e4c6ab0951",
		"creationTimestamp": "2025-09-29T15:50:17Z",
		"lastUpdatedTimestamp": "2025-09-29T15:50:17Z",
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
	}`),
			err: nil,
		},
		{
			name:   "Search by name fails for name, but get works for Id",
			client: new(mocks.FailingSearcherWorkingGetter),
			id:     "986ce864-af76-4fcb-8b4f-f4e4c6ab0951",
			expected: gjson.Parse(`{
		"type": "mongo",
		"name": "MongoDB Atlas server clone",
		"labels": [],
		"active": true,
		"id": "986ce864-af76-4fcb-8b4f-f4e4c6ab0951",
		"creationTimestamp": "2025-09-29T15:50:17Z",
		"lastUpdatedTimestamp": "2025-09-29T15:50:17Z",
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
	}`),
			err: nil,
		},

		// Error cases
		{
			name:     "Search by name fails with error Bad Request",
			client:   new(mocks.FailingSearcher),
			id:       "MongoDB Atlas Server",
			expected: gjson.Result{},
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
		{
			name:     "Search by name fails with error Not Found, and Get fails with 404 Not found",
			client:   new(mocks.FailingSearcherFailingGetter),
			id:       "986ce864-af76-4fcb-8b4f-f4e4c6ab0951",
			expected: gjson.Result{},
			err: discoveryPackage.Error{
				Status: http.StatusNotFound,
				Body: gjson.Parse(`{
			"status": 404,
			"code": 1003,
			"messages": [
				"Server not found: 986ce864-af76-4fcb-8b4f-f4e4c6ab0951"
			],
			"timestamp": "2025-10-16T00:15:31.888410500Z"
		}`),
			},
		},
		{
			name:     "Search by name fails with error 404 Not Found and id is not a UUID",
			client:   new(mocks.FailingSearcherWorkingGetter),
			id:       "MongoDB Atlas Server",
			expected: gjson.Result{},
			err: discoveryPackage.Error{
				Status: http.StatusNotFound,
				Body:   gjson.Parse("{\n\t\"status\": 404,\n\t\"code\": 1003,\n\t\"messages\": [\n\t\t\"Entity not found: entity with name \"MongoDB Atlas Server\" does not exist\"\n\t]\n}"),
			},
		},
		{
			name:     "Search returns an error different from discovery.Error",
			client:   new(mocks.SearcherReturnsOtherError),
			id:       "MongoDB Atlas Server",
			expected: gjson.Result{},
			err:      errors.New("not discovery error"),
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
			result, err := d.searchEntity(tc.client, tc.id)
			assert.Equal(t, tc.expected, result)
			if tc.err != nil {
				require.Error(t, err)
				assert.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestSearchEntity tests the cli.SearchEntity() function.
func TestSearchEntity(t *testing.T) {
	tests := []struct {
		name      string
		client    Searcher
		id        string
		expected  gjson.Result
		outWriter io.Writer
		err       error
	}{
		{
			name:   "Search by name works for name",
			client: new(mocks.WorkingSearcher),
			id:     "MongoDB Atlas server",
			expected: gjson.Parse(`{
		"type": "mongo",
		"name": "MongoDB Atlas server",
		"labels": [],
		"active": true,
		"id": "986ce864-af76-4fcb-8b4f-f4e4c6ab0951",
		"creationTimestamp": "2025-09-29T15:50:17Z",
		"lastUpdatedTimestamp": "2025-09-29T15:50:17Z",
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
	}`),
			err: nil,
		},

		// Error cases
		{
			name:     "Search by name fails with error Not Found, and Get fails with 404 Not found",
			client:   new(mocks.FailingSearcherFailingGetter),
			id:       "986ce864-af76-4fcb-8b4f-f4e4c6ab0951",
			expected: gjson.Result{},
			err: discoveryPackage.Error{
				Status: http.StatusNotFound,
				Body: gjson.Parse(`{
			"status": 404,
			"code": 1003,
			"messages": [
				"Server not found: 986ce864-af76-4fcb-8b4f-f4e4c6ab0951"
			],
			"timestamp": "2025-10-16T00:15:31.888410500Z"
		}`),
			},
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
			result, err := SearchEntity(d, tc.client, tc.id)
			assert.Equal(t, tc.expected, result)
			if tc.err != nil {
				require.Error(t, err)
				assert.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Test_discovery_SearchEntity tests the discovery.SearchEntity() function.
func Test_discovery_SearchEntity(t *testing.T) {
	tests := []struct {
		name           string
		client         Searcher
		id             string
		expectedOutput string
		printer        Printer
		outWriter      io.Writer
		err            error
	}{
		{
			name:           "SearchEntity correctly prints an object with the pretty printer",
			client:         new(mocks.WorkingSearcher),
			id:             "MongoDB Atlas Server",
			printer:        nil,
			expectedOutput: "{\n  \"active\": true,\n  \"config\": {\n    \"connection\": {\n      \"connectTimeout\": \"1m\",\n      \"readTimeout\": \"30s\"\n    },\n    \"credentialId\": \"9ababe08-0b74-4672-bb7c-e7a8227d6d4c\",\n    \"servers\": [\n      \"mongodb+srv://cluster0.dleud.mongodb.net/\"\n    ]\n  },\n  \"creationTimestamp\": \"2025-09-29T15:50:17Z\",\n  \"id\": \"986ce864-af76-4fcb-8b4f-f4e4c6ab0951\",\n  \"labels\": [],\n  \"lastUpdatedTimestamp\": \"2025-09-29T15:50:17Z\",\n  \"name\": \"MongoDB Atlas server\",\n  \"type\": \"mongo\"\n}\n",
			err:            nil,
		},
		{
			name:           "SearchEntity correctly prints an object with JSON ugly printer",
			client:         new(mocks.FailingSearcherWorkingGetter),
			id:             "986ce864-af76-4fcb-8b4f-f4e4c6ab0951",
			printer:        JsonObjectPrinter(false),
			expectedOutput: "{\"active\":true,\"config\":{\"connection\":{\"connectTimeout\":\"1m\",\"readTimeout\":\"30s\"},\"credentialId\":\"9ababe08-0b74-4672-bb7c-e7a8227d6d4c\",\"servers\":[\"mongodb+srv://cluster0.dleud.mongodb.net/\"]},\"creationTimestamp\":\"2025-09-29T15:50:17Z\",\"id\":\"986ce864-af76-4fcb-8b4f-f4e4c6ab0951\",\"labels\":[],\"lastUpdatedTimestamp\":\"2025-09-29T15:50:17Z\",\"name\":\"MongoDB Atlas server clone\",\"type\":\"mongo\"}\n",
			err:            nil,
		},

		// Error case
		{
			name:           "Search returns 404 Not Found",
			client:         new(mocks.FailingSearcherFailingGetter),
			id:             "MongoDB Atlas Server",
			printer:        nil,
			expectedOutput: "",
			err: NewErrorWithCause(ErrorExitCode, discoveryPackage.Error{
				Status: http.StatusNotFound,
				Body: gjson.Parse(`{
	"status": 404,
	"code": 1003,
	"messages": [
		"Entity not found: entity with name "MongoDB Atlas Server" does not exist"
	]
}`),
			}, "Could not search for entity with id \"MongoDB Atlas Server\""),
		},
		{
			name:           "Printing fails",
			client:         new(mocks.WorkingSearcher),
			id:             "MongoDB Atlas Server",
			expectedOutput: "",
			printer:        nil,
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
			err := d.SearchEntity(tc.client, tc.id, tc.printer)

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

// Test_discovery_SearchEntities tests the discovery.SearchEntities() function.
func Test_discovery_SearchEntities(t *testing.T) {
	tests := []struct {
		name           string
		client         Searcher
		printer        Printer
		expectedOutput string
		outWriter      io.Writer
		err            error
	}{
		// Working case
		{
			name:           "SearchEntities correctly prints an array with the sent printer",
			client:         new(mocks.WorkingSearcher),
			printer:        JsonArrayPrinter(true),
			expectedOutput: "[\n{\n  \"highlight\": {},\n  \"score\": 0.20970252,\n  \"source\": {\n    \"active\": true,\n    \"creationTimestamp\": \"2025-09-29T15:50:17Z\",\n    \"id\": \"986ce864-af76-4fcb-8b4f-f4e4c6ab0951\",\n    \"labels\": [],\n    \"lastUpdatedTimestamp\": \"2025-09-29T15:50:17Z\",\n    \"name\": \"MongoDB Atlas server clone\",\n    \"type\": \"mongo\"\n  }\n},\n{\n  \"highlight\": {},\n  \"score\": 0.20970252,\n  \"source\": {\n    \"active\": true,\n    \"creationTimestamp\": \"2025-09-29T15:50:19Z\",\n    \"id\": \"8f14c11c-bb66-49d3-aa2a-dedff4608c17\",\n    \"labels\": [],\n    \"lastUpdatedTimestamp\": \"2025-09-29T15:50:19Z\",\n    \"name\": \"MongoDB Atlas server clone 1\",\n    \"type\": \"mongo\"\n  }\n},\n{\n  \"highlight\": {},\n  \"score\": 0.20970252,\n  \"source\": {\n    \"active\": true,\n    \"creationTimestamp\": \"2025-09-29T15:50:20Z\",\n    \"id\": \"3a0214a4-72cc-4eee-ad0c-9e3af9b08a6c\",\n    \"labels\": [],\n    \"lastUpdatedTimestamp\": \"2025-09-29T15:50:20Z\",\n    \"name\": \"MongoDB Atlas server clone 3\",\n    \"type\": \"mongo\"\n  }\n}\n]\n",
			err:            nil,
		},
		{
			name:           "SearchEntities correctly prints an array with JSON ugly printer",
			client:         new(mocks.WorkingSearcher),
			printer:        nil,
			expectedOutput: "{\"highlight\":{},\"score\":0.20970252,\"source\":{\"active\":true,\"creationTimestamp\":\"2025-09-29T15:50:17Z\",\"id\":\"986ce864-af76-4fcb-8b4f-f4e4c6ab0951\",\"labels\":[],\"lastUpdatedTimestamp\":\"2025-09-29T15:50:17Z\",\"name\":\"MongoDB Atlas server clone\",\"type\":\"mongo\"}}\n{\"highlight\":{},\"score\":0.20970252,\"source\":{\"active\":true,\"creationTimestamp\":\"2025-09-29T15:50:19Z\",\"id\":\"8f14c11c-bb66-49d3-aa2a-dedff4608c17\",\"labels\":[],\"lastUpdatedTimestamp\":\"2025-09-29T15:50:19Z\",\"name\":\"MongoDB Atlas server clone 1\",\"type\":\"mongo\"}}\n{\"highlight\":{},\"score\":0.20970252,\"source\":{\"active\":true,\"creationTimestamp\":\"2025-09-29T15:50:20Z\",\"id\":\"3a0214a4-72cc-4eee-ad0c-9e3af9b08a6c\",\"labels\":[],\"lastUpdatedTimestamp\":\"2025-09-29T15:50:20Z\",\"name\":\"MongoDB Atlas server clone 3\",\"type\":\"mongo\"}}\n",
			err:            nil,
		},

		// Error case
		{
			name:           "SearchAll returns 401 Unauthorized",
			client:         new(mocks.FailingSearcherFailingGetter),
			printer:        nil,
			expectedOutput: "",
			err:            NewErrorWithCause(ErrorExitCode, discoveryPackage.Error{Status: http.StatusUnauthorized, Body: gjson.Parse(`{"error":"unauthorized"}`)}, "Could not search for the entities"),
		},
		{
			name:      "Printing fails",
			client:    new(mocks.WorkingSearcher),
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
			err := d.SearchEntities(tc.client, gjson.Result{}, tc.printer)

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

// Test_parseFilter tests the parseFilter() function.
func Test_parseFilter(t *testing.T) {
	tests := []struct {
		name               string
		filter             string
		expectedFilterType string
		expectedFilters    []string
		err                error
	}{
		{
			name:               "Send label with key and value",
			filter:             "label=A:C",
			expectedFilterType: "label",
			expectedFilters: []string{`{
	"equals": {
		"field": "labels.key",
		"value": "A"
		}
	}`, `{
	"equals": {
		"field": "labels.value",
		"value": "C"
		}
	}`},
			err: nil,
		},
		{
			name:               "Send label with only key",
			filter:             "label=B",
			expectedFilterType: "label",
			expectedFilters: []string{`{
	"equals": {
		"field": "labels.key",
		"value": "B"
		}
	}`},
			err: nil,
		},
		{
			name:               "Send type filter",
			filter:             "type=mongo",
			expectedFilterType: "type",
			expectedFilters: []string{`{
	"equals": {
		"field": "type",
		"value": "mongo"
		}
	}`,
			},
			err: nil,
		},
		{
			name:   "Send unknown filter",
			filter: "name=mongo",
			err:    NewError(ErrorExitCode, "Filter type \"name\" does not exist"),
		},
		{
			name:   "Send filter with no =",
			filter: "label",
			err:    NewError(ErrorExitCode, "Filter \"label\" does not follow the format {type}={key}[:{value}]"),
		},
		{
			name:   "Send label filter with empty key",
			filter: "label=",
			err:    NewError(ErrorExitCode, "The label's key in the filter \"label=\" cannot be empty"),
		},
		{
			name:   "Send label filter with empty value",
			filter: "label=key:",
			err:    NewError(ErrorExitCode, "The label's value in the filter \"label=key:\" cannot be empty if ':' is included"),
		},
		{
			name:   "Send type filter with empty type",
			filter: "type=",
			err:    NewError(ErrorExitCode, "The value in the type filter \"type=\" cannot be empty"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			filterType, filters, err := parseFilter(tc.filter)

			if tc.err != nil {
				assert.Empty(t, filterType)
				assert.Empty(t, filters)
				require.Error(t, err)
				var errStruct Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedFilterType, filterType)
				assert.Equal(t, tc.expectedFilters, filters)
			}
		})
	}
}

// Test_getAndFilterString tests the getAndFilterString() function.
func Test_getAndFilterString(t *testing.T) {
	tests := []struct {
		name                 string
		filters              []string
		expectedFilterString string
		err                  error
	}{
		{
			name: "Send two filters",
			filters: []string{`{
	"equals": {
		"field": "labels.key",
		"value": "A"
		}
	}`, `{
	"equals": {
		"field": "labels.value",
		"value": "C"
		}
	}`},
			expectedFilterString: "{\"and\":[{\n\t\"equals\": {\n\t\t\"field\": \"labels.key\",\n\t\t\"value\": \"A\"\n\t\t}\n\t},{\n\t\"equals\": {\n\t\t\"field\": \"labels.value\",\n\t\t\"value\": \"C\"\n\t\t}\n\t}]}",
			err:                  nil,
		},
		{
			name: "Send one filter",
			filters: []string{`{
	"equals": {
		"field": "labels.key",
		"value": "A"
		}
	}`},
			expectedFilterString: `{
	"equals": {
		"field": "labels.key",
		"value": "A"
		}
	}`,
			err: nil,
		},
		{
			name:                 "Send no filters",
			filters:              []string{},
			expectedFilterString: `{}`,
			err:                  nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			filterString, err := getAndFilterString(tc.filters)

			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedFilterString, filterString)
			}
		})
	}
}

// TestBuildEntitiesFilter tests the BuildEntitiesFilter() function.
func TestBuildEntitiesFilter(t *testing.T) {
	tests := []struct {
		name           string
		filters        []string
		expectedFilter string
		err            error
	}{
		// Working cases
		{
			name:    "1 label key/value",
			filters: []string{"label=A:B"},
			expectedFilter: `{
  "and": [
    {
      "equals": {
        "field": "labels.key",
        "value": "A"
      }
    },
    {
      "equals": {
        "field": "labels.value",
        "value": "B"
      }
    }
  ]
}
`,
			err: nil,
		},
		{
			name:    "1 label only key",
			filters: []string{"label=A"},
			expectedFilter: `{
  "equals": {
    "field": "labels.key",
    "value": "A"
  }
}
`,
			err: nil,
		},
		{
			name:    "1 type",
			filters: []string{"type=mongo"},
			expectedFilter: `{
  "equals": {
    "field": "type",
    "value": "mongo"
  }
}
`,
			err: nil,
		},
		{

			name:    "1 label key/value and 1 type",
			filters: []string{"label=A:B", "type=mongo"},
			expectedFilter: `{
  "and": [
    {
      "and": [
        {
          "equals": {
            "field": "labels.key",
            "value": "A"
          }
        },
        {
          "equals": {
            "field": "labels.value",
            "value": "B"
          }
        }
      ]
    },
    {
      "equals": {
        "field": "type",
        "value": "mongo"
      }
    }
  ]
}
`,
			err: nil,
		},
		{

			name:    "2 label key/value, only key, and 1 type",
			filters: []string{"label=A:B", "type=mongo", "label=C"},
			expectedFilter: `{
  "and": [
    {
      "and": [
        {
          "equals": {
            "field": "labels.key",
            "value": "A"
          }
        },
        {
          "equals": {
            "field": "labels.value",
            "value": "B"
          }
        },
        {
          "equals": {
            "field": "labels.key",
            "value": "C"
          }
        }
      ]
    },
    {
      "equals": {
        "field": "type",
        "value": "mongo"
      }
    }
  ]
}
`,
			err: nil,
		},
		{

			name:    "1 label only key, and 2 type",
			filters: []string{"label=A", "type=mongo", "type=openai"},
			expectedFilter: `{
  "and": [
    {
      "equals": {
        "field": "labels.key",
        "value": "A"
      }
    },
    {
      "and": [
        {
          "equals": {
            "field": "type",
            "value": "mongo"
          }
        },
        {
          "equals": {
            "field": "type",
            "value": "openai"
          }
        }
      ]
    }
  ]
}
`,
			err: nil,
		},
		{

			name:    "2 label key/value, only key, and 2 type",
			filters: []string{"label=A:B", "type=mongo", "label=C", "type=openai"},
			expectedFilter: `{
  "and": [
    {
      "and": [
        {
          "equals": {
            "field": "labels.key",
            "value": "A"
          }
        },
        {
          "equals": {
            "field": "labels.value",
            "value": "B"
          }
        },
        {
          "equals": {
            "field": "labels.key",
            "value": "C"
          }
        }
      ]
    },
    {
      "and": [
        {
          "equals": {
            "field": "type",
            "value": "mongo"
          }
        },
        {
          "equals": {
            "field": "type",
            "value": "openai"
          }
        }
      ]
    }
  ]
}
`,
			err: nil,
		},
		{

			name:    "2 label key/value, and 2 type",
			filters: []string{"label=A:B", "type=mongo", "label=C:D", "type=openai"},
			expectedFilter: `{
  "and": [
    {
      "and": [
        {
          "equals": {
            "field": "labels.key",
            "value": "A"
          }
        },
        {
          "equals": {
            "field": "labels.value",
            "value": "B"
          }
        },
        {
          "equals": {
            "field": "labels.key",
            "value": "C"
          }
        },
        {
          "equals": {
            "field": "labels.value",
            "value": "D"
          }
        }
      ]
    },
    {
      "and": [
        {
          "equals": {
            "field": "type",
            "value": "mongo"
          }
        },
        {
          "equals": {
            "field": "type",
            "value": "openai"
          }
        }
      ]
    }
  ]
}
`,
			err: nil,
		},

		// Error case
		{

			name:           "Filter that does not exist",
			filters:        []string{"name=mongo"},
			expectedFilter: ``,
			err:            NewError(ErrorExitCode, "Filter type \"name\" does not exist"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out := &bytes.Buffer{}

			ios := iostreams.IOStreams{
				In:  os.Stdin,
				Out: out,
				Err: os.Stderr,
			}

			filter, err := BuildEntitiesFilter(tc.filters)

			if tc.err != nil {
				assert.Equal(t, gjson.Result{}, filter)
				require.Error(t, err)
				var errStruct Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				printer := JsonObjectPrinter(true)
				err = printer(ios, filter)
				require.NoError(t, err)
				require.Equal(t, tc.expectedFilter, out.String())
			}
		})
	}
}

// Test_discovery_UpsertEntity tests the discovery.UpsertEntity() function.
func Test_discovery_UpsertEntity(t *testing.T) {
	tests := []struct {
		name     string
		client   Creator
		config   gjson.Result
		expected gjson.Result
		err      error
	}{
		// Working case
		{
			name:   "UpsertEntity creates an entity",
			client: new(mocks.WorkingCreator),
			config: gjson.Parse(`{
		"type": "mongo",
		"name": "MongoDB credential",
		"labels": [],
		"active": true,
		"creationTimestamp": "2025-08-14T18:02:11Z",
		"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
		"secret": "mongo-secret"
	}`),
			expected: gjson.Parse(`{
		"type": "mongo",
		"name": "MongoDB credential",
		"labels": [],
		"active": true,
		"id": "9ababe08-0b74-4672-bb7c-e7a8227d6d4c",
		"creationTimestamp": "2025-08-14T18:02:11Z",
		"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
		"secret": "mongo-secret"
	}`),
			err: nil,
		},
		{
			name:   "UpsertEntity updates an entity",
			client: new(mocks.WorkingCreator),
			config: gjson.Parse(`{
		"type": "mongo",
		"name": "OpenAI credential",
		"labels": [],
		"active": true,
		"id": "9ababe08-0b74-4672-bb7c-e7a8227d6d4c",
		"creationTimestamp": "2025-08-14T18:02:11Z",
		"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
		"secret": "openai-secret"
	}`),
			expected: gjson.Parse(`{
		"type": "mongo",
		"name": "MongoDB credential",
		"labels": [],
		"active": true,
		"id": "9ababe08-0b74-4672-bb7c-e7a8227d6d4c",
		"creationTimestamp": "2025-08-14T18:02:11Z",
		"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
		"secret": "mongo-secret"
	}`),
			err: nil,
		},

		// Error case
		{
			name:   "UpsertEntity fails to create an entity",
			client: new(mocks.FailingCreator),
			config: gjson.Parse(`{
		"type": "mongo",
		"name": "MongoDB credential",
		"labels": [],
		"active": true,
		"creationTimestamp": "2025-08-14T18:02:11Z",
		"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
		"secret": "mongo-secret"
	}`),
			expected: gjson.Result{},
			err: discoveryPackage.Error{Status: http.StatusBadRequest, Body: gjson.Parse(`{
  "status": 400,
  "code": 3002,
  "messages": [
    "Invalid JSON: Illegal unquoted character ((CTRL-CHAR, code 10)): has to be escaped using backslash to be included in name\n at [Source: REDACTED (StreamReadFeature.INCLUDE_SOURCE_IN_LOCATION disabled); line: 5, column: 17]"
  ],
  "timestamp": "2025-10-29T14:46:48.055840300Z"
}`)},
		},
		{
			name:   "UpsertEntity fails to update an entity",
			client: new(mocks.FailingCreator),
			config: gjson.Parse(`{
		"type": "mongo",
		"name": "OpenAI credential",
		"labels": [],
		"active": true,
		"id": "9ababe08-0b74-4672-bb7c-e7a8227d6d4c",
		"creationTimestamp": "2025-08-14T18:02:11Z",
		"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
		"secret": "openai-secret"
	}`),
			expected: gjson.Result{},
			err: discoveryPackage.Error{Status: http.StatusBadRequest, Body: gjson.Parse(`{
  "status": 404,
  "code": 1003,
  "messages": [
    "Entity not found: 9ababe08-0b74-4672-bb7c-e7a8227d6d4d"
  ],
  "timestamp": "2025-10-29T14:47:36.290329Z"
}`)},
		},
		{
			name:   "UpsertEntity receives an ID that is not a UUID",
			client: new(mocks.WorkingCreator),
			config: gjson.Parse(`{
		"type": "mongo",
		"name": "OpenAI credential",
		"labels": [],
		"active": true,
		"id": "test",
		"creationTimestamp": "2025-08-14T18:02:11Z",
		"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
		"secret": "openai-secret"
	}`),
			expected: gjson.Result{},
			err:      errors.New("invalid UUID length: 4"),
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
			response, err := d.UpsertEntity(tc.client, tc.config)

			assert.Equal(t, tc.expected, response)
			if tc.err != nil {
				require.Error(t, err)
				assert.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Test_discovery_UpsertEntities tests the discovery.UpsertEntities() function.
func Test_discovery_UpsertEntities(t *testing.T) {
	tests := []struct {
		name           string
		client         Creator
		configurations gjson.Result
		abortOnError   bool
		printer        Printer
		expectedOutput string
		outWriter      io.Writer
		err            error
	}{
		// Working case
		{
			name:   "Upsert Entities works when it does not receive an array",
			client: new(mocks.WorkingCreator),
			configurations: gjson.Parse(`{
		"type": "mongo",
		"name": "MongoDB credential",
		"labels": [],
		"active": true,
		"creationTimestamp": "2025-08-14T18:02:11Z",
		"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
		"secret": "mongo-secret"
	}`),
			printer:        JsonArrayPrinter(true),
			expectedOutput: "[\n{\n  \"active\": true,\n  \"creationTimestamp\": \"2025-08-14T18:02:11Z\",\n  \"id\": \"9ababe08-0b74-4672-bb7c-e7a8227d6d4c\",\n  \"labels\": [],\n  \"lastUpdatedTimestamp\": \"2025-08-14T18:02:11Z\",\n  \"name\": \"MongoDB credential\",\n  \"secret\": \"mongo-secret\",\n  \"type\": \"mongo\"\n}\n]\n",
			err:            nil,
		},
		{
			name:   "Upsert Entities works when it receives an array",
			client: new(mocks.WorkingCreator),
			configurations: gjson.Parse(`[{
		"type": "mongo",
		"name": "MongoDB credential",
		"labels": [],
		"active": true,
		"creationTimestamp": "2025-08-14T18:02:11Z",
		"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
		"secret": "mongo-secret"
	},{
		"type": "mongo",
		"name": "MongoDB credential 2",
		"labels": [],
		"active": true,
		"id": "9ababe08-0b74-4672-bb7c-e7a8227d6d4c",
		"creationTimestamp": "2025-08-14T18:02:11Z",
		"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
		"secret": "mongo-secret"
	},
	{
		"type": "mongo",
		"name": "MongoDB credential 3",
		"labels": [],
		"active": true,
		"creationTimestamp": "2025-08-14T18:02:11Z",
		"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
		"secret": "mongo-secret"
	}]`),
			printer:        nil,
			expectedOutput: "{\"active\":true,\"creationTimestamp\":\"2025-08-14T18:02:11Z\",\"id\":\"9ababe08-0b74-4672-bb7c-e7a8227d6d4c\",\"labels\":[],\"lastUpdatedTimestamp\":\"2025-08-14T18:02:11Z\",\"name\":\"MongoDB credential\",\"secret\":\"mongo-secret\",\"type\":\"mongo\"}\n{\"active\":true,\"creationTimestamp\":\"2025-08-14T18:02:11Z\",\"id\":\"9ababe08-0b74-4672-bb7c-e7a8227d6d4c\",\"labels\":[],\"lastUpdatedTimestamp\":\"2025-08-14T18:02:11Z\",\"name\":\"MongoDB credential\",\"secret\":\"mongo-secret\",\"type\":\"mongo\"}\n{\"active\":true,\"creationTimestamp\":\"2025-08-14T18:02:11Z\",\"id\":\"9ababe08-0b74-4672-bb7c-e7a8227d6d4c\",\"labels\":[],\"lastUpdatedTimestamp\":\"2025-08-14T18:02:11Z\",\"name\":\"MongoDB credential\",\"secret\":\"mongo-secret\",\"type\":\"mongo\"}\n",
			err:            nil,
		},
		{
			name:   "Upsert Entities works and does not abort with false",
			client: new(mocks.FailingCreator),
			configurations: gjson.Parse(`[{
		"type": "mongo",
		"name": "MongoDB credential",
		"labels": [],
		"active": true,
		"creationTimestamp": "2025-08-14T18:02:11Z",
		"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
		"secret": "mongo-secret"
	},{
		"type": "mongo",
		"name": "MongoDB credential 2",
		"labels": [],
		"active": true,
		"id": "9ababe08-0b74-4672-bb7c-e7a8227d6d4c",
		"creationTimestamp": "2025-08-14T18:02:11Z",
		"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
		"secret": "mongo-secret"
	},
	{
		"type": "mongo",
		"name": "MongoDB credential 3",
		"labels": [],
		"active": true,
		"creationTimestamp": "2025-08-14T18:02:11Z",
		"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
		"secret": "mongo-secret"
	}]`),
			abortOnError:   false,
			printer:        nil,
			expectedOutput: "{\"code\":3002,\"messages\":[\"Invalid JSON: Illegal unquoted character ((CTRL-CHAR, code 10)): has to be escaped using backslash to be included in name\\n at [Source: REDACTED (StreamReadFeature.INCLUDE_SOURCE_IN_LOCATION disabled); line: 5, column: 17]\"],\"status\":400,\"timestamp\":\"2025-10-29T14:46:48.055840300Z\"}\n{\"code\":1003,\"messages\":[\"Entity not found: 9ababe08-0b74-4672-bb7c-e7a8227d6d4d\"],\"status\":404,\"timestamp\":\"2025-10-29T14:47:36.290329Z\"}\n{\"code\":3002,\"messages\":[\"Invalid JSON: Illegal unquoted character ((CTRL-CHAR, code 10)): has to be escaped using backslash to be included in name\\n at [Source: REDACTED (StreamReadFeature.INCLUDE_SOURCE_IN_LOCATION disabled); line: 5, column: 17]\"],\"status\":400,\"timestamp\":\"2025-10-29T14:46:48.055840300Z\"}\n",
			err:            nil,
		},
		{
			name:   "Upsert Entities not abort and receives an error that is not a Discovery.Error",
			client: new(mocks.FailingCreatorCreateWorksUpdateFails),
			configurations: gjson.Parse(`[{
		"type": "mongo",
		"name": "MongoDB credential",
		"labels": [],
		"active": true,
		"creationTimestamp": "2025-08-14T18:02:11Z",
		"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
		"secret": "mongo-secret"
	},{
		"type": "mongo",
		"name": "MongoDB credential 2",
		"labels": [],
		"active": true,
		"id": "9ababe08-0b74-4672-bb7c-e7a8227d6d4c",
		"creationTimestamp": "2025-08-14T18:02:11Z",
		"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
		"secret": "mongo-secret"
	},
	{
		"type": "mongo",
		"name": "MongoDB credential 3",
		"labels": [],
		"active": true,
		"creationTimestamp": "2025-08-14T18:02:11Z",
		"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
		"secret": "mongo-secret"
	}]`),
			abortOnError:   false,
			printer:        nil,
			expectedOutput: "{\"active\":true,\"creationTimestamp\":\"2025-08-14T18:02:11Z\",\"id\":\"9ababe08-0b74-4672-bb7c-e7a8227d6d4c\",\"labels\":[],\"lastUpdatedTimestamp\":\"2025-08-14T18:02:11Z\",\"name\":\"MongoDB credential\",\"secret\":\"mongo-secret\",\"type\":\"mongo\"}\n{\"error\":\"invalid UUID length: 4\"}\n{\"active\":true,\"creationTimestamp\":\"2025-08-14T18:02:11Z\",\"id\":\"9ababe08-0b74-4672-bb7c-e7a8227d6d4c\",\"labels\":[],\"lastUpdatedTimestamp\":\"2025-08-14T18:02:11Z\",\"name\":\"MongoDB credential\",\"secret\":\"mongo-secret\",\"type\":\"mongo\"}\n",
			err:            nil,
		},
		// Error cases
		{
			name:   "Upsert Entities does abort with true",
			client: new(mocks.FailingCreatorCreateWorksUpdateFails),
			configurations: gjson.Parse(`[{
		"type": "mongo",
		"name": "MongoDB credential",
		"labels": [],
		"active": true,
		"creationTimestamp": "2025-08-14T18:02:11Z",
		"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
		"secret": "mongo-secret"
	},{
		"type": "mongo",
		"name": "MongoDB credential 2",
		"labels": [],
		"active": true,
		"id": "9ababe08-0b74-4672-bb7c-e7a8227d6d4c",
		"creationTimestamp": "2025-08-14T18:02:11Z",
		"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
		"secret": "mongo-secret"
	},
	{
		"type": "mongo",
		"name": "MongoDB credential 3",
		"labels": [],
		"active": true,
		"creationTimestamp": "2025-08-14T18:02:11Z",
		"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
		"secret": "mongo-secret"
	}]`),
			abortOnError:   true,
			printer:        nil,
			expectedOutput: "{\"active\":true,\"creationTimestamp\":\"2025-08-14T18:02:11Z\",\"id\":\"9ababe08-0b74-4672-bb7c-e7a8227d6d4c\",\"labels\":[],\"lastUpdatedTimestamp\":\"2025-08-14T18:02:11Z\",\"name\":\"MongoDB credential\",\"secret\":\"mongo-secret\",\"type\":\"mongo\"}\n",
			err:            NewErrorWithCause(ErrorExitCode, errors.New(`invalid UUID length: 4`), "Could not store entities"),
		},
		{
			name:   "Printing fails",
			client: new(mocks.WorkingCreator),
			configurations: gjson.Parse(`{
		"type": "mongo",
		"name": "MongoDB credential",
		"labels": [],
		"active": true,
		"creationTimestamp": "2025-08-14T18:02:11Z",
		"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
		"secret": "mongo-secret"
	}`),
			printer:   nil,
			outWriter: testutils.ErrWriter{Err: errors.New("write failed")},
			err:       NewErrorWithCause(ErrorExitCode, errors.New("write failed"), "Could not print JSON Array"),
		},
		{
			name:   "Upsert fails with abort true and printing fails",
			client: new(mocks.FailingCreatorCreateWorksUpdateFails),
			configurations: gjson.Parse(`[{
		"type": "mongo",
		"name": "MongoDB credential",
		"labels": [],
		"active": true,
		"creationTimestamp": "2025-08-14T18:02:11Z",
		"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
		"secret": "mongo-secret"
	},{
		"type": "mongo",
		"name": "MongoDB credential 2",
		"labels": [],
		"active": true,
		"id": "9ababe08-0b74-4672-bb7c-e7a8227d6d4c",
		"creationTimestamp": "2025-08-14T18:02:11Z",
		"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
		"secret": "mongo-secret"
	}]`),
			abortOnError: true,
			printer:      nil,
			outWriter:    testutils.ErrWriter{Err: errors.New("write failed")},
			err:          errors.Join(NewErrorWithCause(ErrorExitCode, errors.New("write failed"), "Could not print JSON Array"), NewErrorWithCause(ErrorExitCode, errors.New(`invalid UUID length: 4`), "Could not store entities")),
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
			err := d.UpsertEntities(tc.client, tc.configurations, tc.abortOnError, tc.printer)

			assert.Equal(t, tc.expectedOutput, buf.String())
			if tc.err != nil {
				require.Error(t, err)
				assert.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Test_discovery_DeleteEntity tests the deleter.DeleteEntity() function.
func Test_discovery_DeleteEntity(t *testing.T) {
	tests := []struct {
		name           string
		client         Deleter
		printer        Printer
		expectedOutput string
		outWriter      io.Writer
		err            error
	}{
		// Working case
		{
			name:           "DeleteEntity correctly prints the deletion confirmation with the pretty printer",
			client:         new(mocks.WorkingDeleter),
			printer:        nil,
			expectedOutput: "{\n  \"acknowledged\": true\n}\n",
			err:            nil,
		},
		{
			name:           "DeleteEntity correctly prints an object with JSON ugly printer",
			client:         new(mocks.WorkingDeleter),
			printer:        JsonObjectPrinter(false),
			expectedOutput: "{\"acknowledged\":true}\n",
			err:            nil,
		},

		// Error case
		{
			name:           "Delete returns 400 Bad Request",
			client:         new(mocks.FailingDeleter),
			printer:        nil,
			expectedOutput: "",
			err: NewErrorWithCause(ErrorExitCode, discoveryPackage.Error{
				Status: http.StatusBadRequest,
				Body: gjson.Parse(`{
			"status": 400,
			"code": 3002,
			"messages": [
				"Failed to convert argument [id] for value [test] due to: Invalid UUID string: test"
			],
			"timestamp": "2025-10-23T22:35:38.345647200Z"
			}`),
			}, "Could not delete entity with id \"5f125024-1e5e-4591-9fee-365dc20eeeed\""),
		},
		{
			name:      "Printing fails",
			client:    new(mocks.WorkingDeleter),
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

			id, err := uuid.Parse("5f125024-1e5e-4591-9fee-365dc20eeeed")
			require.NoError(t, err)
			d := NewDiscovery(&ios, viper.New(), "")
			err = d.DeleteEntity(tc.client, id, tc.printer)

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

// Test_discovery_SearchDeleteEntity tests the searchDeleter.SearchDeleteEntity() function.
func Test_discovery_SearchDeleteEntity(t *testing.T) {
	tests := []struct {
		name           string
		client         SearchDeleter
		printer        Printer
		expectedOutput string
		outWriter      io.Writer
		err            error
	}{
		// Working case
		{
			name:           "SearchDeleteEntity correctly prints the deletion confirmation with the pretty printer",
			client:         new(mocks.WorkingSearchDeleter),
			printer:        nil,
			expectedOutput: "{\n  \"acknowledged\": true\n}\n",
			err:            nil,
		},
		{
			name:           "SearchDeleteEntity correctly prints an object with JSON ugly printer",
			client:         new(mocks.WorkingSearchDeleter),
			printer:        JsonObjectPrinter(false),
			expectedOutput: "{\"acknowledged\":true}\n",
			err:            nil,
		},

		// Error case
		{
			name:           "Delete returns 400 Bad Request",
			client:         new(mocks.FailingSearchDeleterDeleteFails),
			printer:        nil,
			expectedOutput: "",
			err: NewErrorWithCause(ErrorExitCode, discoveryPackage.Error{
				Status: http.StatusBadRequest,
				Body: gjson.Parse(`{
			"status": 400,
			"code": 3002,
			"messages": [
				"Failed to convert argument [id] for value [test] due to: Invalid UUID string: test"
			],
			"timestamp": "2025-10-23T22:35:38.345647200Z"
			}`),
			}, "Could not delete entity with id \"5f125024-1e5e-4591-9fee-365dc20eeeed\""),
		},
		{
			name:           "Delete fails because it cannot parse UUID",
			client:         new(mocks.FailingSearchDeleterParsingUUIDFails),
			printer:        nil,
			expectedOutput: "",
			err:            NewErrorWithCause(ErrorExitCode, errors.New("invalid UUID length: 7"), "Could not delete entity with name \"MongoDB Atlas Server\""),
		},
		{
			name:           "Search returns 404 Not Found",
			client:         new(mocks.FailingSearchDeleterSearchFails),
			printer:        nil,
			expectedOutput: "",
			err: NewErrorWithCause(ErrorExitCode, discoveryPackage.Error{
				Status: http.StatusNotFound,
				Body: gjson.Parse(`{
	"status": 404,
	"code": 1003,
	"messages": [
		"Entity not found: entity with name "MongoDB Atlas Server" does not exist"
	],
	"timestamp": "2025-09-30T15:38:42.885125200Z"
}`),
			}, "Could not search for entity with name \"MongoDB Atlas Server\""),
		},
		{
			name:      "Printing fails",
			client:    new(mocks.WorkingSearchDeleter),
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
			err := d.SearchDeleteEntity(tc.client, "MongoDB Atlas Server", tc.printer)

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
