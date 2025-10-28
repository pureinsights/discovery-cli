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

// TestDeleteCommand tests the DeleteCommand() function.
func TestDeleteCommand(t *testing.T) {
	tests := []struct {
		name           string
		client         cli.Deleter
		args           string
		url            string
		apiKey         string
		componentName  string
		expectedOutput string
		outWriter      io.Writer
		err            error
	}{
		// Working case
		{
			name:           "DeleteEntity correctly prints an object",
			url:            "http://localhost:12010/v2",
			apiKey:         "core123",
			componentName:  "Core",
			args:           "5f125024-1e5e-4591-9fee-365dc20eeeed",
			client:         new(WorkingDeleter),
			expectedOutput: "{\"acknowledged\":true}\n",
			err:            nil,
		},

		// Error case
		{
			name:          "CheckCredentials fails",
			client:        new(WorkingDeleter),
			url:           "",
			apiKey:        "core123",
			componentName: "Core",
			args:          "",
			outWriter:     testutils.ErrWriter{Err: errors.New("write failed")},
			err:           cli.NewError(cli.ErrorExitCode, "The Discovery Core URL is missing for profile \"default\".\nTo set the URL for the Discovery Core API, run any of the following commands:\n      discovery config  --profile {profile}\n      discovery core config --profile {profile}"),
		},
		{
			name:           "id is not a UUID",
			args:           "test",
			client:         new(WorkingDeleter),
			url:            "coreUrl",
			apiKey:         "core123",
			componentName:  "Core",
			expectedOutput: "",
			err:            cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("invalid UUID length: 4"), "Could not convert given id \"test\" to UUID. This command does not support referencing an entity by name."),
		},
		{
			name:           "DeleteEntity returns 400 Bad Request",
			client:         new(FailingDeleter),
			url:            "http://localhost:12010/v2",
			apiKey:         "core123",
			componentName:  "Core",
			args:           "5f125024-1e5e-4591-9fee-365dc20eeeed",
			expectedOutput: "",
			err: cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{
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
			name:          "Printing JSON fails",
			client:        new(WorkingDeleter),
			url:           "http://localhost:12010/v2",
			apiKey:        "core123",
			componentName: "Core",
			args:          "5f125024-1e5e-4591-9fee-365dc20eeeed",
			outWriter:     testutils.ErrWriter{Err: errors.New("write failed")},
			err:           cli.NewErrorWithCause(cli.ErrorExitCode, errors.New("write failed"), "Could not print JSON object"),
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
			err := DeleteCommand(tc.args, d, tc.client, GetCommandConfig("default", "json", tc.componentName, "core_url", "core_key"))

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

// TestSearchDeleteCommand tests the SearchDeleteCommand() function
func TestSearchDeleteCommand(t *testing.T) {
	tests := []struct {
		name           string
		client         cli.SearchDeleter
		args           string
		url            string
		apiKey         string
		componentName  string
		expectedOutput string
		outWriter      io.Writer
		err            error
	}{
		// Working case
		{
			name:           "SearchDeleteEntity correctly prints an object",
			url:            "http://localhost:12010/v2",
			apiKey:         "core123",
			componentName:  "Core",
			args:           "MongoDB Atlas Server",
			client:         new(WorkingSearchDeleter),
			expectedOutput: "{\"acknowledged\":true}\n",
			err:            nil,
		},

		// Error case
		{
			name:          "CheckCredentials fails",
			client:        new(WorkingSearchDeleter),
			url:           "",
			apiKey:        "core123",
			componentName: "Core",
			args:          "",
			outWriter:     testutils.ErrWriter{Err: errors.New("write failed")},
			err:           NewError(ErrorExitCode, "The Discovery Core URL is missing for profile \"default\".\nTo set the URL for the Discovery Core API, run any of the following commands:\n      discovery config  --profile {profile}\n      discovery core config --profile {profile}"),
		},
		{
			name:           "SearchDeleteEntity returns 400 Bad Request",
			client:         new(FailingSearchDeleterSearchFails),
			url:            "http://localhost:12010/v2",
			apiKey:         "core123",
			componentName:  "Core",
			args:           "MongoDB Atlas Server",
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
			name:          "Printing JSON fails",
			client:        new(WorkingSearchDeleter),
			url:           "http://localhost:12010/v2",
			apiKey:        "core123",
			componentName: "Core",
			args:          "MongoDB Atlas Server",
			outWriter:     testutils.ErrWriter{Err: errors.New("write failed")},
			err:           NewErrorWithCause(ErrorExitCode, errors.New("write failed"), "Could not print JSON object"),
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
			err := SearchDeleteCommand(tc.args, d, tc.client, GetCommandConfig("default", "json", tc.componentName, "core_url", "core_key"))

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
