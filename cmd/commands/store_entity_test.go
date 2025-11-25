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

// TestStoreCommandConfig tests the StoreCommandConfig() function.
func TestStoreCommandConfig(t *testing.T) {
	base := commandConfig{
		profile:       "cn",
		output:        "json",
		url:           "https://localhost:12010",
		apiKey:        "core123",
		componentName: "Core",
	}

	data := "[{\"active\":true,\"creationTimestamp\":\"2025-08-14T18:02:11Z\",\"labels\":[],\"lastUpdatedTimestamp\":\"2025-08-14T18:02:11Z\",\"name\":\"MongoDB credential\",\"secret\":\"mongo-secret\",\"type\":\"mongo\"},{\"active\":true,\"creationTimestamp\":\"2025-08-14T18:02:11Z\",\"id\":\"9ababe08-0b74-4672-bb7c-e7a8227d6d4c\",\"labels\":[],\"lastUpdatedTimestamp\":\"2025-08-14T18:02:11Z\",\"name\":\"MongoDB credential\",\"secret\":\"mongo-secret\",\"type\":\"mongo\"},{\"active\":true,\"creationTimestamp\":\"2025-08-14T18:02:11Z\",\"id\":\"9ababe08-0b74-4672-bb7c-e7a8227d6d4c\",\"labels\":[],\"lastUpdatedTimestamp\":\"2025-08-14T18:02:11Z\",\"name\":\"MongoDB credential\",\"secret\":\"mongo-secret\",\"type\":\"mongo\"}]"
	file := "config.json"
	got := StoreCommandConfig(base, true, data, file)

	require.Equal(t, base, got.commandConfig)

	assert.Equal(t, "cn", got.profile)
	assert.Equal(t, "json", got.output)
	assert.Equal(t, "https://localhost:12010", got.url)
	assert.Equal(t, "core123", got.apiKey)
	assert.Equal(t, "Core", got.componentName)

	assert.True(t, got.abortOnError)
	assert.Equal(t, data, got.data)
	assert.Equal(t, file, got.file)
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

// TestStoreCommand tests the StoreCommand() function.
func TestStoreCommand(t *testing.T) {
	tests := []struct {
		name           string
		client         cli.Creator
		url            string
		apiKey         string
		componentName  string
		abortOnError   bool
		data           string
		file           string
		expectedOutput string
		outWriter      io.Writer
		err            error
	}{
		// Working case
		{
			name:           "UpsertEntities correctly prints the array",
			url:            "http://localhost:12010/v2",
			apiKey:         "",
			componentName:  "Core",
			client:         new(WorkingCreator),
			abortOnError:   false,
			data:           "[{\"active\":true,\"creationTimestamp\":\"2025-08-14T18:02:11Z\",\"labels\":[],\"lastUpdatedTimestamp\":\"2025-08-14T18:02:11Z\",\"name\":\"MongoDB credential\",\"secret\":\"mongo-secret\",\"type\":\"mongo\"},{\"active\":true,\"creationTimestamp\":\"2025-08-14T18:02:11Z\",\"id\":\"9ababe08-0b74-4672-bb7c-e7a8227d6d4c\",\"labels\":[],\"lastUpdatedTimestamp\":\"2025-08-14T18:02:11Z\",\"name\":\"MongoDB credential\",\"secret\":\"mongo-secret\",\"type\":\"mongo\"},{\"active\":true,\"creationTimestamp\":\"2025-08-14T18:02:11Z\",\"id\":\"9ababe08-0b74-4672-bb7c-e7a8227d6d4c\",\"labels\":[],\"lastUpdatedTimestamp\":\"2025-08-14T18:02:11Z\",\"name\":\"MongoDB credential\",\"secret\":\"mongo-secret\",\"type\":\"mongo\"}]",
			file:           "",
			expectedOutput: "{\"active\":true,\"creationTimestamp\":\"2025-08-14T18:02:11Z\",\"id\":\"9ababe08-0b74-4672-bb7c-e7a8227d6d4c\",\"labels\":[],\"lastUpdatedTimestamp\":\"2025-08-14T18:02:11Z\",\"name\":\"MongoDB credential\",\"secret\":\"mongo-secret\",\"type\":\"mongo\"}\n{\"active\":true,\"creationTimestamp\":\"2025-08-14T18:02:11Z\",\"id\":\"9ababe08-0b74-4672-bb7c-e7a8227d6d4c\",\"labels\":[],\"lastUpdatedTimestamp\":\"2025-08-14T18:02:11Z\",\"name\":\"MongoDB credential\",\"secret\":\"mongo-secret\",\"type\":\"mongo\"}\n{\"active\":true,\"creationTimestamp\":\"2025-08-14T18:02:11Z\",\"id\":\"9ababe08-0b74-4672-bb7c-e7a8227d6d4c\",\"labels\":[],\"lastUpdatedTimestamp\":\"2025-08-14T18:02:11Z\",\"name\":\"MongoDB credential\",\"secret\":\"mongo-secret\",\"type\":\"mongo\"}\n",
			err:            nil,
		},
		{
			name:           "UpsertEntities correctly reads the file and prints the array",
			url:            "http://localhost:12010/v2",
			apiKey:         "core123",
			componentName:  "Core",
			client:         new(WorkingCreator),
			abortOnError:   false,
			data:           "",
			file:           "testdata/StoreCommand_JSONFile.json",
			expectedOutput: "{\"active\":true,\"creationTimestamp\":\"2025-08-14T18:02:11Z\",\"id\":\"9ababe08-0b74-4672-bb7c-e7a8227d6d4c\",\"labels\":[],\"lastUpdatedTimestamp\":\"2025-08-14T18:02:11Z\",\"name\":\"MongoDB credential\",\"secret\":\"mongo-secret\",\"type\":\"mongo\"}\n{\"active\":true,\"creationTimestamp\":\"2025-08-14T18:02:11Z\",\"id\":\"9ababe08-0b74-4672-bb7c-e7a8227d6d4c\",\"labels\":[],\"lastUpdatedTimestamp\":\"2025-08-14T18:02:11Z\",\"name\":\"MongoDB credential\",\"secret\":\"mongo-secret\",\"type\":\"mongo\"}\n{\"active\":true,\"creationTimestamp\":\"2025-08-14T18:02:11Z\",\"id\":\"9ababe08-0b74-4672-bb7c-e7a8227d6d4c\",\"labels\":[],\"lastUpdatedTimestamp\":\"2025-08-14T18:02:11Z\",\"name\":\"MongoDB credential\",\"secret\":\"mongo-secret\",\"type\":\"mongo\"}\n",
			err:            nil,
		},

		// Error case
		{
			name:          "CheckCredentials fails",
			client:        new(WorkingCreator),
			url:           "",
			apiKey:        "core123",
			componentName: "Core",
			outWriter:     testutils.ErrWriter{Err: errors.New("write failed")},
			err:           cli.NewError(cli.ErrorExitCode, "The Discovery Core URL is missing for profile \"default\".\nTo set the URL for the Discovery Core API, run any of the following commands:\n      discovery config  --profile \"default\"\n      discovery core config --profile \"default\""),
		},
		{
			name:           "UpsertEntities returns 404 Not Found",
			client:         new(FailingCreator),
			url:            "http://localhost:12010/v2",
			apiKey:         "core123",
			componentName:  "Core",
			abortOnError:   true,
			data:           "[{\"active\":true,\"creationTimestamp\":\"2025-08-14T18:02:11Z\",\"id\":\"9ababe08-0b74-4672-bb7c-e7a8227d6d4c\",\"labels\":[],\"lastUpdatedTimestamp\":\"2025-08-14T18:02:11Z\",\"name\":\"MongoDB credential\",\"secret\":\"mongo-secret\",\"type\":\"mongo\"}]",
			file:           "",
			expectedOutput: "",
			err: cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{Status: http.StatusBadRequest, Body: gjson.Parse(`{
  "status": 404,
  "code": 1003,
  "messages": [
    "Entity not found: 9ababe08-0b74-4672-bb7c-e7a8227d6d4d"
  ],
  "timestamp": "2025-10-29T14:47:36.290329Z"
}`)}, "Could not store entities"),
		},
		{
			name:           "StoreCommand reads an empty file",
			url:            "http://localhost:12010/v2",
			apiKey:         "core123",
			componentName:  "Core",
			client:         new(WorkingCreator),
			abortOnError:   false,
			data:           "",
			file:           "testdata/StoreCommand_EmptyFile.json",
			expectedOutput: "",
			err:            cli.NewError(cli.ErrorExitCode, "Data cannot be empty"),
		},
		{
			name:           "StoreCommand receives empty data",
			url:            "http://localhost:12010/v2",
			apiKey:         "core123",
			componentName:  "Core",
			client:         new(WorkingCreator),
			abortOnError:   false,
			data:           "",
			file:           "",
			expectedOutput: "",
			err:            cli.NewError(cli.ErrorExitCode, "Data cannot be empty"),
		},
		{
			name:           "StoreCommand tries to read a file that does not exist",
			url:            "http://localhost:12010/v2",
			apiKey:         "core123",
			componentName:  "Core",
			client:         new(WorkingCreator),
			abortOnError:   false,
			data:           "",
			file:           "doesnotexist",
			expectedOutput: "",
			err:            cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("file does not exist: doesnotexist"), "Could not read file \"doesnotexist\""),
		},
		{
			name:          "Printing Array fails",
			client:        new(WorkingCreator),
			url:           "http://localhost:12010/v2",
			apiKey:        "core123",
			componentName: "Core",
			abortOnError:  false,
			data:          "[{\"active\":true,\"creationTimestamp\":\"2025-08-14T18:02:11Z\",\"labels\":[],\"lastUpdatedTimestamp\":\"2025-08-14T18:02:11Z\",\"name\":\"MongoDB credential\",\"secret\":\"mongo-secret\",\"type\":\"mongo\"},{\"active\":true,\"creationTimestamp\":\"2025-08-14T18:02:11Z\",\"id\":\"9ababe08-0b74-4672-bb7c-e7a8227d6d4c\",\"labels\":[],\"lastUpdatedTimestamp\":\"2025-08-14T18:02:11Z\",\"name\":\"MongoDB credential\",\"secret\":\"mongo-secret\",\"type\":\"mongo\"},{\"active\":true,\"creationTimestamp\":\"2025-08-14T18:02:11Z\",\"id\":\"9ababe08-0b74-4672-bb7c-e7a8227d6d4c\",\"labels\":[],\"lastUpdatedTimestamp\":\"2025-08-14T18:02:11Z\",\"name\":\"MongoDB credential\",\"secret\":\"mongo-secret\",\"type\":\"mongo\"}]",
			file:          "",
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
			err := StoreCommand(d, tc.client, StoreCommandConfig(GetCommandConfig("default", "json", tc.componentName, "core_url"), tc.abortOnError, tc.data, tc.file))

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
