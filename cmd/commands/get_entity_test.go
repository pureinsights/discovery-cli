package commands

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/google/uuid"
	discoveryPackage "github.com/pureinsights/pdp-cli/discovery"
	"github.com/pureinsights/pdp-cli/internal/cli"
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

// TestGetCommand tests the GetCommand() function.
func TestGetCommand(t *testing.T) {
	tests := []struct {
		name           string
		client         cli.Getter
		args           []string
		url            string
		apiKey         string
		componentName  string
		expectedOutput string
		outWriter      io.Writer
		err            error
	}{
		// Working case
		{
			name:           "GetEntity correctly prints an object",
			url:            "http://localhost:12010/v2",
			apiKey:         "core123",
			componentName:  "Core",
			args:           []string{"3d51beef-8b90-40aa-84b5-033241dc6239"},
			client:         new(WorkingGetter),
			expectedOutput: "{\"active\":true,\"creationTimestamp\":\"2025-08-14T18:02:38Z\",\"id\":\"5f125024-1e5e-4591-9fee-365dc20eeeed\",\"labels\":[],\"lastUpdatedTimestamp\":\"2025-08-18T20:55:43Z\",\"name\":\"MongoDB text processor\",\"type\":\"mongo\"}\n",
			err:            nil,
		},
		{
			name:           "GetEntities correctly prints an array of objects",
			client:         new(WorkingGetter),
			url:            "http://localhost:12010/v2",
			apiKey:         "core123",
			componentName:  "Core",
			args:           []string{},
			expectedOutput: "{\"active\":true,\"creationTimestamp\":\"2025-08-21T17:57:16Z\",\"id\":\"3393f6d9-94c1-4b70-ba02-5f582727d998\",\"labels\":[],\"lastUpdatedTimestamp\":\"2025-08-21T17:57:16Z\",\"name\":\"MongoDB text processor 4\",\"type\":\"mongo\"}\n{\"active\":true,\"creationTimestamp\":\"2025-08-14T18:02:38Z\",\"id\":\"5f125024-1e5e-4591-9fee-365dc20eeeed\",\"labels\":[],\"lastUpdatedTimestamp\":\"2025-08-18T20:55:43Z\",\"name\":\"MongoDB text processor\",\"type\":\"mongo\"}\n{\"active\":true,\"creationTimestamp\":\"2025-08-14T18:02:38Z\",\"id\":\"86e7f920-a4e4-4b64-be84-5437a7673db8\",\"labels\":[],\"lastUpdatedTimestamp\":\"2025-08-14T18:02:38Z\",\"name\":\"Script processor\",\"type\":\"script\"}\n",
			err:            nil,
		},

		// Error case
		{
			name:          "CheckCredentials fails",
			client:        new(WorkingGetter),
			url:           "",
			apiKey:        "core123",
			componentName: "Core",
			args:          []string{},
			outWriter:     testutils.ErrWriter{Err: errors.New("write failed")},
			err:           cli.NewError(cli.ErrorExitCode, "The Discovery Core URL is missing for profile \"default\".\nTo set the URL for the Discovery Core API, run any of the following commands:\n      discovery config  --profile {profile}\n      discovery core config --profile {profile}"),
		},
		{
			name:           "id is not a UUID",
			args:           []string{"test"},
			client:         new(WorkingGetter),
			url:            "coreUrl",
			apiKey:         "core123",
			componentName:  "Core",
			expectedOutput: "",
			err:            cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("invalid UUID length: 4"), "Could not convert given id \"test\" to UUID. This command does not support filters or referencing an entity by name."),
		},
		{
			name:           "GetEntity returns 404 Not Found",
			client:         new(FailingGetter),
			url:            "http://localhost:12010/v2",
			apiKey:         "core123",
			componentName:  "Core",
			args:           []string{"3d51beef-8b90-40aa-84b5-033241dc6239"},
			expectedOutput: "",
			err: cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{
				Status: http.StatusNotFound,
				Body: gjson.Parse(`{
			"status": 404,
			"code": 1003,
			"messages": [
				"Secret not found: 5f125024-1e5e-4591-9fee-365dc20eeeed"
			],
			"timestamp": "2025-10-16T00:15:31.888410500Z"
		}`),
			}, "Could not get entity with id \"3d51beef-8b90-40aa-84b5-033241dc6239\""),
		},
		{
			name:           "GetAll returns 401 Unauthorized",
			client:         new(FailingGetter),
			url:            "http://localhost:12010/v2",
			apiKey:         "core123",
			componentName:  "Core",
			args:           []string{},
			expectedOutput: "",
			err:            cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{Status: http.StatusUnauthorized, Body: gjson.Parse(`{"error":"unauthorized"}`)}, "Could not get all entities"),
		},
		{
			name:          "Printing JSON fails",
			client:        new(WorkingGetter),
			url:           "http://localhost:12010/v2",
			apiKey:        "core123",
			componentName: "Core",
			args:          []string{"5f125024-1e5e-4591-9fee-365dc20eeeed"},
			outWriter:     testutils.ErrWriter{Err: errors.New("write failed")},
			err:           cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("write failed"), "Could not print JSON object"),
		},
		{
			name:          "Printing Array fails",
			client:        new(WorkingGetter),
			url:           "http://localhost:12010/v2",
			apiKey:        "core123",
			componentName: "Core",
			args:          []string{},
			outWriter:     testutils.ErrWriter{Err: errors.New("write failed")},
			err:           cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("write failed"), "Could not print JSON Array"),
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

			vpr := viper.New()
			vpr.Set("profile", "default")
			vpr.Set("output", "json")
			if tc.url != "" {
				vpr.Set("default.core_url", tc.url)
			}
			if tc.apiKey != "" {
				vpr.Set("default.core_key", tc.apiKey)
			}

			d := cli.NewDiscovery(&ios, vpr, "")
			err := GetCommand(tc.args, d, tc.client, GetCommandConfig("default", "json", tc.componentName, "core_url", "core_key"))

			if tc.err != nil {
				require.Error(t, err)
				var errStruct cli.Error
				require.ErrorAs(t, err, &errStruct)
				assert.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedOutput, buf.String())
			}
		})
	}
}

