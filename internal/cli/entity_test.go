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
		client         Getter
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
