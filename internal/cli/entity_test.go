package cli

import (
	"bytes"
	"errors"
	"fmt"
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

// WorkingGetter mocks the discovery.Getter struct to always answer a working result
type WorkingGetter struct {
	mock.Mock
}

// Get returns a working processor as if the request worked successfully.
func (g *WorkingGetter) Get(id uuid.UUID) (gjson.Result, error) {
	return gjson.Parse(`{
		"type": "mongo",
		"name": "MongoDB text processor",
		"labels": [],
		"active": true,
		"id": "5f125024-1e5e-4591-9fee-365dc20eeeed",
		"creationTimestamp": "2025-08-14T18:02:38Z",
		"lastUpdatedTimestamp": "2025-08-18T20:55:43Z"
	}`), nil
}

// GetAll returns a list of processors
func (g *WorkingGetter) GetAll() ([]gjson.Result, error) {
	return gjson.Parse(`[
		{
		"type": "mongo",
		"name": "MongoDB text processor 4",
		"labels": [],
		"active": true,
		"id": "3393f6d9-94c1-4b70-ba02-5f582727d998",
		"creationTimestamp": "2025-08-21T17:57:16Z",
		"lastUpdatedTimestamp": "2025-08-21T17:57:16Z"
		},
		{
		"type": "mongo",
		"name": "MongoDB text processor",
		"labels": [],
		"active": true,
		"id": "5f125024-1e5e-4591-9fee-365dc20eeeed",
		"creationTimestamp": "2025-08-14T18:02:38Z",
		"lastUpdatedTimestamp": "2025-08-18T20:55:43Z"
		},
		{
		"type": "script",
		"name": "Script processor",
		"labels": [],
		"active": true,
		"id": "86e7f920-a4e4-4b64-be84-5437a7673db8",
		"creationTimestamp": "2025-08-14T18:02:38Z",
		"lastUpdatedTimestamp": "2025-08-14T18:02:38Z"
		}
	]`).Array(), nil
}

// FailingGetter mocks the discovery.Getter struct to always return an HTTP error.
type FailingGetter struct {
	mock.Mock
}

// Get returns a 404 Not Found
func (g *FailingGetter) Get(id uuid.UUID) (gjson.Result, error) {
	return gjson.Result{}, discoveryPackage.Error{
		Status: http.StatusNotFound,
		Body: gjson.Parse(`{
			"status": 404,
			"code": 1003,
			"messages": [
				"Secret not found: 5f125024-1e5e-4591-9fee-365dc20eeeed"
			],
			"timestamp": "2025-10-16T00:15:31.888410500Z"
		}`),
	}
}

// GetAll returns 401 unauthorized
func (g *FailingGetter) GetAll() ([]gjson.Result, error) {
	return []gjson.Result(nil), discoveryPackage.Error{Status: http.StatusUnauthorized, Body: gjson.Parse(`{"error":"unauthorized"}`)}
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
			client:         new(WorkingGetter),
			printer:        nil,
			expectedOutput: "{\n  \"active\": true,\n  \"creationTimestamp\": \"2025-08-14T18:02:38Z\",\n  \"id\": \"5f125024-1e5e-4591-9fee-365dc20eeeed\",\n  \"labels\": [],\n  \"lastUpdatedTimestamp\": \"2025-08-18T20:55:43Z\",\n  \"name\": \"MongoDB text processor\",\n  \"type\": \"mongo\"\n}\n",
			err:            nil,
		},
		{
			name:           "GetEntity correctly prints an object with JSON ugly printer",
			client:         new(WorkingGetter),
			printer:        JsonObjectPrinter(false),
			expectedOutput: "{\"active\":true,\"creationTimestamp\":\"2025-08-14T18:02:38Z\",\"id\":\"5f125024-1e5e-4591-9fee-365dc20eeeed\",\"labels\":[],\"lastUpdatedTimestamp\":\"2025-08-18T20:55:43Z\",\"name\":\"MongoDB text processor\",\"type\":\"mongo\"}\n",
			err:            nil,
		},

		// Error case
		{
			name:           "Get returns 404 Not Found",
			client:         new(FailingGetter),
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
			client:    new(WorkingGetter),
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
			client:         new(WorkingGetter),
			printer:        JsonArrayPrinter(true),
			expectedOutput: "[\n{\n  \"active\": true,\n  \"creationTimestamp\": \"2025-08-21T17:57:16Z\",\n  \"id\": \"3393f6d9-94c1-4b70-ba02-5f582727d998\",\n  \"labels\": [],\n  \"lastUpdatedTimestamp\": \"2025-08-21T17:57:16Z\",\n  \"name\": \"MongoDB text processor 4\",\n  \"type\": \"mongo\"\n},\n{\n  \"active\": true,\n  \"creationTimestamp\": \"2025-08-14T18:02:38Z\",\n  \"id\": \"5f125024-1e5e-4591-9fee-365dc20eeeed\",\n  \"labels\": [],\n  \"lastUpdatedTimestamp\": \"2025-08-18T20:55:43Z\",\n  \"name\": \"MongoDB text processor\",\n  \"type\": \"mongo\"\n},\n{\n  \"active\": true,\n  \"creationTimestamp\": \"2025-08-14T18:02:38Z\",\n  \"id\": \"86e7f920-a4e4-4b64-be84-5437a7673db8\",\n  \"labels\": [],\n  \"lastUpdatedTimestamp\": \"2025-08-14T18:02:38Z\",\n  \"name\": \"Script processor\",\n  \"type\": \"script\"\n}\n]\n",
			err:            nil,
		},
		{
			name:           "GetEntities correctly prints an array with JSON ugly printer",
			client:         new(WorkingGetter),
			printer:        nil,
			expectedOutput: "{\"active\":true,\"creationTimestamp\":\"2025-08-21T17:57:16Z\",\"id\":\"3393f6d9-94c1-4b70-ba02-5f582727d998\",\"labels\":[],\"lastUpdatedTimestamp\":\"2025-08-21T17:57:16Z\",\"name\":\"MongoDB text processor 4\",\"type\":\"mongo\"}\n{\"active\":true,\"creationTimestamp\":\"2025-08-14T18:02:38Z\",\"id\":\"5f125024-1e5e-4591-9fee-365dc20eeeed\",\"labels\":[],\"lastUpdatedTimestamp\":\"2025-08-18T20:55:43Z\",\"name\":\"MongoDB text processor\",\"type\":\"mongo\"}\n{\"active\":true,\"creationTimestamp\":\"2025-08-14T18:02:38Z\",\"id\":\"86e7f920-a4e4-4b64-be84-5437a7673db8\",\"labels\":[],\"lastUpdatedTimestamp\":\"2025-08-14T18:02:38Z\",\"name\":\"Script processor\",\"type\":\"script\"}\n",
			err:            nil,
		},

		// Error case
		{
			name:           "GetAll returns 401 Unauthorized",
			client:         new(FailingGetter),
			printer:        nil,
			expectedOutput: "",
			err:            NewErrorWithCause(ErrorExitCode, discoveryPackage.Error{Status: http.StatusUnauthorized, Body: gjson.Parse(`{"error":"unauthorized"}`)}, "Could not get all entities"),
		},
		{
			name:      "Printing fails",
			client:    new(WorkingGetter),
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

// WorkingSearcher mocks the discovery.Searcher struct.
// This struct is used to test when the search functions work.
type WorkingSearcher struct {
	mock.Mock
}

// Search returns an array of results as it if correctly found matches.
func (s *WorkingSearcher) Search(gjson.Result) ([]gjson.Result, error) {
	return gjson.Parse(`[
                  {
					"source": {
						"type": "mongo",
						"name": "MongoDB Atlas server clone",
						"labels": [],
						"active": true,
						"id": "986ce864-af76-4fcb-8b4f-f4e4c6ab0951",
						"creationTimestamp": "2025-09-29T15:50:17Z",
						"lastUpdatedTimestamp": "2025-09-29T15:50:17Z"
					},
					"highlight": {},
					"score": 0.20970252
                  },
                  {
					"source": {
						"type": "mongo",
						"name": "MongoDB Atlas server clone 1",
						"labels": [],
						"active": true,
						"id": "8f14c11c-bb66-49d3-aa2a-dedff4608c17",
						"creationTimestamp": "2025-09-29T15:50:19Z",
						"lastUpdatedTimestamp": "2025-09-29T15:50:19Z"
					},
					"highlight": {},
					"score": 0.20970252
                  },
                  {
					"source": {
						"type": "mongo",
						"name": "MongoDB Atlas server clone 3",
						"labels": [],
						"active": true,
						"id": "3a0214a4-72cc-4eee-ad0c-9e3af9b08a6c",
						"creationTimestamp": "2025-09-29T15:50:20Z",
						"lastUpdatedTimestamp": "2025-09-29T15:50:20Z"
					},
					"highlight": {},
					"score": 0.20970252
                  }
          ]`).Array(), nil
}

// SearchByName returns an object as if it found correctly the entity.
func (s *WorkingSearcher) SearchByName(name string) (gjson.Result, error) {
	return gjson.Parse(`{
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
	}`), nil
}

// Get returns a JSON object as if the searcher found the entity by its ID.
func (s *WorkingSearcher) Get(id uuid.UUID) (gjson.Result, error) {
	return gjson.Parse(`{
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
	}`), nil
}

// GetAll returns
func (s *WorkingSearcher) GetAll() ([]gjson.Result, error) {
	return gjson.Parse(`[
		{
		"type": "mongo",
		"name": "my-credential",
		"labels": [
			{
			"key": "A",
			"value": "A"
			}
		],
		"active": true,
		"id": "3b32e410-2f33-412d-9fb8-17970131921c",
		"creationTimestamp": "2025-10-17T22:37:57Z",
		"lastUpdatedTimestamp": "2025-10-17T22:37:57Z"
		},
		{
		"type": "openai",
		"name": "OpenAI credential clone clone",
		"labels": [],
		"active": true,
		"id": "5c09589e-b643-41aa-a766-3b7fc3660473",
		"creationTimestamp": "2025-10-17T22:38:12Z",
		"lastUpdatedTimestamp": "2025-10-17T22:38:12Z"
		},
	]`).Array(), nil
}

// FailingSearcher mocks the discovery.Searcher struct when its functions return errors.
type FailingSearcher struct {
	mock.Mock
}

// Search implements the searcher interface
func (s *FailingSearcher) Search(gjson.Result) ([]gjson.Result, error) {
	return []gjson.Result(nil), discoveryPackage.Error{
		Status: http.StatusNotFound,
		Body:   gjson.Result{},
	}
}

// SearchByName returns 404 so that the searchEntity function enters the err != nil code branch
func (s *FailingSearcher) SearchByName(name string) (gjson.Result, error) {
	return gjson.Result{}, discoveryPackage.Error{
		Status: http.StatusBadRequest,
		Body: gjson.Parse(`{
	"status": 400,
	"code": 3002,
	"messages": [
		"Invalid JSON: Unexpected end-of-input:"
	],
	"timestamp": "2025-10-17T17:43:52.817308100Z"
	}`),
	}
}

// Get returns an error so mock when the user searches for an entity that does not exist.
func (s *FailingSearcher) Get(id uuid.UUID) (gjson.Result, error) {
	return gjson.Result{}, discoveryPackage.Error{
		Status: http.StatusNotFound,
		Body: gjson.Parse(`{
			"status": 404,
			"code": 1003,
			"messages": [
				"Server not found: 986ce864-af76-4fcb-8b4f-f4e4c6ab0951"
			],
			"timestamp": "2025-10-16T00:15:31.888410500Z"
		}`),
	}
}

// GetAll implements the searcher interface
func (s *FailingSearcher) GetAll() ([]gjson.Result, error) {
	return []gjson.Result(nil), discoveryPackage.Error{Status: http.StatusUnauthorized, Body: gjson.Parse(`{"error":"unauthorized"}`)}
}

// FailingSearcherWorkingGetter mocks the discovery.searcher struct when the search by name fails, but the get does succeed
type FailingSearcherWorkingGetter struct {
	mock.Mock
}

// Search implements the searcher interface
func (s *FailingSearcherWorkingGetter) Search(gjson.Result) ([]gjson.Result, error) {
	return []gjson.Result(nil), discoveryPackage.Error{
		Status: http.StatusNotFound,
		Body:   gjson.Result{},
	}
}

// SearchByName returns 404 not found to make the test go through the err != nil code branch
func (s *FailingSearcherWorkingGetter) SearchByName(name string) (gjson.Result, error) {
	return gjson.Result{}, discoveryPackage.Error{
		Status: http.StatusNotFound,
		Body: gjson.Parse(fmt.Sprintf(`{
	"status": 404,
	"code": 1003,
	"messages": [
		"Entity not found: entity with name %q does not exist"
	]
}`, name)),
	}
}

// Get returns a JSON object to mock that the SearchByName failed, but the Get succeeded.
func (s *FailingSearcherWorkingGetter) Get(id uuid.UUID) (gjson.Result, error) {
	return gjson.Parse(`{
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
	}`), nil
}

// GetAll implements the searcher interface.
func (s *FailingSearcherWorkingGetter) GetAll() ([]gjson.Result, error) {
	return gjson.Parse(`[]`).Array(), nil
}

// FailingSearcherFailingGetter mocks the discovery.searcher struct when both the searchByName and Get function fails
type FailingSearcherFailingGetter struct {
	mock.Mock
}

// Search returns an error
func (s *FailingSearcherFailingGetter) Search(gjson.Result) ([]gjson.Result, error) {
	return []gjson.Result(nil), discoveryPackage.Error{Status: http.StatusUnauthorized, Body: gjson.Parse(`{"error":"unauthorized"}`)}
}

// SearchByName returns a 404 Not Found error to make the test go through the err != nil branch
func (s *FailingSearcherFailingGetter) SearchByName(name string) (gjson.Result, error) {
	return gjson.Result{}, discoveryPackage.Error{
		Status: http.StatusNotFound,
		Body: gjson.Parse(fmt.Sprintf(`{
	"status": 404,
	"code": 1003,
	"messages": [
		"Entity not found: entity with name %q does not exist"
	]
}`, name)),
	}
}

// Get Returns error to mock that it also failed to find the entity.
func (s *FailingSearcherFailingGetter) Get(id uuid.UUID) (gjson.Result, error) {
	return gjson.Result{}, discoveryPackage.Error{
		Status: http.StatusNotFound,
		Body: gjson.Parse(`{
			"status": 404,
			"code": 1003,
			"messages": [
				"Server not found: 986ce864-af76-4fcb-8b4f-f4e4c6ab0951"
			],
			"timestamp": "2025-10-16T00:15:31.888410500Z"
		}`),
	}
}

// GetAll implements the searcher interface.
func (s *FailingSearcherFailingGetter) GetAll() ([]gjson.Result, error) {
	return []gjson.Result(nil), discoveryPackage.Error{Status: http.StatusUnauthorized, Body: gjson.Parse(`{"error":"unauthorized"}`)}
}

// SearcherReturnsOtherError is a struct that mocks discovery.searcher when the search functions do not return a discovery.Error
type SearcherReturnsOtherError struct {
	mock.Mock
}

// Search implements the searcher interface
func (s *SearcherReturnsOtherError) Search(gjson.Result) ([]gjson.Result, error) {
	return []gjson.Result(nil), discoveryPackage.Error{
		Status: http.StatusNotFound,
		Body:   gjson.Parse(``),
	}
}

// SearchByName does not return a discovery.Error
func (s *SearcherReturnsOtherError) SearchByName(name string) (gjson.Result, error) {
	return gjson.Result{}, errors.New("not discovery error")
}

// Get implements the searcher interface.
func (s *SearcherReturnsOtherError) Get(id uuid.UUID) (gjson.Result, error) {
	return gjson.Result{}, discoveryPackage.Error{
		Status: http.StatusNotFound,
		Body:   gjson.Parse(``),
	}
}

// GetAll implements the searcher interface
func (s *SearcherReturnsOtherError) GetAll() ([]gjson.Result, error) {
	return []gjson.Result(nil), discoveryPackage.Error{Status: http.StatusUnauthorized, Body: gjson.Parse(``)}
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
			client: new(WorkingSearcher),
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
			client: new(FailingSearcherWorkingGetter),
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
			client:   new(FailingSearcher),
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
			client:   new(FailingSearcherFailingGetter),
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
			client:   new(FailingSearcherWorkingGetter),
			id:       "MongoDB Atlas Server",
			expected: gjson.Result{},
			err: discoveryPackage.Error{
				Status: http.StatusNotFound,
				Body:   gjson.Parse("{\n\t\"status\": 404,\n\t\"code\": 1003,\n\t\"messages\": [\n\t\t\"Entity not found: entity with name \"MongoDB Atlas Server\" does not exist\"\n\t]\n}"),
			},
		},
		{
			name:     "Search returns an error different from discovery.Error",
			client:   new(SearcherReturnsOtherError),
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
			client: new(WorkingSearcher),
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
			client:   new(FailingSearcherFailingGetter),
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
			client:         new(WorkingSearcher),
			id:             "MongoDB Atlas Server",
			printer:        nil,
			expectedOutput: "{\n  \"active\": true,\n  \"config\": {\n    \"connection\": {\n      \"connectTimeout\": \"1m\",\n      \"readTimeout\": \"30s\"\n    },\n    \"credentialId\": \"9ababe08-0b74-4672-bb7c-e7a8227d6d4c\",\n    \"servers\": [\n      \"mongodb+srv://cluster0.dleud.mongodb.net/\"\n    ]\n  },\n  \"creationTimestamp\": \"2025-09-29T15:50:17Z\",\n  \"id\": \"986ce864-af76-4fcb-8b4f-f4e4c6ab0951\",\n  \"labels\": [],\n  \"lastUpdatedTimestamp\": \"2025-09-29T15:50:17Z\",\n  \"name\": \"MongoDB Atlas server\",\n  \"type\": \"mongo\"\n}\n",
			err:            nil,
		},
		{
			name:           "SearchEntity correctly prints an object with JSON ugly printer",
			client:         new(FailingSearcherWorkingGetter),
			id:             "986ce864-af76-4fcb-8b4f-f4e4c6ab0951",
			printer:        JsonObjectPrinter(false),
			expectedOutput: "{\"active\":true,\"config\":{\"connection\":{\"connectTimeout\":\"1m\",\"readTimeout\":\"30s\"},\"credentialId\":\"9ababe08-0b74-4672-bb7c-e7a8227d6d4c\",\"servers\":[\"mongodb+srv://cluster0.dleud.mongodb.net/\"]},\"creationTimestamp\":\"2025-09-29T15:50:17Z\",\"id\":\"986ce864-af76-4fcb-8b4f-f4e4c6ab0951\",\"labels\":[],\"lastUpdatedTimestamp\":\"2025-09-29T15:50:17Z\",\"name\":\"MongoDB Atlas server clone\",\"type\":\"mongo\"}\n",
			err:            nil,
		},

		// Error case
		{
			name:           "Search returns 404 Not Found",
			client:         new(FailingSearcherFailingGetter),
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
			client:         new(WorkingSearcher),
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
			client:         new(WorkingSearcher),
			printer:        JsonArrayPrinter(true),
			expectedOutput: "[\n{\n  \"highlight\": {},\n  \"score\": 0.20970252,\n  \"source\": {\n    \"active\": true,\n    \"creationTimestamp\": \"2025-09-29T15:50:17Z\",\n    \"id\": \"986ce864-af76-4fcb-8b4f-f4e4c6ab0951\",\n    \"labels\": [],\n    \"lastUpdatedTimestamp\": \"2025-09-29T15:50:17Z\",\n    \"name\": \"MongoDB Atlas server clone\",\n    \"type\": \"mongo\"\n  }\n},\n{\n  \"highlight\": {},\n  \"score\": 0.20970252,\n  \"source\": {\n    \"active\": true,\n    \"creationTimestamp\": \"2025-09-29T15:50:19Z\",\n    \"id\": \"8f14c11c-bb66-49d3-aa2a-dedff4608c17\",\n    \"labels\": [],\n    \"lastUpdatedTimestamp\": \"2025-09-29T15:50:19Z\",\n    \"name\": \"MongoDB Atlas server clone 1\",\n    \"type\": \"mongo\"\n  }\n},\n{\n  \"highlight\": {},\n  \"score\": 0.20970252,\n  \"source\": {\n    \"active\": true,\n    \"creationTimestamp\": \"2025-09-29T15:50:20Z\",\n    \"id\": \"3a0214a4-72cc-4eee-ad0c-9e3af9b08a6c\",\n    \"labels\": [],\n    \"lastUpdatedTimestamp\": \"2025-09-29T15:50:20Z\",\n    \"name\": \"MongoDB Atlas server clone 3\",\n    \"type\": \"mongo\"\n  }\n}\n]\n",
			err:            nil,
		},
		{
			name:           "SearchEntities correctly prints an array with JSON ugly printer",
			client:         new(WorkingSearcher),
			printer:        nil,
			expectedOutput: "{\"highlight\":{},\"score\":0.20970252,\"source\":{\"active\":true,\"creationTimestamp\":\"2025-09-29T15:50:17Z\",\"id\":\"986ce864-af76-4fcb-8b4f-f4e4c6ab0951\",\"labels\":[],\"lastUpdatedTimestamp\":\"2025-09-29T15:50:17Z\",\"name\":\"MongoDB Atlas server clone\",\"type\":\"mongo\"}}\n{\"highlight\":{},\"score\":0.20970252,\"source\":{\"active\":true,\"creationTimestamp\":\"2025-09-29T15:50:19Z\",\"id\":\"8f14c11c-bb66-49d3-aa2a-dedff4608c17\",\"labels\":[],\"lastUpdatedTimestamp\":\"2025-09-29T15:50:19Z\",\"name\":\"MongoDB Atlas server clone 1\",\"type\":\"mongo\"}}\n{\"highlight\":{},\"score\":0.20970252,\"source\":{\"active\":true,\"creationTimestamp\":\"2025-09-29T15:50:20Z\",\"id\":\"3a0214a4-72cc-4eee-ad0c-9e3af9b08a6c\",\"labels\":[],\"lastUpdatedTimestamp\":\"2025-09-29T15:50:20Z\",\"name\":\"MongoDB Atlas server clone 3\",\"type\":\"mongo\"}}\n",
			err:            nil,
		},

		// Error case
		{
			name:           "SearchAll returns 401 Unauthorized",
			client:         new(FailingSearcherFailingGetter),
			printer:        nil,
			expectedOutput: "",
			err:            NewErrorWithCause(ErrorExitCode, discoveryPackage.Error{Status: http.StatusUnauthorized, Body: gjson.Parse(`{"error":"unauthorized"}`)}, "Could not search for the entities"),
		},
		{
			name:      "Printing fails",
			client:    new(WorkingSearcher),
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

// Test_getAndFilterString tests the getAndFilterString() function
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

// WorkingCreator mocks when creating and updating entities works.
type WorkingCreator struct {
	mock.Mock
}

// Create returns a JSON as if it worked successfully.
func (g *WorkingCreator) Create(config gjson.Result) (gjson.Result, error) {
	return gjson.Parse(`{
		"type": "mongo",
		"name": "MongoDB credential",
		"labels": [],
		"active": true,
		"id": "9ababe08-0b74-4672-bb7c-e7a8227d6d4c",
		"creationTimestamp": "2025-08-14T18:02:11Z",
		"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
		"secret": "mongo-secret"
	}`), nil
}

// Update returns a JSON as if it worked successfully.
func (g *WorkingCreator) Update(id uuid.UUID, config gjson.Result) (gjson.Result, error) {
	return gjson.Parse(`{
		"type": "mongo",
		"name": "MongoDB credential",
		"labels": [],
		"active": true,
		"id": "9ababe08-0b74-4672-bb7c-e7a8227d6d4c",
		"creationTimestamp": "2025-08-14T18:02:11Z",
		"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
		"secret": "mongo-secret"
	}`), nil
}

// FailingCreator mocks when creating and updating entities fails.
type FailingCreator struct {
	mock.Mock
}

// Create returns a JSON as if it worked successfully.
func (g *FailingCreator) Create(config gjson.Result) (gjson.Result, error) {
	return gjson.Result{}, discoveryPackage.Error{Status: http.StatusBadRequest, Body: gjson.Parse(`{
  "status": 400,
  "code": 3002,
  "messages": [
    "Invalid JSON: Illegal unquoted character ((CTRL-CHAR, code 10)): has to be escaped using backslash to be included in name\n at [Source: REDACTED (StreamReadFeature.INCLUDE_SOURCE_IN_LOCATION disabled); line: 5, column: 17]"
  ],
  "timestamp": "2025-10-29T14:46:48.055840300Z"
}`)}
}

// Update returns a JSON as if it worked successfully.
func (g *FailingCreator) Update(id uuid.UUID, config gjson.Result) (gjson.Result, error) {
	return gjson.Result{}, discoveryPackage.Error{Status: http.StatusBadRequest, Body: gjson.Parse(`{
  "status": 404,
  "code": 1003,
  "messages": [
    "Entity not found: 9ababe08-0b74-4672-bb7c-e7a8227d6d4d"
  ],
  "timestamp": "2025-10-29T14:47:36.290329Z"
}`)}
}

// FailingCreator mocks when creating and updating entities fails.
type FailingCreatorCreateWorksUpdateFails struct {
	mock.Mock
}

// Create returns a JSON as if it worked successfully.
func (g *FailingCreatorCreateWorksUpdateFails) Create(config gjson.Result) (gjson.Result, error) {
	return gjson.Parse(`{
		"type": "mongo",
		"name": "MongoDB credential",
		"labels": [],
		"active": true,
		"id": "9ababe08-0b74-4672-bb7c-e7a8227d6d4c",
		"creationTimestamp": "2025-08-14T18:02:11Z",
		"lastUpdatedTimestamp": "2025-08-14T18:02:11Z",
		"secret": "mongo-secret"
	}`), nil
}

// Update returns an error that is not a Discovery.Error
func (g *FailingCreatorCreateWorksUpdateFails) Update(id uuid.UUID, config gjson.Result) (gjson.Result, error) {
	return gjson.Result{}, errors.New(`invalid UUID length: 4`)
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
			client: new(WorkingCreator),
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
			client: new(WorkingCreator),
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
			client: new(FailingCreator),
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
			client: new(FailingCreator),
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
			client: new(WorkingCreator),
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

// Test_discovery_UpsertEntities tests the discovery.UpsertEntities() function?
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
			client: new(WorkingCreator),
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
			client: new(WorkingCreator),
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
			client: new(FailingCreator),
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
			client: new(FailingCreatorCreateWorksUpdateFails),
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
			client: new(FailingCreatorCreateWorksUpdateFails),
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
			client: new(WorkingCreator),
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
			client: new(FailingCreatorCreateWorksUpdateFails),
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

// WorkingDeleter mocks when the deleter interface works correctly.
type WorkingDeleter struct {
	mock.Mock
}

// Get returns a working processor as if the request worked successfully.
func (g *WorkingDeleter) Delete(id uuid.UUID) (gjson.Result, error) {
	return gjson.Parse(`{
		"acknowledged": true
	}`), nil
}

// FailingDeleter mocks the deleter interface when the Delete() method fails.
type FailingDeleter struct {
	mock.Mock
}

// Get returns a working processor as if the request worked successfully.
func (g *FailingDeleter) Delete(id uuid.UUID) (gjson.Result, error) {
	return gjson.Result{}, discoveryPackage.Error{
		Status: http.StatusBadRequest,
		Body: gjson.Parse(`{
			"status": 400,
			"code": 3002,
			"messages": [
				"Failed to convert argument [id] for value [test] due to: Invalid UUID string: test"
			],
			"timestamp": "2025-10-23T22:35:38.345647200Z"
			}`),
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
			client:         new(WorkingDeleter),
			printer:        nil,
			expectedOutput: "{\n  \"acknowledged\": true\n}\n",
			err:            nil,
		},
		{
			name:           "DeleteEntity correctly prints an object with JSON ugly printer",
			client:         new(WorkingDeleter),
			printer:        JsonObjectPrinter(false),
			expectedOutput: "{\"acknowledged\":true}\n",
			err:            nil,
		},

		// Error case
		{
			name:           "Delete returns 400 Bad Request",
			client:         new(FailingDeleter),
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
			client:    new(WorkingDeleter),
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

// WorkingSearchDeleter correctly finds the entity by its name and deletes it with its ID.
type WorkingSearchDeleter struct {
	mock.Mock
}

// Get returns a working processor as if the request worked successfully.
func (g *WorkingSearchDeleter) Delete(id uuid.UUID) (gjson.Result, error) {
	return gjson.Parse(`{
		"acknowledged": true
	}`), nil
}

// SearchByName returns a valid JSON.
func (g *WorkingSearchDeleter) SearchByName(name string) (gjson.Result, error) {
	return gjson.Parse(`{
		"type": "mongo",
		"name": "MongoDB text processor",
		"labels": [],
		"active": true,
		"id": "5f125024-1e5e-4591-9fee-365dc20eeeed",
		"creationTimestamp": "2025-08-14T18:02:38Z",
		"lastUpdatedTimestamp": "2025-08-18T20:55:43Z"
	}`), nil
}

// Search implements the searchDeleter interface.
func (g *WorkingSearchDeleter) Search(filter gjson.Result) ([]gjson.Result, error) {
	return gjson.Parse(`[]`).Array(), nil
}

// Get implements the searchDeleter interface.
func (g *WorkingSearchDeleter) Get(id uuid.UUID) (gjson.Result, error) {
	return gjson.Parse(`{
		"type": "mongo",
		"name": "MongoDB text processor",
		"labels": [],
		"active": true,
		"id": "5f125024-1e5e-4591-9fee-365dc20eeeed",
		"creationTimestamp": "2025-08-14T18:02:38Z",
		"lastUpdatedTimestamp": "2025-08-18T20:55:43Z"
	}`), nil
}

// GetAll implements the searchDeleter interface.
func (g *WorkingSearchDeleter) GetAll() ([]gjson.Result, error) {
	return gjson.Parse(`[]`).Array(), nil
}

// FailingSearchDeleterSearchFails fails in the SearchByName() function.
type FailingSearchDeleterSearchFails struct {
	mock.Mock
}

// SearchByName returns a not found error, so the entity does not exist.
func (g *FailingSearchDeleterSearchFails) SearchByName(name string) (gjson.Result, error) {
	return gjson.Result{}, discoveryPackage.Error{
		Status: http.StatusNotFound,
		Body: gjson.Parse(fmt.Sprintf(`{
	"status": 404,
	"code": 1003,
	"messages": [
		"Entity not found: entity with name %q does not exist"
	],
	"timestamp": "2025-09-30T15:38:42.885125200Z"
}`, name)),
	}
}

// Search implements the searchDeleter interface.
func (g *FailingSearchDeleterSearchFails) Search(filter gjson.Result) ([]gjson.Result, error) {
	return gjson.Parse(`[]`).Array(), nil
}

// Get implements the searchDeleter interface.
func (g *FailingSearchDeleterSearchFails) Get(id uuid.UUID) (gjson.Result, error) {
	return gjson.Result{}, nil
}

// GetAll implements the searchDeleter interface.
func (g *FailingSearchDeleterSearchFails) GetAll() ([]gjson.Result, error) {
	return gjson.Parse(`[]`).Array(), nil
}

// Get returns a working processor as if the request worked successfully.
func (g *FailingSearchDeleterSearchFails) Delete(id uuid.UUID) (gjson.Result, error) {
	return gjson.Result{}, discoveryPackage.Error{
		Status: http.StatusBadRequest,
		Body: gjson.Parse(`{
			"status": 400,
			"code": 3002,
			"messages": [
				"Failed to convert argument [id] for value [test] due to: Invalid UUID string: test"
			],
			"timestamp": "2025-10-23T22:35:38.345647200Z"
			}`),
	}
}

// FailingSearchDeleterSearchFails fails in the DeleteEntity() function.
type FailingSearchDeleterDeleteFails struct {
	mock.Mock
}

// SearchByName returns a valid JSON.
func (g *FailingSearchDeleterDeleteFails) SearchByName(name string) (gjson.Result, error) {
	return gjson.Parse(`{
		"type": "mongo",
		"name": "MongoDB text processor",
		"labels": [],
		"active": true,
		"id": "5f125024-1e5e-4591-9fee-365dc20eeeed",
		"creationTimestamp": "2025-08-14T18:02:38Z",
		"lastUpdatedTimestamp": "2025-08-18T20:55:43Z"
	}`), nil
}

// Search implements the searchDeleter interface.
func (g *FailingSearchDeleterDeleteFails) Search(filter gjson.Result) ([]gjson.Result, error) {
	return gjson.Parse(`[]`).Array(), nil
}

// Get implements the searchDeleter interface.
func (g *FailingSearchDeleterDeleteFails) Get(id uuid.UUID) (gjson.Result, error) {
	return gjson.Parse(`{
		"type": "mongo",
		"name": "MongoDB text processor",
		"labels": [],
		"active": true,
		"id": "5f125024-1e5e-4591-9fee-365dc20eeeed",
		"creationTimestamp": "2025-08-14T18:02:38Z",
		"lastUpdatedTimestamp": "2025-08-18T20:55:43Z"
	}`), nil
}

// GetAll implements the searchDeleter interface.
func (g *FailingSearchDeleterDeleteFails) GetAll() ([]gjson.Result, error) {
	return gjson.Parse(`[]`).Array(), nil
}

// Delete fails due to bad request.
func (g *FailingSearchDeleterDeleteFails) Delete(id uuid.UUID) (gjson.Result, error) {
	return gjson.Result{}, discoveryPackage.Error{
		Status: http.StatusBadRequest,
		Body: gjson.Parse(`{
			"status": 400,
			"code": 3002,
			"messages": [
				"Failed to convert argument [id] for value [test] due to: Invalid UUID string: test"
			],
			"timestamp": "2025-10-23T22:35:38.345647200Z"
			}`),
	}
}

// FailingSearchDeleterParsingUUIDFails fails when trying to parse the id in the received search result.
type FailingSearchDeleterParsingUUIDFails struct {
	mock.Mock
}

// SearchByName returns a valid JSON.
func (g *FailingSearchDeleterParsingUUIDFails) SearchByName(name string) (gjson.Result, error) {
	return gjson.Parse(`{
		"type": "mongo",
		"name": "MongoDB text processor",
		"labels": [],
		"active": true,
		"id": "notuuid",
		"creationTimestamp": "2025-08-14T18:02:38Z",
		"lastUpdatedTimestamp": "2025-08-18T20:55:43Z"
	}`), nil
}

// Search implements the searchDeleter interface.
func (g *FailingSearchDeleterParsingUUIDFails) Search(filter gjson.Result) ([]gjson.Result, error) {
	return gjson.Parse(`[]`).Array(), nil
}

// Get implements the searchDeleter interface.
func (g *FailingSearchDeleterParsingUUIDFails) Get(id uuid.UUID) (gjson.Result, error) {
	return gjson.Parse(`{
		"type": "mongo",
		"name": "MongoDB text processor",
		"labels": [],
		"active": true,
		"id": "notuuid",
		"creationTimestamp": "2025-08-14T18:02:38Z",
		"lastUpdatedTimestamp": "2025-08-18T20:55:43Z"
	}`), nil
}

// GetAll implements the searchDeleter interface.
func (g *FailingSearchDeleterParsingUUIDFails) GetAll() ([]gjson.Result, error) {
	return gjson.Parse(`[]`).Array(), nil
}

// Delete implements the searchDeleter interface.
func (g *FailingSearchDeleterParsingUUIDFails) Delete(id uuid.UUID) (gjson.Result, error) {
	return gjson.Result{}, nil
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
			client:         new(WorkingSearchDeleter),
			printer:        nil,
			expectedOutput: "{\n  \"acknowledged\": true\n}\n",
			err:            nil,
		},
		{
			name:           "SearchDeleteEntity correctly prints an object with JSON ugly printer",
			client:         new(WorkingSearchDeleter),
			printer:        JsonObjectPrinter(false),
			expectedOutput: "{\"acknowledged\":true}\n",
			err:            nil,
		},

		// Error case
		{
			name:           "Delete returns 400 Bad Request",
			client:         new(FailingSearchDeleterDeleteFails),
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
			client:         new(FailingSearchDeleterParsingUUIDFails),
			printer:        nil,
			expectedOutput: "",
			err:            NewErrorWithCause(ErrorExitCode, errors.New("invalid UUID length: 7"), "Could not delete entity with name \"MongoDB Atlas Server\""),
		},
		{
			name:           "Search returns 404 Not Found",
			client:         new(FailingSearchDeleterSearchFails),
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
			client:    new(WorkingSearchDeleter),
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
