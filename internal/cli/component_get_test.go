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
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// TestGetCommand tests the GetCommand() function.
func TestGetCommand(t *testing.T) {
	tests := []struct {
		name           string
		client         getter
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
			err:           NewError(ErrorExitCode, "The Discovery Core URL is missing for profile \"default\".\nTo set the URL for the Discovery Core API, run any of the following commands:\n      discovery config  --profile {profile}\n      discovery core config --profile {profile}"),
		},
		{
			name:           "id is not a UUID",
			args:           []string{"test"},
			client:         new(WorkingGetter),
			url:            "coreUrl",
			apiKey:         "core123",
			componentName:  "Core",
			expectedOutput: "",
			err:            NewErrorWithCause(ErrorExitCode, errors.New("invalid UUID length: 4"), "Could not convert given id \"test\" to UUID. This command does not support filters or referencing an entity by name."),
		},
		{
			name:           "GetEntity returns 404 Not Found",
			client:         new(FailingGetter),
			url:            "http://localhost:12010/v2",
			apiKey:         "core123",
			componentName:  "Core",
			args:           []string{"3d51beef-8b90-40aa-84b5-033241dc6239"},
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
			err:            NewErrorWithCause(ErrorExitCode, discoveryPackage.Error{Status: http.StatusUnauthorized, Body: gjson.Parse(`{"error":"unauthorized"}`)}, "Could not get all entities"),
		},
		{
			name:          "Printing JSON fails",
			client:        new(WorkingGetter),
			url:           "http://localhost:12010/v2",
			apiKey:        "core123",
			componentName: "Core",
			args:          []string{"5f125024-1e5e-4591-9fee-365dc20eeeed"},
			outWriter:     testutils.ErrWriter{Err: errors.New("write failed")},
			err:           NewErrorWithCause(ErrorExitCode, errors.New("write failed"), "Could not print JSON object"),
		},
		{
			name:          "Printing Array fails",
			client:        new(WorkingGetter),
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
			err := GetCommand(tc.args, d, tc.client, GetCommandConfig("default", "json", tc.componentName, "core_url", "core_key"))

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
