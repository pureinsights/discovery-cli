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

type WorkingGetter struct {
	mock.Mock
}

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

type FailingGetter struct {
	mock.Mock
}

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

func (g *FailingGetter) GetAll() ([]gjson.Result, error) {
	return []gjson.Result(nil), discoveryPackage.Error{Status: http.StatusUnauthorized, Body: gjson.Parse(`{"error":"unauthorized"}`)}
}

// Test_discovery_GetEntity tests the discovery.GetEntity() function.
func Test_discovery_GetEntity(t *testing.T) {
	tests := []struct {
		name           string
		client         getter
		printer        Printer
		expectedOutput string
		outWriter      io.Writer
		err            error
	}{
		// Working case
		{
			name:           "GetEntity correctly prints an object with the sent printer",
			client:         new(WorkingGetter),
			printer:        JsonObjectPrinter(true),
			expectedOutput: "{\n  \"active\": true,\n  \"creationTimestamp\": \"2025-08-14T18:02:38Z\",\n  \"id\": \"5f125024-1e5e-4591-9fee-365dc20eeeed\",\n  \"labels\": [],\n  \"lastUpdatedTimestamp\": \"2025-08-18T20:55:43Z\",\n  \"name\": \"MongoDB text processor\",\n  \"type\": \"mongo\"\n}\n",
			err:            nil,
		},
		{
			name:           "GetEntity correctly prints an object with JSON ugly printer",
			client:         new(WorkingGetter),
			printer:        nil,
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
		client         getter
		printer        Printer
		expectedOutput string
		outWriter      io.Writer
		err            error
	}{
		// Working case
		{
			name:           "GetEntities correctly prints an array with the sent printer",
			client:         new(WorkingGetter),
			printer:        JsonArrayPrinter(true),
			expectedOutput: "[\n{\n  \"active\": true,\n  \"creationTimestamp\": \"2025-08-21T17:57:16Z\",\n  \"id\": \"3393f6d9-94c1-4b70-ba02-5f582727d998\",\n  \"labels\": [],\n  \"lastUpdatedTimestamp\": \"2025-08-21T17:57:16Z\",\n  \"name\": \"MongoDB text processor 4\",\n  \"type\": \"mongo\"\n}\n{\n  \"active\": true,\n  \"creationTimestamp\": \"2025-08-14T18:02:38Z\",\n  \"id\": \"5f125024-1e5e-4591-9fee-365dc20eeeed\",\n  \"labels\": [],\n  \"lastUpdatedTimestamp\": \"2025-08-18T20:55:43Z\",\n  \"name\": \"MongoDB text processor\",\n  \"type\": \"mongo\"\n}\n{\n  \"active\": true,\n  \"creationTimestamp\": \"2025-08-14T18:02:38Z\",\n  \"id\": \"86e7f920-a4e4-4b64-be84-5437a7673db8\",\n  \"labels\": [],\n  \"lastUpdatedTimestamp\": \"2025-08-14T18:02:38Z\",\n  \"name\": \"Script processor\",\n  \"type\": \"script\"\n}\n]\n",
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
			err:       NewErrorWithCause(ErrorExitCode, errors.New("write failed"), "Could not print JSON array"),
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

type WorkingSearcher struct {
	mock.Mock
}

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

func (g *WorkingSearcher) SearchByName() (gjson.Result, error) {
	return gjson.Parse(` {
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
	}`), nil
}

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
			err:            NewError(ErrorExitCode, "Filter \"name\" does not exist"),
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