// TestSearchCommand tests the SearchCommand() function.
func TestSearchCommand(t *testing.T) {
	tests := []struct {
		name           string
		client         searcher
		args           []string
		filters        []string
		url            string
		apiKey         string
		componentName  string
		expectedOutput string
		outWriter      io.Writer
		err            error
	}{
		// Working case
		{
			name:           "Search by name returns an object",
			args:           []string{"label test clone 10"},
			url:            "http://localhost:12010/v2",
			apiKey:         "apiKey123",
			componentName:  "Core",
			client:         new(WorkingSearcher),
			expectedOutput: "{\"highlight\":{\"name\":[\"\\u003cem\\u003eMongoDB\\u003c/em\\u003e \\u003cem\\u003eAtlas\\u003c/em\\u003e \\u003cem\\u003eserver\\u003c/em\\u003e clone\"]},\"score\":0.50769836,\"source\":{\"active\":true,\"creationTimestamp\":\"2025-09-29T15:50:17Z\",\"id\":\"986ce864-af76-4fcb-8b4f-f4e4c6ab0951\",\"labels\":[],\"lastUpdatedTimestamp\":\"2025-09-29T15:50:17Z\",\"name\":\"MongoDB Atlas server clone\",\"type\":\"mongo\"}}\n",
			err:            nil,
		},
		{
			name:           "Search by Id returns an object",
			args:           []string{"986ce864-af76-4fcb-8b4f-f4e4c6ab0951"},
			url:            "http://localhost:12010/v2",
			apiKey:         "apiKey123",
			componentName:  "Core",
			client:         new(FailingSearcherWorkingGetter),
			expectedOutput: "{\"active\":true,\"config\":{\"connection\":{\"connectTimeout\":\"1m\",\"readTimeout\":\"30s\"},\"credentialId\":\"9ababe08-0b74-4672-bb7c-e7a8227d6d4c\",\"servers\":[\"mongodb+srv://cluster0.dleud.mongodb.net/\"]},\"creationTimestamp\":\"2025-09-29T15:50:17Z\",\"id\":\"986ce864-af76-4fcb-8b4f-f4e4c6ab0951\",\"labels\":[],\"lastUpdatedTimestamp\":\"2025-09-29T15:50:17Z\",\"name\":\"MongoDB Atlas server clone\",\"type\":\"mongo\"}\n",
			err:            nil,
		},
		{
			name:           "Get with no args returns an array",
			args:           []string{},
			url:            "http://localhost:12010/v2",
			apiKey:         "apiKey123",
			componentName:  "Core",
			client:         new(WorkingSearcher),
			expectedOutput: "{\"active\":true,\"creationTimestamp\":\"2025-10-17T22:37:57Z\",\"id\":\"3b32e410-2f33-412d-9fb8-17970131921c\",\"labels\":[{\"key\":\"A\",\"value\":\"A\"}],\"lastUpdatedTimestamp\":\"2025-10-17T22:37:57Z\",\"name\":\"label test 1 clone 10\",\"type\":\"mongo\"}\n{\"active\":true,\"creationTimestamp\":\"2025-10-17T22:38:12Z\",\"id\":\"5c09589e-b643-41aa-a766-3b7fc3660473\",\"labels\":[],\"lastUpdatedTimestamp\":\"2025-10-17T22:38:12Z\",\"name\":\"OpenAI credential clone clone\",\"type\":\"openai\"}\n",
			err:            nil,
		},
		{
			name:           "Get with args returns a search array",
			args:           []string{},
			filters:        []string{"type=mongo"},
			url:            "http://localhost:12010/v2",
			apiKey:         "apiKey123",
			componentName:  "Core",
			client:         new(WorkingSearcher),
			expectedOutput: "{\"highlight\":{},\"score\":0.20970252,\"source\":{\"active\":true,\"creationTimestamp\":\"2025-09-29T15:50:17Z\",\"id\":\"986ce864-af76-4fcb-8b4f-f4e4c6ab0951\",\"labels\":[],\"lastUpdatedTimestamp\":\"2025-09-29T15:50:17Z\",\"name\":\"MongoDB Atlas server clone\",\"type\":\"mongo\"}}\n{\"highlight\":{},\"score\":0.20970252,\"source\":{\"active\":true,\"creationTimestamp\":\"2025-09-29T15:50:19Z\",\"id\":\"8f14c11c-bb66-49d3-aa2a-dedff4608c17\",\"labels\":[],\"lastUpdatedTimestamp\":\"2025-09-29T15:50:19Z\",\"name\":\"MongoDB Atlas server clone 1\",\"type\":\"mongo\"}}\n{\"highlight\":{},\"score\":0.20970252,\"source\":{\"active\":true,\"creationTimestamp\":\"2025-09-29T15:50:20Z\",\"id\":\"3a0214a4-72cc-4eee-ad0c-9e3af9b08a6c\",\"labels\":[],\"lastUpdatedTimestamp\":\"2025-09-29T15:50:20Z\",\"name\":\"MongoDB Atlas server clone 3\",\"type\":\"mongo\"}}\n",
			err:            nil,
		},

		// Error case
		{
			name:          "CheckCredentials fails",
			client:        new(WorkingSearcher),
			apiKey:        "core123",
			componentName: "Core",
			args:          []string{},
			outWriter:     testutils.ErrWriter{Err: errors.New("write failed")},
			err:           NewError(ErrorExitCode, "The Discovery Core URL is missing for profile \"default\".\nTo set the URL for the Discovery Core API, run any of the following commands:\n      discovery config  --profile {profile}\n      discovery core config --profile {profile}"),
		},
		{
			name:           "user sends a name that does not exist",
			args:           []string{"test"},
			url:            "http://localhost:12010/v2",
			apiKey:         "apiKey123",
			client:         new(FailingSearcherFailingGetter),
			expectedOutput: "",
			err: NewErrorWithCause(ErrorExitCode, discoveryPackage.Error{
				Status: http.StatusNotFound,
				Body: gjson.Parse(`{
	"status": 404,
	"code": 1003,
	"messages": [
		"Entity not found: entity with name "test" does not exist"
	]
}`),
			}, "Could not search for entity with id \"test\""),
		},

		{
			name:   "Search By Name returns HTTP error",
			args:   []string{"3b32e410-2F33-412d-9fb8-17970131921c"},
			url:    "http://localhost:12010/v2",
			apiKey: "apiKey123",
			client: new(FailingSearcher),
			err: NewErrorWithCause(ErrorExitCode, discoveryPackage.Error{
				Status: http.StatusBadRequest,
				Body: gjson.Parse(`{
	"status": 400,
	"code": 3002,
	"messages": [
		"Invalid JSON: Unexpected end-of-input:"
	],
	"timestamp": "2025-10-17T17:43:52.817308100Z"
	}`)}, "Could not search for entity with id \"3b32e410-2F33-412d-9fb8-17970131921c\""),
		},
		{
			name:   "Get with no args returns HTTP error",
			args:   []string{},
			url:    "http://localhost:12010/v2",
			apiKey: "apiKey123",
			client: new(FailingSearcher),
			err:    NewErrorWithCause(ErrorExitCode, discoveryPackage.Error{Status: http.StatusUnauthorized, Body: gjson.Parse(`{"error":"unauthorized"}`)}, "Could not get all entities"),
		},
		{
			name:    "Get with filters returns HTTP error",
			args:    []string{},
			filters: []string{"label=A"},
			url:     "http://localhost:12010/v2",
			apiKey:  "apiKey123",
			client:  new(FailingSearcher),
			err: NewErrorWithCause(ErrorExitCode, discoveryPackage.Error{
				Status: http.StatusNotFound,
				Body:   gjson.Result{},
			}, "Could not search for the entities"),
		},
		{
			name:    "Filter does not exist",
			args:    []string{},
			filters: []string{"gte=field:1"},
			url:     "http://localhost:12010/v2",
			apiKey:  "apiKey123",
			client:  new(WorkingSearcher),
			err:     NewError(ErrorExitCode, "Filter type \"gte\" does not exist"),
		},
		{
			name:          "Printing JSON fails",
			client:        new(WorkingSearcher),
			url:           "http://localhost:12010/v2",
			apiKey:        "core123",
			componentName: "Core",
			args:          []string{"5f125024-1e5e-4591-9fee-365dc20eeeed"},
			outWriter:     testutils.ErrWriter{Err: errors.New("write failed")},
			err:           NewErrorWithCause(ErrorExitCode, errors.New("write failed"), "Could not print JSON object"),
		},
		{
			name:          "Printing Array fails",
			client:        new(WorkingSearcher),
			url:           "http://localhost:12010/v2",
			apiKey:        "core123",
			componentName: "Core",
			args:          []string{},
			outWriter:     testutils.ErrWriter{Err: errors.New("write failed")},
			err:           NewErrorWithCause(ErrorExitCode, errors.New("write failed"), "Could not print JSON Array"),
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

			vpr := viper.New()
			vpr.Set("profile", "default")
			vpr.Set("output", "json")
			if tc.url != "" {
				vpr.Set("default.core_url", tc.url)
			}
			if tc.apiKey != "" {
				vpr.Set("default.core_key", tc.apiKey)
			}

			d := NewDiscovery(&ios, vpr, "")
			err := SearchCommand(tc.args, d, tc.client, GetCommandConfig("default", "json", tc.componentName, "core_url", "core_key"), &tc.filters)

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
