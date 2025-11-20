package commands

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

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

// WorkingBackupRestore mocks a working backup restore
type WorkingBackupRestore struct {
	mock.Mock
}

// Export returns zip bytes as if the request worked successfully.
func (g *WorkingBackupRestore) Export() ([]byte, string, error) {
	return []byte("Exportfiles"), "export-20251110T1455.zip", nil
}

// Import implements the interface
func (g *WorkingBackupRestore) Import(discoveryPackage.OnConflict, string) (gjson.Result, error) {
	return gjson.Result{}, nil
}

// FailingBackupRestore mocks a failing backup restore
type FailingBackupRestore struct {
	mock.Mock
}

// Get returns an error as if the request failed.
func (g *FailingBackupRestore) Export() ([]byte, string, error) {
	return []byte(nil), "discovery.zip", discoveryPackage.Error{Status: http.StatusUnauthorized, Body: gjson.Parse(`{"error":"unauthorized"}`)}
}

// Import implements the interface
func (g *FailingBackupRestore) Import(discoveryPackage.OnConflict, string) (gjson.Result, error) {
	return gjson.Result{}, nil
}

// TestExportCommand tests the ExportCommand() function.
func TestExportCommand(t *testing.T) {
	tests := []struct {
		name           string
		client         cli.BackupRestore
		file           string
		url            string
		apiKey         string
		componentName  string
		expectedOutput string
		outWriter      io.Writer
		err            error
	}{
		// Working case
		{
			name:           "Export correctly prints an object",
			url:            "http://localhost:12010",
			apiKey:         "core123",
			componentName:  "Core",
			file:           filepath.Join(t.TempDir(), "export.zip"),
			client:         new(WorkingBackupRestore),
			expectedOutput: "{\"acknowledged\":true}\n",
			err:            nil,
		},

		// Error case
		{
			name:          "CheckCredentials fails",
			client:        new(WorkingBackupRestore),
			url:           "",
			apiKey:        "core123",
			componentName: "Core",
			file:          "",
			outWriter:     testutils.ErrWriter{Err: errors.New("write failed")},
			err:           cli.NewError(cli.ErrorExitCode, "The Discovery Core URL is missing for profile \"default\".\nTo set the URL for the Discovery Core API, run any of the following commands:\n      discovery config  --profile \"default\"\n      discovery core config --profile \"default\""),
		},
		{
			name:           "Export fails",
			client:         new(FailingBackupRestore),
			url:            "http://localhost:12010",
			apiKey:         "core123",
			componentName:  "Core",
			file:           "",
			expectedOutput: "",
			err:            cli.NewErrorWithCause(cli.ErrorExitCode, discoveryPackage.Error{Status: http.StatusUnauthorized, Body: gjson.Parse(`{"error":"unauthorized"}`)}, "Could not export entities"),
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
			err := ExportCommand(d, tc.client, tc.file, GetCommandConfig("default", "json", tc.componentName, "core_url"))

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
